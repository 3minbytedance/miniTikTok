package observability

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	cacheRequestsTotal  *prometheus.CounterVec
)

func init() {
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "douyin_http_requests_total",
		Help: "Total number of HTTP requests handled by the API gateway.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "douyin_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	cacheRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "douyin_cache_requests_total",
		Help: "Redis cache lookups by business and result (hit/miss).",
	}, []string{"business", "result"})

	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, cacheRequestsTotal)
}

// HTTPMetricsMiddleware 记录网关 QPS、延迟、状态码。
func HTTPMetricsMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)

		method := string(c.Method())
		path := string(c.Path())
		status := strconv.Itoa(c.Response.StatusCode())
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	}
}

// RecordCache 记录一次缓存查询（business 如 favorite/comment/relation，result 为 hit/miss）。
func RecordCache(business, result string) {
	cacheRequestsTotal.WithLabelValues(business, result).Inc()
}

// MetricsHandler 返回供 Hertz 路由挂载的 /metrics 处理器。
func MetricsHandler() app.HandlerFunc {
	return adaptor.HertzHandler(promhttp.Handler())
}
