// package service 实现文章浏览量统计与同步。
//
// 设计思路：
//   1. 用户每次访问文章，浏览量先写入 Redis（Incr），避免直接写 MySQL 造成压力。
//   2. 后台定时任务（默认 10 秒）将 Redis 中的增量同步到 MySQL，然后删除 Redis 中的增量。
//   3. 读取文章时，返回 MySQL 中的持久化值 + Redis 中的未同步增量，保证数据实时性。
//   4. Redis 不可用时，自动降级为直接更新 MySQL。
package service

import (
	"blog/database"
	"blog/model"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// ViewCountRedisKeyPrefix 是 Redis 中文章浏览量增量的 key 前缀。
	ViewCountRedisKeyPrefix = "blog:post:views:"

	// ViewCountSyncInterval 是浏览量同步到 MySQL 的间隔。
	ViewCountSyncInterval = 10 * time.Second
)

// viewCountKey 根据文章 ID 生成 Redis key。
func viewCountKey(postID uint) string {
	return fmt.Sprintf("%s%d", ViewCountRedisKeyPrefix, postID)
}

// IncrementPostViewCount 增加指定文章的浏览量。
// Redis 可用时写入 Redis；不可用时直接更新 MySQL。
func IncrementPostViewCount(postID uint) error {
	if database.Redis == nil {
		return database.DB.Model(&model.Post{}).Where("id = ?", postID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := database.Redis.Incr(ctx, viewCountKey(postID)).Err(); err != nil {
		return fmt.Errorf("redis incr view count failed: %w", err)
	}
	return nil
}

// GetPostViewCount 获取文章当前浏览量（MySQL 持久化值 + Redis 未同步增量）。
// Redis 不可用时仅返回 MySQL 值。
func GetPostViewCount(postID uint) (int64, error) {
	var post model.Post
	if err := database.DB.Select("view_count").First(&post, postID).Error; err != nil {
		return 0, err
	}

	if database.Redis == nil {
		return post.ViewCount, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := database.Redis.Get(ctx, viewCountKey(postID)).Result()
	if err == redis.Nil {
		return post.ViewCount, nil
	}
	if err != nil {
		// Redis 读取失败时降级，返回 MySQL 中的值
		return post.ViewCount, nil
	}

	delta, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return post.ViewCount, nil
	}

	return post.ViewCount + delta, nil
}

// SyncViewCountsToDB 将 Redis 中的浏览量增量同步到 MySQL。
// 同步成功后删除 Redis 中的增量 key。
func SyncViewCountsToDB() error {
	if database.Redis == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	keys, err := database.Redis.Keys(ctx, ViewCountRedisKeyPrefix+"*").Result()
	if err != nil || len(keys) == 0 {
		return nil
	}

	// 批量读取增量值
	values, err := database.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}

	// 按 postID 聚合增量
	updates := make(map[uint]int64)
	for i, key := range keys {
		if values[i] == nil {
			continue
		}
		postIDStr := strings.TrimPrefix(key, ViewCountRedisKeyPrefix)
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			continue
		}
		deltaStr, ok := values[i].(string)
		if !ok {
			continue
		}
		delta, err := strconv.ParseInt(deltaStr, 10, 64)
		if err != nil || delta <= 0 {
			continue
		}
		updates[uint(postID)] = delta
	}

	if len(updates) == 0 {
		return nil
	}

	// 在事务中批量更新 MySQL
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for postID, delta := range updates {
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).
				UpdateColumn("view_count", gorm.Expr("view_count + ?", delta)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 同步成功后删除 Redis 增量
	_, err = database.Redis.Del(ctx, keys...).Result()
	return err
}

// StartViewCountSync 启动浏览量后台同步任务。
// ctx 用于接收退出信号，确保服务关闭前完成最后一次同步。
// 返回的 stop 函数会阻塞等待后台 goroutine 完全退出。
func StartViewCountSync(ctx context.Context) (stop func()) {
	done := make(chan struct{})
	ticker := time.NewTicker(ViewCountSyncInterval)
	go func() {
		defer close(done)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := SyncViewCountsToDB(); err != nil {
					// 记录错误但不中断同步循环
					_ = err
				}
			case <-ctx.Done():
				// 退出前再同步一次，避免丢失数据
				_ = SyncViewCountsToDB()
				return
			}
		}
	}()
	return func() {
		<-done
	}
}
