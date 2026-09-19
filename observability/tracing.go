package observability

import (
	"context"

	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// tracerProvider 全局 TracerProvider 关闭函数；未初始化 collector 时为 no-op。
var shutdownFunc func(context.Context) error

// InitTracing 初始化 OpenTelemetry。
// collectorAddr 为空时不创建 OTLP exporter（span 为 no-op，但 trace 上下文仍在服务间透传）。
func InitTracing(serviceName, collectorAddr string, traceEnabled, metricsEnabled bool) func() {
	if !traceEnabled || collectorAddr == "" {
		zap.L().Info("opentelemetry exporter disabled (no collector configured), trace propagation still active")
		return func() {}
	}

	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint(collectorAddr),
		provider.WithInsecure(),
		provider.WithEnableTracing(traceEnabled),
		provider.WithEnableMetrics(metricsEnabled),
	)
	shutdownFunc = p.Shutdown
	otel.SetTracerProvider(otel.GetTracerProvider())
	zap.L().Info("opentelemetry provider initialized", zap.String("service", serviceName), zap.String("collector", collectorAddr))

	return func() {
		if shutdownFunc != nil {
			_ = shutdownFunc(context.Background())
		}
	}
}
