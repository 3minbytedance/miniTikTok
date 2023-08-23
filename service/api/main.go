// Code generated by hertz generator.

package main

import (
	"douyin/config"
	"douyin/constant"
	"douyin/logger"
	"douyin/mw/redis"
	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func main() {
	// OpenTelemetry 链路跟踪
	//p := provider.NewOpenTelemetryProvider(
	//	provider.WithServiceName(config.CommentServiceName),
	//	provider.WithExportEndpoint("localhost:4317"),
	//	provider.WithInsecure(),
	//)
	//defer p.Shutdown(context.Background())

	// 加载配置
	if err := config.Init(); err != nil {
		zap.L().Error("Load config failed, err:%v\n", zap.Error(err))
		return
	}
	// 加载日志
	if err := logger.Init(config.Conf.LogConfig, config.Conf.Mode); err != nil {
		zap.L().Error("Init logger failed, err:%v\n", zap.Error(err))
		return
	}

	// 初始化中间件: redis
	if err := redis.Init(config.Conf); err != nil {
		zap.L().Error("Init redis failed, err:%v\n", zap.Error(err))
		return
	}
	h := server.Default(
		server.WithHostPorts(constant.ApiServicePort),
		server.WithMaxRequestBodySize(50*1024*1024),
	)

	customizedRegister(h)
	h.Spin()
}
