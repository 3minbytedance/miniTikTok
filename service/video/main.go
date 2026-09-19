package main

import (
	"net"
	"strconv"

	"douyin/common"
	"douyin/config"
	"douyin/constant"
	"douyin/dal/mysql"
	"douyin/kitex_gen/videoapp/videoappservice"
	"douyin/logger"
	"douyin/mw/kafka"
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
		observability.ServiceName(obsCfg, constant.VideoServiceName),
		observability.CollectorAddr(obsCfg), observability.TraceEnabled(obsCfg), observability.MetricsEnabled(obsCfg),
	)
	defer shutdown()

	// 2. 存储、缓存与消息队列
	if err := mysql.Init(config.Conf); err != nil {
		zap.L().Fatal("init mysql failed", zap.Error(err))
	}
	if err := redis.Init(config.Conf); err != nil {
		zap.L().Fatal("init redis failed", zap.Error(err))
	}
	if err := kafka.Init(config.Conf); err != nil {
		zap.L().Fatal("init kafka failed", zap.Error(err))
	}

	// 3. 雪花 ID、敏感词与 Bloom
	node, _ := strconv.ParseInt(config.Conf.Node, 10, 64)
	if err := common.InitSnowflake(node); err != nil {
		zap.L().Fatal("init snowflake failed", zap.Error(err))
	}
	if err := common.InitSensitiveFilter(); err != nil {
		zap.L().Warn("init sensitive filter failed", zap.Error(err))
	}
	common.InitCommentBloomFilter()
	common.InitWorkCountFilter()
	common.InitIsFavoriteFilter()
	common.InitFavoriteVideoIdFilter()
	common.LoadCommentVideoIdToBloomFilter()
	common.LoadWorkCountToBloomFilter()
	common.LoadIsFavoriteToBloomFilter()
	common.LoadFavoriteVideoIdToBloomFilter()

	// 4. kafka 视频发布消费者
	kafka.InitVideoKafka()

	// 5. 跨域 RPC client
	if err := initSocialClient(); err != nil {
		zap.L().Fatal("init social client failed", zap.Error(err))
	}

	// 6. 注册中心 + Kitex Server
	r, err := etcd.NewEtcdRegistry([]string{constant.EtcdAddr})
	if err != nil {
		zap.L().Fatal("new etcd registry failed", zap.Error(err))
	}
	addr, _ := net.ResolveTCPAddr("tcp", constant.VideoServicePort)
	basicInfo := observability.EndpointInfo(constant.VideoServiceName)
	svr := videoappservice.NewServer(
		&VideoAppServiceImpl{},
		server.WithServerBasicInfo(&basicInfo),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(tracing.NewServerSuite()),
	)
	zap.L().Info("video service starting", zap.String("addr", constant.VideoServicePort))
	if err := svr.Run(); err != nil {
		zap.L().Fatal("video service stopped", zap.Error(err))
	}
}
