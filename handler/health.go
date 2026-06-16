// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现健康检查接口，用于负载均衡、容器探针或监控系统判断服务可用性。
package handler

import (
	"blog/database"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck 返回服务健康状态。
//
// 检查项：
//   - MySQL：通过 sqlDB.Ping() 探测连接是否可用
//   - Redis：通过 Ping 命令探测连接是否可用（Redis 为可选依赖）
//
// 响应格式：
//
//	{
//	  "status": "ok" | "unhealthy",
//	  "time": "2024-01-01T12:00:00+08:00",
//	  "dependencies": {
//	    "mysql": true,
//	    "redis": true
//	  }
//	}
//
// HTTP 状态码：
//   - 200：所有关键依赖正常
//   - 503：MySQL 不可用（Redis 不可用不影响整体状态，仅标记依赖异常）
func HealthCheck(c *gin.Context) {
	status := "ok"
	httpStatus := http.StatusOK

	// 探测 MySQL
	mysqlOK := false
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil && sqlDB != nil {
			if err := sqlDB.Ping(); err == nil {
				mysqlOK = true
			}
		}
	}

	// 探测 Redis（可选依赖，失败不影响整体健康状态）
	redisOK := false
	if database.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := database.Redis.Ping(ctx).Err(); err == nil {
			redisOK = true
		}
	}

	if !mysqlOK {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status": status,
		"time":   time.Now().Format(time.RFC3339),
		"dependencies": gin.H{
			"mysql": mysqlOK,
			"redis": redisOK,
		},
	})
}
