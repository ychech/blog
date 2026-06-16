// package main 是博客服务的入口包。
// 负责按顺序初始化日志、数据库、路由，然后启动 HTTP 服务。
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
	"blog/utils"
	"log"
)

func main() {
	// 加载配置：失败则直接退出
	config.Load()

	// 初始化 zap 日志：开发环境彩色输出，并记录调用位置
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}

	// 初始化 MySQL 与 Redis；Redis 失败只会降级缓存，不会阻断启动
	if err := database.Init(); err != nil {
		utils.Logger.Fatalf("数据库初始化失败: %v", err)
	}

	// 注册路由并获取 Gin 引擎实例
	r := router.Setup()

	addr := config.C.Server.ListenAddr()
	utils.Logger.Infof("博客服务启动: http://%s", addr)
	// 监听指定端口；若启动失败则记录 Fatal 日志并退出
	if err := r.Run(addr); err != nil {
		utils.Logger.Fatalf("服务启动失败: %v", err)
	}
}
