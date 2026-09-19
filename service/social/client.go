package main

import (
	"douyin/constant"
	"douyin/kitex_gen/videoapp/videoappservice"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
)

// videoClient 访问 video 域（作品数/点赞数/获赞数）。
var videoClient videoappservice.Client

func initVideoClient() error {
	r, err := etcd.NewEtcdResolver([]string{constant.EtcdAddr})
	if err != nil {
		return err
	}
	c, err := videoappservice.NewClient(
		constant.VideoServiceName,
		client.WithResolver(r),
	)
	if err != nil {
		return err
	}
	videoClient = c
	return nil
}
