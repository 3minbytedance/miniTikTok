package main

import (
	"net"
	"strconv"

	"douyin/common"
	"douyin/config"
	"douyin/constant"
	"douyin/dal/mysql"
	"douyin/kitex_gen/social/socialservice"
	"douyin/logger"
	"douyin/mw/redis"
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
		observability.ServiceName(obsCfg, constant.SocialServiceName),
		observability.CollectorAddr(obsCfg), observability.TraceEnabled(obsCfg), observability.MetricsEnabled(obsCfg),
	)
	defer shutdown()

	// 2. 存储与缓存（social 独立数据库：douyin_social）
	if err := mysql.InitSocial(config.Conf); err != nil {
		zap.L().Fatal("init mysql failed", zap.Error(err))
	}
	if err := redis.Init(config.Conf); err != nil {
		zap.L().Fatal("init redis failed", zap.Error(err))
	}

	// 3. 雪花 ID 与 Bloom
	node, _ := strconv.ParseInt(config.Conf.Node, 10, 64)
	if err := common.InitSnowflake(node); err != nil {
		zap.L().Fatal("init snowflake failed", zap.Error(err))
	}
	InitUserBloomFilter()
	InitRelationFollowIdFilter()
	InitRelationFollowerIdFilter()
	LoadUsernamesToBloomFilter()
	LoadRelationFollowIdToBloomFilter()
	LoadRelationFollowerIdToBloomFilter()

	// 4. 跨域 RPC client
	if err := initVideoClient(); err != nil {
		zap.L().Fatal("init video client failed", zap.Error(err))
	}

	// 5. 注册中心 + Kitex Server
	r, err := etcd.NewEtcdRegistry([]string{constant.EtcdAddr})
	if err != nil {
		zap.L().Fatal("new etcd registry failed", zap.Error(err))
	}
	addr, _ := net.ResolveTCPAddr("tcp", constant.SocialServicePort)
	basicInfo := observability.EndpointInfo(constant.SocialServiceName)
	svr := socialservice.NewServer(
		&SocialServiceImpl{},
		server.WithServerBasicInfo(&basicInfo),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(tracing.NewServerSuite()),
	)
	zap.L().Info("social service starting", zap.String("addr", constant.SocialServicePort))
	if err := svr.Run(); err != nil {
		zap.L().Fatal("social service stopped", zap.Error(err))
	}
}
