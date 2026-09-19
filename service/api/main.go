package main

import (
	"douyin/config"
	"douyin/constant"
	"douyin/logger"
	"douyin/mw/redis"
	"douyin/observability"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/zap"
)

func main() {
	// 1. 配置与日志
	if err := config.Init(); err != nil {
		panic(err)
	}
	if err := logger.Init(config.Conf.LogConfig, config.Conf.Mode); err != nil {
		panic(err)
	}
	obsCfg := config.Conf.ObsConfig
	shutdown := observability.InitTracing(
		observability.ServiceName(obsCfg, constant.ApiServiceName),
		observability.CollectorAddr(obsCfg), observability.TraceEnabled(obsCfg), observability.MetricsEnabled(obsCfg),
	)
	defer shutdown()

	// 2. Redis（token 会话与网关限流）
	if err := redis.Init(config.Conf); err != nil {
		zap.L().Fatal("init redis failed", zap.Error(err))
	}

	// 3. 下游 RPC client
	if err := rpc.Init(); err != nil {
		zap.L().Fatal("init rpc clients failed", zap.Error(err))
	}

	// 4. HTTP Server（上传文件上限 50MB）
	h := server.Default(
		server.WithHostPorts(constant.ApiServicePort),
		server.WithMaxRequestBodySize(50*1024*1024),
	)
	register(h)
	zap.L().Info("api gateway starting", zap.String("addr", constant.ApiServicePort))
	h.Spin()
}
