// package middleware 提供 Gin 中间件：JWT 认证、请求日志、panic 恢复、跨域处理。
package middleware

import (
	"blog/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger zap 请求日志中间件。
// 记录每个请求的方法、路径、状态码、耗时、客户端 IP
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		utils.Logger.Infof("[%s] %s %d %s %s",
			method,
			path,
			status,
			cost,
			clientIP,
		)
	}
}
