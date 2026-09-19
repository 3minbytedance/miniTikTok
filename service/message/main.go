package main

import (
	"net"

	"douyin/config"
	"douyin/constant"
	"douyin/dal/mongo"
	"douyin/kitex_gen/message/messageservice"
	"douyin/logger"
	"douyin/observability"

	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
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
		observability.ServiceName(obsCfg, constant.MessageServiceName),
		observability.CollectorAddr(obsCfg), observability.TraceEnabled(obsCfg), observability.MetricsEnabled(obsCfg),
	)
	defer shutdown()

	// 2. MongoDB（消息直接读写，无缓存、无 MQ）
	if err := mongo.Init(config.Conf); err != nil {
		zap.L().Fatal("init mongo failed", zap.Error(err))
	}

	// 3. 跨域 RPC client（好友关系校验）
	if err := initSocialClient(); err != nil {
		zap.L().Fatal("init social client failed", zap.Error(err))
	}

	// 4. 注册中心 + Kitex Server
	r, err := etcd.NewEtcdRegistry([]string{constant.EtcdAddr})
	if err != nil {
		zap.L().Fatal("new etcd registry failed", zap.Error(err))
	}
	addr, _ := net.ResolveTCPAddr("tcp", constant.MessageServicePort)
	basicInfo := observability.EndpointInfo(constant.MessageServiceName)
	svr := messageservice.NewServer(
		&MessageServiceImpl{},
		server.WithServerBasicInfo(&basicInfo),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(tracing.NewServerSuite()),
	)
	zap.L().Info("message service starting", zap.String("addr", constant.MessageServicePort))
	if err := svr.Run(); err != nil {
		zap.L().Fatal("message service stopped", zap.Error(err))
	}
}
