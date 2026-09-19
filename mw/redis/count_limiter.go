package redis

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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

// limiterScript 原子完成自增与首次设置过期时间：
// 若拆成 INCR + EXPIRE 两步，中间失败会留下永不过期的计数 key，导致该 IP 被永久限流。
var limiterScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func incrementLimiterCount(key string, maxCount int) bool {
	count, err := limiterScript.Run(Ctx, Rdb, []string{key}, int(limiterTime.Seconds())).Int64()
	if err != nil {
		return false
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
