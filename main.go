// package main 是博客服务的入口包。
// 负责按顺序初始化日志、数据库、路由，然后启动 HTTP 服务。
//
// 启动后监听 OS 信号（SIGINT/SIGTERM），收到信号时执行优雅关闭：
//   - 停止接收新连接
//   - 等待正在处理的请求完成（最多 5 秒）
//   - 关闭 MySQL 与 Redis 连接
//
// @title 个人博客 API
// @version 1.0
// @description 基于 Gin + GORM + MySQL + Redis 的个人博客后端 RESTful API
// @termsOfService http://localhost:8080
//
// @contact.name 博客作者
// @contact.url http://localhost:8080
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 请输入 "Bearer {token}"，例如：Bearer eyJhbGciOiJIUzI1NiIs...
package main

import (
	"blog/config"
	"blog/database"
	"blog/router"
	"blog/service"
	"blog/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 加载配置：失败则直接退出
	// 使用 Viper 支持多环境配置文件（config.dev.yaml / config.prod.yaml）
	cfg, err := config.LoadWithViper(config.LoadOptions{
		EnvFile: ".env",
	})
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	config.C = cfg

	// 初始化第三方 OAuth2 登录
	service.InitGitHubOAuth(config.C.OAuth)

	// 初始化 zap 日志：开发环境彩色输出，并记录调用位置
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 初始化 OpenTelemetry 链路追踪
	shutdownTracing, err := utils.InitTracing(config.C.Tracing)
	if err != nil {
		utils.Logger.Fatalf("链路追踪初始化失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(ctx); err != nil {
			utils.Logger.Errorf("链路追踪关闭失败: %v", err)
		}
	}()

	// 初始化 MySQL 与 Redis；Redis 失败只会降级缓存，不会阻断启动
	if err := database.Init(); err != nil {
		utils.Logger.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化 Meilisearch 搜索引擎；失败仅记录日志，不影响主服务启动
	if err := service.InitSearch(config.C.Meilisearch); err != nil {
		utils.Logger.Warnf("Meilisearch 初始化失败（非致命）: %v", err)
	}

	// 启动 WebSocket 实时通知 Hub
	service.StartNotificationHub()

	// 注册路由并获取 Gin 引擎实例
	r := router.Setup()

	// 启动后台任务：文章浏览量 Redis -> MySQL 定时同步
	bgCtx, bgCancel := context.WithCancel(context.Background())
	stopViewCountSync := service.StartViewCountSync(bgCtx)

	// 启动定时发布调度器
	go service.StartScheduler(bgCtx)

	addr := config.C.Server.ListenAddr()
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在独立 goroutine 中启动 HTTP 服务
	go func() {
		utils.Logger.Infof("博客服务启动: http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 监听系统退出信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	// SIGINT: Ctrl+C；SIGTERM: kill / docker stop
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Logger.Info("服务正在关闭，等待正在处理的请求完成...")

	// 取消后台任务，触发最后一次浏览量同步，并等待同步完成
	bgCancel()
	stopViewCountSync()

	// 关闭 WebSocket Hub，停止接收新连接并等待事件循环退出
	service.StopNotificationHub()

	// 设置 5 秒超时，强制关闭未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.Logger.Errorf("服务关闭失败: %v", err)
	}

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		utils.Logger.Errorf("数据库连接关闭失败: %v", err)
	}

	utils.Logger.Info("服务已安全退出")
	_ = utils.Logger.Sync()
}
