package redis

import (
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	timeNow                 = time.Now
	bucketSize      int     = 60 // 令牌桶的容量
	refillPerSecond float64 = 30 // 每隔多长时间添加令牌
	refillToken     int     = 10 // 每次令牌桶填充操作时添加的令牌数量

	tokenBucket = "tokenBucket" // key名
)

func Ping() error {
	if _, err := Rdb.Ping(Ctx).Result(); err != nil {
		return err
	}
	return nil
}

func AcquireBucket(key string) (bool, int, error) {
	now := timeNow()
	baseSlice := []string{tokenBucket, key}
	cacheKey := strings.Join(baseSlice, Delimiter)

	remain, err := runScript(
		[]string{cacheKey},
		now.Unix(),
		refillToken,
		refillPerSecond,
		bucketSize,
	)
	if err != nil {
		return false, 0, err
	}
	if remain.(int64) < 0 {
		return false, bucketSize, nil
	}
	return true, int(remain.(int64)), nil
}

func runScript(keys []string, args ...interface{}) (interface{}, error) {
	val, err := redis.NewScript(rateScript).Run(Ctx, Rdb, keys, args).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	return val, nil
}

func ClearAll() error {
	if _, err := Rdb.FlushAll(Ctx).Result(); err != nil {
		return err
	}
	return nil
}

const (
	rateScript string = `
	local capacity = tonumber(ARGV[4])
	local refill = tonumber(ARGV[3])
	local refillToken = tonumber(ARGV[2])
	local ts = tonumber(ARGV[1])
	local lastUpdate = ts
	local remainToken = capacity

	local last = redis.call('HMGET', KEYS[1], 'ts', 'tokens')
	if last[1] then
		local lastTs = tonumber(last[1])
		local lastTokens = tonumber(last[2])
		local refillCount = math.floor((ts - lastTs) / refill)


		remainToken = math.min(capacity, lastTokens + (refillCount * refillToken))
		lastUpdate = math.min(ts, lastTs + (refillCount * refill))
	end

	if remainToken >= 0 then
			remainToken = remainToken - 1
	end
	redis.call('HMSET', KEYS[1], 'ts', ts, 'tokens', remainToken)
	redis.call('EXPIRE', KEYS[1], math.ceil(capacity / refill))
	return remainToken
	`
)
