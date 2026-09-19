package main

import (
	"context"
	"errors"

	"douyin/dal/mysql"
	"douyin/mw/redis"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// social 域 redis key：用户名、关注/粉丝集合与计数。
const (
	userNameKey    = "uname"    // uname:{uid} 用户名
	followSetKey   = "follow"   // follow:{uid} 关注集合
	followerSetKey = "follower" // follower:{uid} 粉丝集合
	followCountKey = "followcnt"
	followerCntKey = "followercnt"
)

func userNameK(uid uint) string    { return redis.BuildKey(userNameKey, uid) }
func followSetK(uid uint) string   { return redis.BuildKey(followSetKey, uid) }
func followerSetK(uid uint) string { return redis.BuildKey(followerSetKey, uid) }
func followCountK(uid uint) string { return redis.BuildKey(followCountKey, uid) }
func followerCntK(uid uint) string { return redis.BuildKey(followerCntKey, uid) }

// ===== 用户名（cache-aside，注册后写入）=====

// GetUserName 读取用户名缓存，未命中回源 MySQL。
func GetUserName(ctx context.Context, uid uint) (string, error) {
	key := userNameK(uid)
	if name, err := redis.Rdb.Get(ctx, key).Result(); err == nil {
		return name, nil
	} else if !errors.Is(err, goredis.Nil) {
		return "", err
	}

	user, _, err := mysql.FindUserByUserID(uid)
	if err != nil {
		return "", err
	}
	if err := redis.Rdb.Set(ctx, key, user.Name, redis.JitterTTL(redis.RdbExpireTime)).Err(); err != nil {
		return user.Name, err
	}
	return user.Name, nil
}

// SetUserName 注册/资料变更后写入用户名缓存。
func SetUserName(ctx context.Context, uid uint, name string) {
	if err := redis.Rdb.Set(ctx, userNameK(uid), name, redis.JitterTTL(redis.RdbExpireTime)).Err(); err != nil {
		zap.L().Warn("set username cache failed", zap.Uint64("uid", uint64(uid)), zap.Error(err))
	}
}

// ===== 关注/粉丝集合（cache-aside）=====

// GetFollowSet 获取用户关注的用户 id 集合。
func GetFollowSet(ctx context.Context, uid uint) ([]uint, error) {
	return redis.LoadUintSet(ctx, followSetK(uid), "relation_follow", redis.RdbExpireTime, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFollowList(uid)
	})
}

// GetFollowerSet 获取用户粉丝 id 集合。
func GetFollowerSet(ctx context.Context, uid uint) ([]uint, error) {
	return redis.LoadUintSet(ctx, followerSetK(uid), "relation_follower", redis.RdbExpireTime, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFollowerList(uid)
	})
}

// IsFollowing 判断 uid 是否关注了 target。
func IsFollowing(ctx context.Context, uid, target uint) (bool, error) {
	return redis.SetContains(ctx, followSetK(uid), "relation_follow", redis.RdbExpireTime, target, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFollowList(uid)
	})
}

// IsFriend 判断双向关注（好友）。
func IsFriend(ctx context.Context, uid, target uint) (bool, error) {
	followed, err := IsFollowing(ctx, uid, target)
	if err != nil || !followed {
		return false, err
	}
	return IsFollowing(ctx, target, uid)
}

// ===== 关注/粉丝计数（cache-aside）=====

func GetFollowCount(ctx context.Context, uid uint) (int32, error) {
	n, err := redis.LoadInt(ctx, followCountK(uid), "relation_follow_count", redis.ShortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetFollowCnt(uid)
	})
	return int32(n), err
}

func GetFollowerCount(ctx context.Context, uid uint) (int32, error) {
	n, err := redis.LoadInt(ctx, followerCntK(uid), "relation_follower_count", redis.ShortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetFollowerCnt(uid)
	})
	return int32(n), err
}

// InvalidateRelation 关注/取关写库成功后，删除双方集合与计数缓存。
func InvalidateRelation(ctx context.Context, uid, other uint) {
	redis.Invalidate(ctx,
		followSetK(uid), followerSetK(uid),
		followSetK(other), followerSetK(other),
		followCountK(uid), followerCntK(uid),
		followCountK(other), followerCntK(other),
	)
}
