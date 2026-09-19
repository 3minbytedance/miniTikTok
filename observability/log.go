package observability

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	// RequestIDHeader 请求 ID 的 HTTP header。
	RequestIDHeader = "X-Request-Id"
	requestIDCtxKey = "request_id"
)

// RequestIDMiddleware 为每个入口请求注入 request_id，并写入响应 header。
func RequestIDMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		rid := string(c.GetHeader(RequestIDHeader))
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(requestIDCtxKey, rid)
		c.Response.Header.Set(RequestIDHeader, rid)
		c.Next(ctx)
	}
}

// Logger 从 context 中提取 trace_id / request_id，返回带关联字段的 logger。
func Logger(ctx context.Context) *zap.Logger {
	fields := make([]zap.Field, 0, 2)
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
	}
	if c, ok := ctx.Value(requestIDCtxKey).(string); ok && c != "" {
		fields = append(fields, zap.String("request_id", c))
	}
	return zap.L().With(fields...)
}
