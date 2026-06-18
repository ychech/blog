// package router 负责注册所有 HTTP 路由。
// 按照“公开接口 / 需登录接口 / 静态资源”进行分组，并统一挂载中间件。
package router

import (
	"blog/config"
	"blog/handler"
	"blog/middleware"
	"blog/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 导入 swagger docs 包，确保 swag init 生成的文档被注册
	_ "blog/docs"
)

// Setup 配置并返回 Gin 路由引擎。
// 注册顺序：全局中间件 → 静态资源 → 公开接口 → JWT 保护接口 → 404 兜底。
func Setup() *gin.Engine {
	// 生产环境可设置为 ReleaseMode，以减少日志输出并提升性能
	// gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 全局中间件：恢复 panic -> 链路追踪 -> 记录请求日志 -> Prometheus 指标 -> 处理跨域 -> 接口限流 -> 审计日志
	// 注意顺序：Recovery 放在最前面，可以捕获后续中间件中的 panic；
	// Tracing 紧随其后，确保请求进入追踪范围；RateLimit 放在最后，对请求频率进行全局控制；
	// AuditLog 调用 c.Next() 后再记录，因此能获取到 JWTAuth 设置的用户上下文。
	r.Use(middleware.Recovery())
	r.Use(middleware.Tracing())
	r.Use(middleware.Logger())
	r.Use(middleware.Locale())
	r.Use(middleware.PrometheusMetrics())
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimit())
	r.Use(middleware.AuditLog())

	// 静态文件服务：上传的图片可通过 http://host/uploads/xxx.png 直接访问
	cfg := config.C.App
	r.Static("/uploads", cfg.UploadPath)

	// 健康检查接口，用于负载均衡、容器探针或服务启动后快速验证
	r.GET("/health", handler.HealthCheck)

	// Prometheus 指标接口，供 Prometheus 服务器抓取（访问 /metrics）
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(middleware.MetricsRegistry(), promhttp.HandlerOpts{})))

	// Swagger API 文档（访问 /swagger/index.html）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 初始化各模块的 HTTP 处理器
	authHandler := handler.NewAuthHandler()
	postHandler := handler.NewPostHandler()
	categoryHandler := handler.NewCategoryHandler()
	tagHandler := handler.NewTagHandler()
	commentHandler := handler.NewCommentHandler()
	uploadHandler := handler.NewUploadHandler()
	likeHandler := handler.NewLikeHandler()
	commentLikeHandler := handler.NewCommentLikeHandler()
	badgeHandler := handler.NewBadgeHandler()
	adminHandler := handler.NewAdminHandler()
	notificationHandler := handler.NewNotificationHandler()
	commentReportHandler := handler.NewCommentReportHandler()
	messageHandler := handler.NewMessageHandler()
	oauthHandler := handler.NewOAuthHandler()
	userFollowHandler := handler.NewUserFollowHandler()

	// 认证路由：注册、登录公开；获取/更新当前用户需要登录
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.GET("/oauth/github", oauthHandler.GitHubLogin)
		auth.GET("/oauth/github/callback", oauthHandler.GitHubCallback)
		auth.GET("/me", middleware.JWTAuth(), authHandler.Me)
		auth.PUT("/me", middleware.JWTAuth(), authHandler.UpdateProfile)
		auth.POST("/send-verification-email", middleware.JWTAuth(), authHandler.SendVerificationEmail)
		auth.POST("/verify-email", middleware.JWTAuth(), authHandler.VerifyEmail)
		auth.POST("/change-password", middleware.JWTAuth(), authHandler.ChangePassword)
		auth.POST("/refresh", middleware.JWTAuth(), authHandler.RefreshToken)
		auth.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
		auth.GET("/badges", middleware.JWTAuth(), badgeHandler.GetMyBadges)
		auth.GET("/favorites", middleware.JWTAuth(), postHandler.ListFavorites)
		auth.GET("/read-history", middleware.JWTAuth(), postHandler.ListReadHistory)
		auth.GET("/users", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminListUsers)
		auth.GET("/users/:id", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminGetUser)
		auth.PUT("/users/:id/role", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminUpdateUserRole)
		auth.PUT("/users/:id/status", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminUpdateUserStatus)
		auth.DELETE("/users/:id", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminDeleteUser)
		auth.GET("/stats", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminGetStats)
	}

	// 公开 API：文章列表、详情、热门文章、分类、标签、评论等均可匿名访问
	api := r.Group("/api")
	{
		api.GET("/posts", postHandler.List)
		api.GET("/posts/hot", postHandler.Hot)
		api.GET("/posts/:id", postHandler.Get)
		api.GET("/posts/:id/like", likeHandler.Status)
		api.GET("/categories", categoryHandler.List)
		api.GET("/tags", tagHandler.List)
		api.GET("/posts/:id/comments", commentHandler.ListByPost)
		api.GET("/comments/:id/like", commentLikeHandler.Status)
		api.GET("/comments/likes", commentLikeHandler.BatchStatus)

		// 勋章：公开列表、详情与单用户勋章
		api.GET("/badges", badgeHandler.List)
		api.GET("/badges/:id", badgeHandler.Get)
		api.GET("/users/:id/badges", badgeHandler.GetUserBadges)

		// 用户关注/粉丝
		api.GET("/users/:id/followers", userFollowHandler.Followers)
		api.GET("/users/:id/following", userFollowHandler.Following)
	}

	// 需要登录的 API：所有写操作都需要在请求头携带 Authorization: Bearer <token>
	authorized := r.Group("/api")
	authorized.Use(middleware.JWTAuth(), middleware.UserRateLimit())
	{
		// 文章管理
		authorized.POST("/posts", postHandler.Create)
		authorized.POST("/posts/drafts", postHandler.SaveDraft)
		authorized.GET("/posts/drafts", postHandler.GetDraft)
		authorized.POST("/posts/:id/favorite", postHandler.AddFavorite)
		authorized.DELETE("/posts/:id/favorite", postHandler.RemoveFavorite)
		authorized.PUT("/posts/:id", postHandler.Update)
		authorized.DELETE("/posts/:id", postHandler.Delete)

		// 用户关注
		authorized.POST("/users/:id/follow", userFollowHandler.Follow)
		authorized.DELETE("/users/:id/follow", userFollowHandler.Unfollow)

		// 评论管理
		authorized.POST("/comments", commentHandler.Create)
		authorized.PUT("/comments/:id", commentHandler.Update)
		authorized.POST("/comments/:id/reports", commentReportHandler.Create)

		// 文件上传
		authorized.POST("/uploads", uploadHandler.UploadImage)

		// 点赞
		authorized.POST("/posts/:id/like", likeHandler.Toggle)
		authorized.POST("/comments/:id/like", commentLikeHandler.Toggle)

		// 标签创建：登录用户即可创建新标签，方便写文章时使用
		authorized.POST("/tags", tagHandler.Create)

	}

	// 管理员 API：需要 JWT 认证 + 管理员角色
	admin := r.Group("/api")
	admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
	{
		// 分类管理
		admin.POST("/categories", categoryHandler.Create)
		admin.PUT("/categories/:id", categoryHandler.Update)
		admin.DELETE("/categories/:id", categoryHandler.Delete)

		// 标签管理：更新与删除由管理员控制
		admin.PUT("/tags/:id", tagHandler.Update)
		admin.DELETE("/tags/:id", tagHandler.Delete)

		// 评论管理：管理员可删除/置顶/加精任意评论
		admin.DELETE("/comments/:id", commentHandler.Delete)
		admin.PUT("/comments/:id/pin", commentHandler.PinComment)
		admin.PUT("/comments/:id/essence", commentHandler.EssenceComment)

		// 勋章管理
		admin.POST("/badges", badgeHandler.Create)
		admin.PUT("/badges/:id", badgeHandler.Update)
		admin.DELETE("/badges/:id", badgeHandler.Delete)
		admin.POST("/badges/award", badgeHandler.Award)
		admin.DELETE("/user-badges/:id", badgeHandler.Revoke)

		// 审计日志
		admin.GET("/audit-logs", adminHandler.ListAuditLogs)
		admin.GET("/audit-logs/export", adminHandler.ExportAuditLogs)

		// 评论举报审核
		admin.GET("/comment-reports", commentReportHandler.List)
		admin.PUT("/comment-reports/:id/status", commentReportHandler.UpdateStatus)
	}

	// 私信路由（需登录）
	authorized.POST("/messages", messageHandler.Send)
	authorized.GET("/messages/conversations", messageHandler.ListConversations)
	authorized.GET("/messages/unread-count", messageHandler.CountUnread)
	authorized.GET("/messages/:user_id", messageHandler.ListMessages)

	// 通知路由（需登录）
	api.GET("/notifications", middleware.JWTAuth(), notificationHandler.List)
	api.GET("/notifications/unread-count", middleware.JWTAuth(), notificationHandler.CountUnread)
	api.PUT("/notifications/:id/read", middleware.JWTAuth(), notificationHandler.MarkAsRead)
	api.PUT("/notifications/read-all", middleware.JWTAuth(), notificationHandler.MarkAllAsRead)
	api.DELETE("/notifications/:id", middleware.JWTAuth(), notificationHandler.Delete)

	// 404 兜底处理：所有未匹配的路由返回统一错误
	r.NoRoute(func(c *gin.Context) {
		utils.Error(c, utils.CodeNotFound, utils.T(utils.GetLocale(c), "route_not_found"))
	})

	return r
}
