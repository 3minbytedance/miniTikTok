package mw

import (
	"context"
	"net/http"

	"douyin/common"
	"douyin/mw/redis"

	"github.com/cloudwego/hertz/pkg/app"
)

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// Auth 必须登录：解析 token 并校验 Redis 中的会话，续期后放行。
func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		claims, err := common.ParseToken(c.Query("token"))
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, Response{StatusCode: common.CodeInvalidToken, StatusMsg: common.MapErrMsg(common.CodeInvalidToken)})
			return
		}
		if !redis.TokenIsExisted(claims.ID) {
			c.Abort()
			c.JSON(http.StatusOK, Response{StatusCode: common.CodeInvalidToken, StatusMsg: common.MapErrMsg(common.CodeInvalidToken)})
			return
		}
		redis.SetToken(claims.ID, c.Query("token"))
		c.Set(common.ContextUserIDKey, claims.ID)
		c.Next(ctx)
	}
}

// AuthWithoutLogin 可选登录：携带合法 token 则注入用户 id，否则以游客（id=0）放行。
func AuthWithoutLogin() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		var userId uint
		var tokenValid bool
		if claims, err := common.ParseToken(c.Query("token")); err == nil && redis.TokenIsExisted(claims.ID) {
			userId = claims.ID
			tokenValid = true
			redis.SetToken(claims.ID, c.Query("token"))
		}
		c.Set(common.TokenValid, tokenValid)
		c.Set(common.ContextUserIDKey, userId)
		c.Next(ctx)
	}
}

// AuthBody 用于 multipart 上传：token 位于请求体。
func AuthBody() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := c.PostForm("token")
		claims, err := common.ParseToken(token)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, Response{StatusCode: common.CodeInvalidToken, StatusMsg: common.MapErrMsg(common.CodeInvalidToken)})
			return
		}
		if !redis.TokenIsExisted(claims.ID) {
			c.Abort()
			c.JSON(http.StatusOK, Response{StatusCode: common.CodeInvalidToken, StatusMsg: common.MapErrMsg(common.CodeInvalidToken)})
			return
		}
		redis.SetToken(claims.ID, token)
		c.Set(common.ContextUserIDKey, claims.ID)
		c.Next(ctx)
	}
}
