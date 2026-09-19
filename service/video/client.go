package main

import (
	"douyin/constant"
	"douyin/kitex_gen/social/socialservice"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
)

// socialClient 访问 social 域（作者信息）。
var socialClient socialservice.Client

func initSocialClient() error {
	r, err := etcd.NewEtcdResolver([]string{constant.EtcdAddr})
	if err != nil {
		return err
	}
	c, err := socialservice.NewClient(
		constant.SocialServiceName,
		client.WithResolver(r),
	)
	if err != nil {
		return err
	}
	socialClient = c
	return nil
}
