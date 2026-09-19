package redis

import (
	"go.uber.org/zap"
	"time"
)

const tokenExpireTime = 7 * 24 * time.Hour // 7天

// SetToken 设置 token
func SetToken(userId uint, token string) {
	if err := Rdb.Set(Ctx, tokenK(userId), token, tokenExpireTime).Err(); err != nil {
		zap.L().Error("SetToken failed", zap.Error(err))
	}
}

// TokenIsExisted 判断用户对应的 token 是否存在
func TokenIsExisted(userId uint) bool {
	exists, err := Rdb.Exists(Ctx, tokenK(userId)).Result()
	if err != nil {
		return false
	}
	return exists == 1
}
