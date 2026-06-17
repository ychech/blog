// package utils 提供项目通用工具。
//
// 本文件实现基于 OpenTelemetry + OTLP/HTTP 的链路追踪初始化与关闭。
// 将 trace 数据推送到 Jaeger、Tempo 等兼容 OTLP 的收集器。
package utils

import (
	"blog/config"
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.41.0"
)

var tracerProvider *sdktrace.TracerProvider

// getAppEnv 获取当前运行环境。
func getAppEnv() string {
	if v := os.Getenv("BLOG_APP_ENV"); v != "" {
		return v
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		return v
	}
	return "dev"
}

// InitTracing 初始化 OpenTelemetry 链路追踪。
// 当配置中 tracing.enabled 为 false 时，直接返回 nil，不创建 exporter。
func InitTracing(cfg config.TracingConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = config.DefaultTracingEndpoint
	}

	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 OTLP trace exporter 失败: %w", err)
	}

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = config.DefaultTracingServiceName
	}

	sampler := sdktrace.TraceIDRatioBased(cfg.SampleRate)
	if cfg.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else if cfg.SampleRate >= 1 {
		sampler = sdktrace.AlwaysSample()
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			attribute.String("environment", getAppEnv()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 tracing resource 失败: %w", err)
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tracerProvider)

	Logger.Info("链路追踪已启用，OTLP endpoint: " + endpoint)

	return tracerProvider.Shutdown, nil
}
