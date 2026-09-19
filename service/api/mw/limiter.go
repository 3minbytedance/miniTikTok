package mw

import (
	"context"
	"net/http"
	"strings"

	"douyin/common"
	"douyin/mw/redis"

	"github.com/cloudwego/hertz/pkg/app"
)

func reject(c *app.RequestContext) {
	c.JSON(http.StatusBadRequest, Response{
		StatusCode: common.CodeLimiterCount,
		StatusMsg:  common.MapErrMsg(common.CodeLimiterCount),
	})
	c.Abort()
}

// RateLimiter 基于 IP 的令牌桶限流。
func RateLimiter() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ipAddr := strings.Split(c.ClientIP(), ":")[0]
		permit, _, err := redis.AcquireBucket(ipAddr)
		if err != nil || !permit {
			reject(c)
			return
		}
		c.Next(ctx)
	}
}

func countLimiter(inc func(ip string) bool) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !inc(c.ClientIP()) {
			reject(c)
			return
		}
		c.Next(ctx)
	}
}

func RequestLoginLimiter() app.HandlerFunc { return countLimiter(redis.IncrementLoginLimiterCount) }
func RequestRegisterLimiter() app.HandlerFunc {
	return countLimiter(redis.IncrementRegisterLimiterCount)
}
func RequestCommentLimiter() app.HandlerFunc { return countLimiter(redis.IncrementCommentLimiterCount) }
func RequestUploadLimiter() app.HandlerFunc  { return countLimiter(redis.IncrementUploadLimiterCount) }
func RequestMessageLimiter() app.HandlerFunc { return countLimiter(redis.IncrementMessageLimiterCount) }
