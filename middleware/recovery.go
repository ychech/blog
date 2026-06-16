package middleware

import (
	"blog/utils"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 自定义 panic 恢复中间件。
// 捕获 handler 中的 panic，记录错误堆栈，并返回统一的 500 错误响应，防止服务崩溃。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				utils.Logger.Errorf("[PANIC] %v\n%s", err, string(stack))

				utils.InternalError(c, fmt.Sprintf("服务器内部错误: %v", err))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Cors 简单跨域中间件。
// 允许所有来源访问，并放行常见的 HTTP 方法与请求头；对 OPTIONS 预检请求直接返回 204。
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
