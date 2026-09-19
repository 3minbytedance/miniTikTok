package rpc

import (
	"douyin/constant"
	"douyin/kitex_gen/message/messageservice"
	"douyin/kitex_gen/social/socialservice"
	"douyin/kitex_gen/videoapp/videoappservice"

	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var (
	Social  socialservice.Client
	Video   videoappservice.Client
	Message messageservice.Client
)

// Init 初始化三个下游 RPC client（etcd 服务发现 + 链路追踪）。
func Init() error {
	r, err := etcd.NewEtcdResolver([]string{constant.EtcdAddr})
	if err != nil {
		return err
	}

	sc, err := socialservice.NewClient(
		constant.SocialServiceName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
	)
	if err != nil {
		return err
	}
	Social = sc

	vc, err := videoappservice.NewClient(
		constant.VideoServiceName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
	)
	if err != nil {
		return err
	}
	Video = vc

	mc, err := messageservice.NewClient(
		constant.MessageServiceName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
	)
	if err != nil {
		return err
	}
	Message = mc
	return nil
}
