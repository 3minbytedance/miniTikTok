package redis

import (
	"context"
	"strconv"

	"douyin/dal/mysql"

	"github.com/redis/go-redis/v9"
)

// ===== 视频 Feed（ZSet，score=发布时间，由 kafka 消费者在落库后维护）=====

// AddVideoToFeed 新视频落库后加入 feed ZSet。
func AddVideoToFeed(ctx context.Context, videoId uint, createdAt int64) error {
	return Rdb.ZAdd(ctx, feedKey, redis.Z{Score: float64(createdAt), Member: videoId}).Err()
}

// GetFeedVideoIDs 取发布时间早于 latestTime 的视频 id（倒序），用于视频流。
// latestTime<=0 时取当前时间。
func GetFeedVideoIDs(ctx context.Context, latestTime int64, limit int64) ([]uint, error) {
	max := float64(latestTime)
	if latestTime <= 0 {
		max = float64(currentUnix())
	}
	members, err := Rdb.ZRevRangeByScore(ctx, feedKey, &redis.ZRangeBy{
		Max:   strconv.FormatFloat(max, 'f', -1, 64),
		Min:   "-inf",
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}
	return parseUintMembers(members), nil
}

// BackfillFeed feed 缓存为空时从 MySQL 全量回填（冷启动兜底）。
func BackfillFeed(ctx context.Context) error {
	videos := mysql.GetAllVideos(strconv.FormatInt(currentUnix(), 10))
	if len(videos) == 0 {
		return nil
	}
	zs := make([]redis.Z, 0, len(videos))
	for _, v := range videos {
		zs = append(zs, redis.Z{Score: float64(v.CreatedAt), Member: v.ID})
	}
	return Rdb.ZAdd(ctx, feedKey, zs...).Err()
}

// ===== 作品数 =====

func GetWorkCount(ctx context.Context, uid uint) (int32, error) {
	n, err := loadInt(ctx, workCountK(uid), "video_work_count", RdbExpireTime, func(ctx context.Context) (int64, error) {
		return mysql.FindWorkCountsByAuthorId(uid), nil
	})
	return int32(n), err
}

// InvalidateWorkCount 新视频发布后删除作品数缓存。
func InvalidateWorkCount(ctx context.Context, uid uint) {
	invalidate(ctx, workCountK(uid))
}

// ===== 评论数（列表不缓存，直接查库）=====

func GetCommentCount(ctx context.Context, videoId uint) (int32, error) {
	n, err := loadInt(ctx, commentCountK(videoId), "comment_count", RdbExpireTime, func(ctx context.Context) (int64, error) {
		return mysql.GetCommentCnt(videoId)
	})
	return int32(n), err
}

// InvalidateCommentCount 发/删评论后删除评论数缓存。
func InvalidateCommentCount(ctx context.Context, videoId uint) {
	invalidate(ctx, commentCountK(videoId))
}

// ===== 点赞 =====

// GetVideoFavoriteCount 视频点赞数。
func GetVideoFavoriteCount(ctx context.Context, videoId uint) (int32, error) {
	n, err := loadInt(ctx, videoFavK(videoId), "favorite_video_count", shortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetVideoFavoriteCountByVideoId(videoId)
	})
	return int32(n), err
}

// GetUserFavoriteCount 用户点赞数。
func GetUserFavoriteCount(ctx context.Context, uid uint) (int32, error) {
	n, err := loadInt(ctx, userFavCntK(uid), "favorite_user_count", shortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetUserFavoriteCount(uid)
	})
	return int32(n), err
}

// GetUserTotalFavoritedCount 作者获赞总数。
func GetUserTotalFavoritedCount(ctx context.Context, uid uint) (int32, error) {
	n, err := loadInt(ctx, userTotalFavK(uid), "favorite_total_count", shortTTL(), func(ctx context.Context) (int64, error) {
		return mysql.GetUserTotalFavoritedCount(uid)
	})
	return int32(n), err
}

// GetUserFavoriteSet 用户点赞过的视频 id 集合。
func GetUserFavoriteSet(ctx context.Context, uid uint) ([]uint, error) {
	return loadUintSet(ctx, userFavSetK(uid), "favorite_set", RdbExpireTime, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFavoritesById(uid), nil
	})
}

// IsUserFavorite 判断用户是否点赞某视频（Bloom 由上层先判否）。
func IsUserFavorite(ctx context.Context, uid, videoId uint) (bool, error) {
	return setContains(ctx, userFavSetK(uid), "favorite_set", RdbExpireTime, videoId, func(ctx context.Context) ([]uint, error) {
		return mysql.GetFavoritesById(uid), nil
	})
}

// InvalidateFavorite 点赞/取消写库成功后删除相关计数与用户点赞集合。
// uid=操作者，videoId=视频，authorId=视频作者。
func InvalidateFavorite(ctx context.Context, uid, videoId, authorId uint) {
	invalidate(ctx,
		videoFavK(videoId),
		userFavCntK(uid),
		userTotalFavK(authorId),
		userFavSetK(uid),
	)
}
