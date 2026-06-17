// package middleware 提供 Gin 中间件。
//
// 本文件将 OpenTelemetry 链路追踪集成到 Gin 请求链路中。
package middleware

import (
	"blog/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Tracing 返回 Gin 链路追踪中间件。
// 当 tracing.enabled 为 false 时返回空中间件，不产生任何开销。
func Tracing() gin.HandlerFunc {
	if !config.C.Tracing.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	serviceName := config.C.Tracing.ServiceName
	if serviceName == "" {
		serviceName = config.DefaultTracingServiceName
	}

	return otelgin.Middleware(serviceName)
}
