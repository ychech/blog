// package database 负责 MySQL 与 Redis 的连接初始化、自动迁移。
// 所有服务层都通过此包暴露的全局变量访问数据库。
package database

import (
	"blog/config"
	"blog/model"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// DB 全局 MySQL 连接（GORM）
var DB *gorm.DB

// Redis 全局 Redis 连接（go-redis）
var Redis *redis.Client

// Init 初始化数据库和缓存。
// MySQL 初始化失败会返回错误；Redis 失败仅记录日志，服务仍可启动（缓存降级）。
func Init() error {
	if err := initMySQL(); err != nil {
		return err
	}
	if err := initRedis(); err != nil {
		log.Printf("Redis 连接失败（非致命）: %v", err)
		// Redis 失败不影响主服务启动，只是缓存不生效
	}
	return nil
}

// initMySQL 初始化 MySQL 连接并执行数据库自动迁移。
func initMySQL() error {
	cfg := config.C.DB

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = db
	log.Println("MySQL 连接成功")

	// 配置读写分离：如果配置了从库，则读请求路由到从库
	if len(cfg.Replicas) > 0 {
		replicas := make([]gorm.Dialector, 0, len(cfg.Replicas))
		for _, r := range cfg.Replicas {
			dsn := replicaDSN(r, cfg)
			replicas = append(replicas, mysql.Open(dsn))
			log.Printf("MySQL 只读从库: %s:%s", r.Host, r.Port)
		}

		resolver := dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		})
		if err := db.Use(resolver); err != nil {
			return fmt.Errorf("配置读写分离失败: %w", err)
		}
		log.Println("MySQL 读写分离已启用")
	}

	return migrate()
}

// replicaDSN 根据从库配置和主库配置生成 DSN。
// 从库配置中未指定的字段（端口、用户名、密码、数据库名、字符集）继承自主库。
func replicaDSN(r config.DBReplicaConfig, master config.DBConfig) string {
	host := r.Host
	port := r.Port
	user := r.User
	password := r.Password

	if port == "" {
		port = master.Port
	}
	if user == "" {
		user = master.User
	}
	if password == "" {
		password = master.Password
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		user, password, host, port, master.Database, master.Charset,
	)
}

// initRedis 初始化 Redis 连接，使用 3 秒超时进行 Ping 探测。
func initRedis() error {
	cfg := config.C.Redis
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		Redis = nil
		return err
	}

	log.Println("Redis 连接成功")
	return nil
}

// migrate 使用 GORM AutoMigrate 自动创建/更新表结构。
// 新增模型时，只需在此追加即可。
func migrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.CommentLike{},
		&model.Badge{},
		&model.UserBadge{},
		&model.Notification{},
		&model.AuditLog{},
		&model.CommentReport{},
		&model.Message{},
		&model.Favorite{},
		&model.ReadHistory{},
		&model.OAuthAccount{},
		&model.UserFollow{},
	)
}

// Close 优雅关闭数据库连接。
// 在程序退出时调用，关闭 MySQL 与 Redis 连接，避免连接泄露。
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	if Redis != nil {
		if err := Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}
