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
// 重试窗口刻意压到 ~60ms：请求路径上等锁太久会拖高尾延迟，超时后直接回源即可。
const (
	lockTTL      = 3 * time.Second
	lockRetryGap = 20 * time.Millisecond
	lockMaxRetry = 3
)

// emptyMarkerTTL 集合回源为空时，空哨兵的短 TTL（防穿透，写操作会主动失效哨兵）。
const emptyMarkerTTL = 30 * time.Second

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

// InvalidateSet 删除集合 key 及其空哨兵。集合发生写操作（关注/点赞等）后必须调用，
// 否则回源为空时写下的短 TTL 哨兵会让写后读在哨兵 TTL 内仍看到空集合。
func InvalidateSet(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	all := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		all = append(all, k, emptyK(k))
	}
	Invalidate(ctx, all...)
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
			observability.RecordCache(business, "hit")
			return v, nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if v, err := Rdb.Get(ctx, key).Int64(); err == nil {
			observability.RecordCache(business, "hit")
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

// LoadString 通用字符串 cache-aside：未命中加锁、双检、回源、回填（带 TTL 抖动）。
func LoadString(ctx context.Context, key, business string, ttl time.Duration, load func(context.Context) (string, error)) (string, error) {
	if v, err := Rdb.Get(ctx, key).Result(); err == nil {
		observability.RecordCache(business, "hit")
		return v, nil
	} else if !errors.Is(err, redis.Nil) {
		return "", err
	}
	observability.RecordCache(business, "miss")

	// 未命中，尝试加锁防击穿；加锁失败则短暂等待后双检，再不行直接回源。
	locked := false
	for range lockMaxRetry {
		if acquireLock(ctx, key) {
			locked = true
			break
		}
		time.Sleep(lockRetryGap)
		if v, err := Rdb.Get(ctx, key).Result(); err == nil {
			observability.RecordCache(business, "hit")
			return v, nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if v, err := Rdb.Get(ctx, key).Result(); err == nil {
			observability.RecordCache(business, "hit")
			return v, nil
		}
	}

	val, err := load(ctx)
	if err != nil {
		return "", err
	}
	if err := Rdb.Set(ctx, key, val, JitterTTL(ttl)).Err(); err != nil {
		zap.L().Warn("redis set string cache failed", zap.String("key", key), zap.Error(err))
	}
	return val, nil
}

// LoadUintSet 通用集合 cache-aside：未命中加锁、双检、回源、回填（带 TTL 抖动）。
// 回源为空时写一个短 TTL 空哨兵（empty:{key}）防止缓存穿透；集合写入方须用 InvalidateSet 连同哨兵一起删除。
func LoadUintSet(ctx context.Context, key, business string, ttl time.Duration, load func(context.Context) ([]uint, error)) ([]uint, error) {
	if vals, hit, err := readUintSetCache(ctx, key); err != nil {
		return nil, err
	} else if hit {
		observability.RecordCache(business, "hit")
		return vals, nil
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
		if vals, hit, _ := readUintSetCache(ctx, key); hit {
			return vals, nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if vals, hit, err := readUintSetCache(ctx, key); err != nil {
			return nil, err
		} else if hit {
			return vals, nil
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
		pipe.Del(ctx, emptyK(key)) // 防御性清理可能残留的空哨兵
		if _, err := pipe.Exec(ctx); err != nil {
			zap.L().Warn("redis set cache failed", zap.String("key", key), zap.Error(err))
		}
	} else {
		// 空结果缓存短 TTL 哨兵，避免不存在/空集合的请求每次都回源。
		if err := Rdb.Set(ctx, emptyK(key), "1", JitterTTL(emptyMarkerTTL)).Err(); err != nil {
			zap.L().Warn("redis set empty marker failed", zap.String("key", key), zap.Error(err))
		}
	}
	return values, nil
}

// ZItem 有序集合成员：Score 决定排序，Member 为业务 id。
type ZItem struct {
	Score  int64
	Member uint
}

// LoadUintZSet 通用有序集合 cache-aside：未命中加锁、双检、回源、回填（带 TTL 抖动）。
// 只读取 score 倒序的前 limit 个成员，避免像集合那样把全量成员拉回应用。
// 回源为空时写空哨兵防穿透，写入方须用 InvalidateSet 连同哨兵一起删除。
func LoadUintZSet(ctx context.Context, key, business string, ttl time.Duration, limit int, load func(context.Context) ([]ZItem, error)) ([]uint, error) {
	if members, hit, err := readUintZSetCache(ctx, key, limit); err != nil {
		return nil, err
	} else if hit {
		observability.RecordCache(business, "hit")
		return members, nil
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
		if members, hit, _ := readUintZSetCache(ctx, key, limit); hit {
			return members, nil
		}
	}
	if locked {
		defer releaseLock(ctx, key)
		if members, hit, err := readUintZSetCache(ctx, key, limit); err != nil {
			return nil, err
		} else if hit {
			return members, nil
		}
	}

	items, err := load(ctx)
	if err != nil {
		return nil, err
	}
	members := make([]uint, 0, len(items))
	if len(items) > 0 {
		zs := make([]redis.Z, 0, len(items))
		for _, it := range items {
			zs = append(zs, redis.Z{Score: float64(it.Score), Member: it.Member})
			members = append(members, it.Member)
		}
		pipe := Rdb.Pipeline()
		pipe.ZAdd(ctx, key, zs...)
		pipe.Expire(ctx, key, JitterTTL(ttl))
		pipe.Del(ctx, emptyK(key)) // 防御性清理可能残留的空哨兵
		if _, err := pipe.Exec(ctx); err != nil {
			zap.L().Warn("redis set zset cache failed", zap.String("key", key), zap.Error(err))
		}
	} else {
		if err := Rdb.Set(ctx, emptyK(key), "1", JitterTTL(emptyMarkerTTL)).Err(); err != nil {
			zap.L().Warn("redis set empty marker failed", zap.String("key", key), zap.Error(err))
		}
	}
	return members, nil
}

// readUintZSetCache 读有序集合缓存：非空集合或空哨兵存在都视为命中，成员按 score 倒序返回。
func readUintZSetCache(ctx context.Context, key string, limit int) ([]uint, bool, error) {
	vals, err := Rdb.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, false, err
	}
	if len(vals) > 0 {
		return ParseUintMembers(vals), true, nil
	}
	ready, err := cacheReady(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if ready {
		return []uint{}, true, nil
	}
	return nil, false, nil
}

// readUintSetCache 读集合缓存：非空集合或空哨兵存在都视为命中。
func readUintSetCache(ctx context.Context, key string) ([]uint, bool, error) {
	if vals, err := Rdb.SMembers(ctx, key).Result(); err == nil && len(vals) > 0 {
		return ParseUintMembers(vals), true, nil
	} else if err != nil && !errors.Is(err, redis.Nil) {
		return nil, false, err
	}
	ready, err := cacheReady(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if ready {
		return []uint{}, true, nil
	}
	return nil, false, nil
}

// cacheReady 集合缓存是否已就绪：集合本体存在，或已写入"回源为空"的哨兵。
func cacheReady(ctx context.Context, key string) (bool, error) {
	n, err := Rdb.Exists(ctx, key, emptyK(key)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SetContains 判断集合成员关系。
// 缓存已就绪时直接 SIsMember（O(1)），避免为了判断单个成员而 SMEMBERS 拉回整个集合。
func SetContains(ctx context.Context, key, business string, ttl time.Duration, target uint, load func(context.Context) ([]uint, error)) (bool, error) {
	ready, err := cacheReady(ctx, key)
	if err != nil {
		return false, err
	}
	if ready {
		observability.RecordCache(business, "hit")
	} else if _, err := LoadUintSet(ctx, key, business, ttl, load); err != nil {
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
