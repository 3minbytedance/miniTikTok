package constant

import "os"

// EtcdAddr 注册中心地址。容器部署时通过 DOUYIN_ETCD_ADDR 覆盖（如 etcd:2379）。
var EtcdAddr = getEnv("DOUYIN_ETCD_ADDR", "127.0.0.1:2379")

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// 服务名
const ApiServiceName = "api-service"
const SocialServiceName = "social-service"
const VideoServiceName = "video-service"
const MessageServiceName = "message-service"

// 服务监听端口
const ApiServicePort = ":8080"
const SocialServicePort = ":4002"
const VideoServicePort = ":4003"
const MessageServicePort = ":4005"
