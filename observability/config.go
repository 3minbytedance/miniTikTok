package observability

import (
	"douyin/config"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// ServiceName 返回可观测性服务名，缺省用传入的 fallback。
func ServiceName(cfg *config.ObsConfig, fallback string) string {
	if cfg != nil && cfg.ServiceName != "" {
		return cfg.ServiceName + "-" + fallback
	}
	return fallback
}

// CollectorAddr 返回 otel collector 地址。
func CollectorAddr(cfg *config.ObsConfig) string {
	if cfg == nil {
		return ""
	}
	return cfg.CollectorAddr
}

// TraceEnabled 是否开启链路追踪。
func TraceEnabled(cfg *config.ObsConfig) bool {
	return cfg == nil || cfg.TraceEnabled
}

// MetricsEnabled 是否开启指标。
func MetricsEnabled(cfg *config.ObsConfig) bool {
	return cfg == nil || cfg.MetricsEnabled
}

// EndpointInfo 构造 Kitex 服务基础信息。
func EndpointInfo(serviceName string) rpcinfo.EndpointBasicInfo {
	return rpcinfo.EndpointBasicInfo{ServiceName: serviceName}
}
