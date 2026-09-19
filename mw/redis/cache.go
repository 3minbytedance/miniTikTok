package redis

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"douyin/observability"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 分布式锁参数（仅用于冷缓存预热，防止击穿）。
const (
	lockTTL      = 3 * time.Second
	lockRetryGap = 50 * time.Millisecond
	lockMaxRetry = 6
)

// JitterTTL 在基础 TTL 上增加 0~20% 随机抖动，防止缓存雪崩。
func JitterTTL(base time.Duration) time.Duration {
	if base <= 0 {
		base = RdbExpireTime
	}
	jitter := int64(base) / 5
	if jitter <= 0 {
		return base
	}
	return base + time.Duration(rand.Int63n(jitter))
}

// CurrentUnix 当前 unix 秒。
func CurrentUnix() int64 {
	return time.Now().Unix()
}

// ShortTTL 聚合类/变化频繁数据使用较短 TTL。
func ShortTTL() time.Duration {
	return JitterTTL(30 * time.Second)
}

// Invalidate 删除一个或多个缓存 key（写操作后调用）。
// key 不存在不视为错误。
func Invalidate(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	if err := Rdb.Del(ctx, keys...).Err(); err != nil {
		zap.L().Warn("redis invalidate keys failed", zap.Strings("keys", keys), zap.Error(err))
	}
}

func acquireLock(ctx context.Context, key string) bool {
	ok, err := Rdb.SetNX(ctx, lockK(key), "1", lockTTL).Result()
	if err != nil {
		zap.L().Warn("redis acquire lock failed", zap.String("key", key), zap.Error(err))
		return false
	}
	return ok
}

func releaseLock(ctx context.Context, key string) {
	if err := Rdb.Del(ctx, lockK(key)).Err(); err != nil {
		zap.L().Warn("redis release lock failed", zap.String("key", key), zap.Error(err))
	}
}

// LoadInt 通用整数 cache-aside：未命中加锁、双检、回源、回填（带 TTL 抖动）。
func LoadInt(ctx context.Context, key, business string, ttl time.Duration, load func(context.Context) (int64, error)) (int64, error) {
	if v, err := Rdb.Get(ctx, key).Int64(); err == nil {
		observability.RecordCache(business, "hit")
		return v, nil
	} else if !errors.Is(err, redis.Nil) {
		return 0, err
	}
	observability.RecordCache(business, "miss")

	// 未命中，尝试加锁防击穿；加锁失败则短暂等待后双检，再不行直接回源。
	locked := false
	for i := 0; i < lockMaxRetry; i++ {
		if acquireLock(ctx, key) {
			locked = true
			break
		}
		time.Sleep(lockRetryGap)
		if v, err := Rdb.Get(ctx, key).Int64(); err == nil {
			return v, nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if v, err := Rdb.Get(ctx, key).Int64(); err == nil {
			return v, nil
		}
	}

	val, err := load(ctx)
	if err != nil {
		return 0, err
	}
	if err := Rdb.Set(ctx, key, val, JitterTTL(ttl)).Err(); err != nil {
		zap.L().Warn("redis set int cache failed", zap.String("key", key), zap.Error(err))
	}
	return val, nil
}

// LoadUintSet 通用集合 cache-aside。空结果不缓存（由上层 Bloom 防穿透）。
func LoadUintSet(ctx context.Context, key, business string, ttl time.Duration, load func(context.Context) ([]uint, error)) ([]uint, error) {
	if vals, err := Rdb.SMembers(ctx, key).Result(); err == nil && len(vals) > 0 {
		observability.RecordCache(business, "hit")
		return ParseUintMembers(vals), nil
	} else if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	observability.RecordCache(business, "miss")

	locked := false
	for i := 0; i < lockMaxRetry; i++ {
		if acquireLock(ctx, key) {
			locked = true
			break
		}
		time.Sleep(lockRetryGap)
		if ms, err := Rdb.SMembers(ctx, key).Result(); err == nil && len(ms) > 0 {
			return ParseUintMembers(ms), nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if ms, err := Rdb.SMembers(ctx, key).Result(); err == nil && len(ms) > 0 {
			return ParseUintMembers(ms), nil
		}
	}

	values, err := load(ctx)
	if err != nil {
		return nil, err
	}
	if len(values) > 0 {
		members := make([]interface{}, 0, len(values))
		for _, v := range values {
			members = append(members, v)
		}
		pipe := Rdb.Pipeline()
		pipe.SAdd(ctx, key, members...)
		pipe.Expire(ctx, key, JitterTTL(ttl))
		if _, err := pipe.Exec(ctx); err != nil {
			zap.L().Warn("redis set cache failed", zap.String("key", key), zap.Error(err))
		}
	}
	return values, nil
}

// SetContains 确保集合已加载后判断成员关系。
func SetContains(ctx context.Context, key, business string, ttl time.Duration, target uint, load func(context.Context) ([]uint, error)) (bool, error) {
	if _, err := LoadUintSet(ctx, key, business, ttl, load); err != nil {
		return false, err
	}
	return Rdb.SIsMember(ctx, key, target).Result()
}

func ParseUintMembers(members []string) []uint {
	out := make([]uint, 0, len(members))
	for _, m := range members {
		if n, err := strconv.ParseUint(m, 10, 64); err == nil {
			out = append(out, uint(n))
		}
	}
	return out
}
