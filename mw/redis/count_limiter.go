package redis

import (
	"strings"
	"time"
)

const (
	limiterTime = 2 * time.Hour // 2 小时

	loginLimit    = "loginLimit"
	registerLimit = "registerLimit"
	commentLimit  = "commentLimit"
	uploadLimit   = "uploadLimit"
	messageLimit  = "messageLimit"
)

const (
	loginMaxCount    = 5
	registerMaxCount = 5
	commentMaxCount  = 30
	uploadMaxCount   = 3
	messageMaxCount  = 50
)

func incrementLimiterCount(key string, maxCount int) bool {
	count, err := Rdb.Incr(Ctx, key).Result()
	if err != nil {
		return false
	}
	if count == 1 {
		// 首次计数，设置过期时间
		if _, err := Rdb.Expire(Ctx, key, limiterTime).Result(); err != nil {
			return false
		}
	}
	return count <= int64(maxCount)
}

func IncrementLoginLimiterCount(ip string) bool {
	return incrementLimiterCount(strings.Join([]string{loginLimit, ip}, Delimiter), loginMaxCount)
}

func IncrementRegisterLimiterCount(ip string) bool {
	return incrementLimiterCount(strings.Join([]string{registerLimit, ip}, Delimiter), registerMaxCount)
}

func IncrementCommentLimiterCount(ip string) bool {
	return incrementLimiterCount(strings.Join([]string{commentLimit, ip}, Delimiter), commentMaxCount)
}

func IncrementUploadLimiterCount(ip string) bool {
	return incrementLimiterCount(strings.Join([]string{uploadLimit, ip}, Delimiter), uploadMaxCount)
}

func IncrementMessageLimiterCount(ip string) bool {
	return incrementLimiterCount(strings.Join([]string{messageLimit, ip}, Delimiter), messageMaxCount)
}
