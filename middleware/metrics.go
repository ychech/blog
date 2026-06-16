// package middleware 提供 Gin 中间件。
//
// 本文件实现 Prometheus 指标收集：
//   - http_requests_total：按方法、路径、状态码统计请求总数
//   - http_request_duration_seconds：按方法、路径统计请求耗时分布
//   - http_request_size_bytes：按方法、路径统计请求体大小分布
//   - http_response_size_bytes：按方法、路径统计响应体大小分布
//
// 指标通过 /metrics 接口暴露，供 Prometheus 抓取。
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// registry 是本项目专用的 Prometheus 注册器。
	// 使用私有注册器可以避免引入全局默认注册器中可能存在的冲突指标。
	registry = prometheus.NewRegistry()

	requestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数，按方法、路径、状态码分类",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求处理耗时分布（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	requestSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP 请求体大小分布（字节）",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	responseSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP 响应体大小分布（字节）",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)
)

// MetricsRegistry 返回项目私有的 Prometheus 注册器。
// 路由层通过此注册器暴露 /metrics 接口。
func MetricsRegistry() *prometheus.Registry {
	return registry
}

// PrometheusMetrics 返回 Gin 中间件，用于收集 Prometheus 指标。
//
// 注意：
//   - 该中间件需要放在 Logger 之后，这样可以拿到准确的状态码；
//   - path 标签使用 c.FullPath()，避免路径参数被展开成具体值（如 /api/posts/1 统一记为 /api/posts/:id）。
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		// 请求体大小
		requestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))

		c.Next()

		// 状态码、耗时、响应体大小
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		requestsTotal.WithLabelValues(method, path, status).Inc()
		requestDuration.WithLabelValues(method, path).Observe(duration)
		responseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}
