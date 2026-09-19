package main

import (
	"context"
	"douyin/common"
	"douyin/observability"
	"douyin/service/api/biz/comment"
	"douyin/service/api/biz/favorite"
	"douyin/service/api/biz/message"
	"douyin/service/api/biz/relation"
	"douyin/service/api/biz/user"
	"douyin/service/api/biz/video"
	"douyin/service/api/mw"
	"path"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/etag"
)

// register 注册全部 HTTP 路由，保持原 /douyin/* 对外契约不变。
func register(h *server.Hertz) {
	// 观测性：请求 ID + trace_id 日志、Prometheus 指标。
	h.Use(common.AccessLog())
	h.Use(observability.RequestIDMiddleware())
	h.Use(observability.HTTPMetricsMiddleware())
	h.Use(etag.New())

	// Prometheus 指标与本地静态资源（OSS 未配置时的视频/封面降级）。
	// URL /static/<file> 直接映射到本地降级目录（hertz h.Static 会重复拼接前缀，故手动实现）。
	h.GET("/metrics", observability.MetricsHandler())
	staticRoot := common.LocalFallbackDir()
	h.GET("/static/*filepath", func(ctx context.Context, c *app.RequestContext) {
		// path.Clean 锚定根目录，防止 /static/../ 路径穿越
		file := path.Clean("/" + c.Param("filepath"))
		c.File(staticRoot + file)
	})

	douyin := h.Group("/douyin")

	// user（social 域）
	userGroup := douyin.Group("/user")
	{
		userGroup.POST("/register/", mw.RequestRegisterLimiter(), user.Register)
		userGroup.POST("/login/", mw.RequestLoginLimiter(), user.Login)
		userGroup.GET("/", mw.RateLimiter(), mw.AuthWithoutLogin(), user.Info)
	}

	// video（video 域）
	douyin.GET("/feed/", mw.RateLimiter(), mw.AuthWithoutLogin(), video.FeedList)
	publishGroup := douyin.Group("/publish")
	{
		publishGroup.GET("/list/", mw.RateLimiter(), mw.AuthWithoutLogin(), video.GetPublishList)
		publishGroup.POST("/action/", mw.RequestUploadLimiter(), mw.AuthBody(), video.Publish)
	}

	// comment（video 域）
	commentGroup := douyin.Group("/comment")
	{
		commentGroup.POST("/action/", mw.RequestCommentLimiter(), mw.Auth(), comment.Action)
		commentGroup.GET("/list/", mw.RateLimiter(), mw.AuthWithoutLogin(), comment.List)
	}

	// favorite（video 域）
	favoriteGroup := douyin.Group("/favorite", mw.RateLimiter())
	{
		favoriteGroup.POST("/action/", mw.Auth(), favorite.Action)
		favoriteGroup.GET("/list/", mw.AuthWithoutLogin(), favorite.List)
	}

	// relation（social 域）
	relationGroup := douyin.Group("/relation")
	{
		relationGroup.POST("/action/", mw.Auth(), relation.Action)
		relationGroup.GET("/follow/list/", mw.AuthWithoutLogin(), relation.FollowList)
		relationGroup.GET("/follower/list/", mw.AuthWithoutLogin(), relation.FollowerList)
		relationGroup.GET("/friend/list/", mw.Auth(), relation.FriendList)
	}

	// message（message 域）
	messageGroup := douyin.Group("/message")
	{
		messageGroup.POST("/action/", mw.RequestMessageLimiter(), mw.Auth(), message.Action)
		messageGroup.GET("/chat/", mw.Auth(), message.Chat)
	}
}
