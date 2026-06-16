// package router 负责注册所有 HTTP 路由。
// 按照“公开接口 / 需登录接口 / 静态资源”进行分组，并统一挂载中间件。
package router

import (
	"blog/config"
	"blog/handler"
	"blog/middleware"
	"blog/utils"

	"github.com/gin-gonic/gin"
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

	// 全局中间件：恢复 panic -> 记录请求日志 -> 处理跨域 -> 接口限流
	// 注意顺序：Recovery 放在最前面，可以捕获后续中间件中的 panic；
	// RateLimit 放在最后，对请求频率进行全局控制。
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimit())

	// 静态文件服务：上传的图片可通过 http://host/uploads/xxx.png 直接访问
	cfg := config.C.App
	r.Static("/uploads", cfg.UploadPath)

	// 健康检查接口，用于负载均衡、容器探针或服务启动后快速验证
	r.GET("/health", handler.HealthCheck)

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

	// 认证路由：注册、登录公开；获取/更新当前用户需要登录
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.JWTAuth(), authHandler.Me)
		auth.PUT("/me", middleware.JWTAuth(), authHandler.UpdateProfile)
		auth.GET("/badges", middleware.JWTAuth(), badgeHandler.GetMyBadges)
		auth.GET("/users", middleware.JWTAuth(), middleware.AdminAuth(), authHandler.AdminListUsers)
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

		// 勋章：公开列表与单用户勋章
		api.GET("/badges", badgeHandler.List)
		api.GET("/users/:id/badges", badgeHandler.GetUserBadges)
	}

	// 需要登录的 API：所有写操作都需要在请求头携带 Authorization: Bearer <token>
	authorized := r.Group("/api")
	authorized.Use(middleware.JWTAuth())
	{
		// 文章管理
		authorized.POST("/posts", postHandler.Create)
		authorized.PUT("/posts/:id", postHandler.Update)
		authorized.DELETE("/posts/:id", postHandler.Delete)

		// 评论管理
		authorized.POST("/comments", commentHandler.Create)

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

		// 标签管理：删除仍由管理员控制
		admin.DELETE("/tags/:id", tagHandler.Delete)

		// 评论管理：管理员可删除任意评论
		admin.DELETE("/comments/:id", commentHandler.Delete)

		// 勋章管理
		admin.POST("/badges", badgeHandler.Create)
		admin.PUT("/badges/:id", badgeHandler.Update)
		admin.DELETE("/badges/:id", badgeHandler.Delete)
		admin.POST("/badges/award", badgeHandler.Award)
		admin.DELETE("/user-badges/:id", badgeHandler.Revoke)
	}

	// 404 兜底处理：所有未匹配的路由返回统一错误
	r.NoRoute(func(c *gin.Context) {
		utils.Error(c, utils.CodeNotFound, "接口不存在")
	})

	return r
}
