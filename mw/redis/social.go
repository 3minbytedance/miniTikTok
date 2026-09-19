package redis

import (
	"context"
	"errors"

	"douyin/dal/mysql"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ===== 用户名（cache-aside，注册后写入）=====

// GetUserName 读取用户名缓存，未命中回源 MySQL。
func GetUserName(ctx context.Context, uid uint) (string, error) {
	key := userNameK(uid)
	if name, err := Rdb.Get(ctx, key).Result(); err == nil {
		return name, nil
	} else if !errors.Is(err, redis.Nil) {
		return "", err
	}

	user, _, err := mysql.FindUserByUserID(uid)
	if err != nil {
		return "", err
	}
	if err := Rdb.Set(ctx, key, user.Name, jitterTTL(RdbExpireTime)).Err(); err != nil {
		return user.Name, err
	}
	return user.Name, nil
}

// SetUserName 注册/资料变更后写入用户名缓存。
func SetUserName(ctx context.Context, uid uint, name string) {
	if err := Rdb.Set(ctx, userNameK(uid), name, jitterTTL(RdbExpireTime)).Err(); err != nil {
		zap.L().Warn("set username cache failed", zap.Uint64("uid", uint64(uid)), zap.Error(err))
	}
}

// ===== 关注/粉丝集合（cache-aside）=====

// GetFollowSet 获取用户关注的用户 id 集合。
func GetFollowSet(ctx context.Context, uid uint) ([]uint, error) {
	return loadUintSet(ctx, followSetK(uid), "relation_follow", RdbExpireTime, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFollowList(uid)
	})
}

// GetFollowerSet 获取用户粉丝 id 集合。
func GetFollowerSet(ctx context.Context, uid uint) ([]uint, error) {
	return loadUintSet(ctx, followerSetK(uid), "relation_follower", RdbExpireTime, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFollowerList(uid)
	})
}

// IsFollowing 判断 uid 是否关注了 target。
func IsFollowing(ctx context.Context, uid, target uint) (bool, error) {
	return setContains(ctx, followSetK(uid), "relation_follow", RdbExpireTime, target, func(ctx context.Context) ([]uint, error) {
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
	n, err := loadInt(ctx, followCountK(uid), "relation_follow_count", shortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetFollowCnt(uid)
	})
	return int32(n), err
}

func GetFollowerCount(ctx context.Context, uid uint) (int32, error) {
	n, err := loadInt(ctx, followerCntK(uid), "relation_follower_count", shortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetFollowerCnt(uid)
	})
	return int32(n), err
}

// InvalidateRelation 关注/取关写库成功后，删除双方集合与计数缓存。
func InvalidateRelation(ctx context.Context, uid, other uint) {
	invalidate(ctx,
		followSetK(uid), followerSetK(uid),
		followSetK(other), followerSetK(other),
		followCountK(uid), followerCntK(uid),
		followCountK(other), followerCntK(other),
	)
}
