package main

import (
	"context"
	"strconv"

	videodao "douyin/dal/mysql/video"
	"douyin/mw/redis"

	goredis "github.com/redis/go-redis/v9"
)

// video 域 redis key：feed、作品数、评论数、点赞。
const (
	feedKey         = "videos" // videos feed ZSet
	workCountKey    = "workcnt"
	commentCountKey = "vcmtcnt"
	videoFavCount   = "vfav"    // 视频点赞数
	userFavList     = "ufavz"   // 用户点赞列表 ZSet，score=点赞时间
	userFavCount    = "ufavcnt" // 用户点赞数
	userTotalFav    = "utfav"   // 作者获赞总数
)

func workCountK(uid uint) string    { return redis.BuildKey(workCountKey, uid) }
func commentCountK(vid uint) string { return redis.BuildKey(commentCountKey, vid) }
func videoFavK(vid uint) string     { return redis.BuildKey(videoFavCount, vid) }
func userFavListK(uid uint) string  { return redis.BuildKey(userFavList, uid) }
func userFavCntK(uid uint) string   { return redis.BuildKey(userFavCount, uid) }
func userTotalFavK(uid uint) string { return redis.BuildKey(userTotalFav, uid) }

// ===== 视频 Feed（ZSet，score=发布时间，由 kafka 消费者在落库后维护）=====

// AddVideoToFeed 新视频落库后加入 feed ZSet。
func AddVideoToFeed(ctx context.Context, videoId uint, createdAt int64) error {
	return redis.Rdb.ZAdd(ctx, feedKey, goredis.Z{Score: float64(createdAt), Member: videoId}).Err()
}

// GetFeedVideoIDs 取发布时间早于 latestTime 的视频 id（倒序），用于视频流。
// latestTime<=0 时取当前时间。
func GetFeedVideoIDs(ctx context.Context, latestTime int64, limit int64) ([]uint, error) {
	maxScore := float64(latestTime)
	if latestTime <= 0 {
		maxScore = float64(redis.CurrentUnix())
	}
	members, err := redis.Rdb.ZRevRangeByScore(ctx, feedKey, &goredis.ZRangeBy{
		Max:   strconv.FormatFloat(maxScore, 'f', -1, 64),
		Min:   "-inf",
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}
	return redis.ParseUintMembers(members), nil
}

// BackfillFeed feed 缓存为空时从 MySQL 全量回填（冷启动兜底）。
func BackfillFeed(ctx context.Context) error {
	videos := videodao.GetAllVideos(strconv.FormatInt(redis.CurrentUnix(), 10))
	if len(videos) == 0 {
		return nil
	}
	zs := make([]goredis.Z, 0, len(videos))
	for _, v := range videos {
		zs = append(zs, goredis.Z{Score: float64(v.CreatedAt), Member: v.ID})
	}
	return redis.Rdb.ZAdd(ctx, feedKey, zs...).Err()
}

// ===== 作品数 =====

func GetWorkCount(ctx context.Context, uid uint) (int32, error) {
	n, err := redis.LoadInt(ctx, workCountK(uid), "video_work_count", redis.RdbExpireTime, func(ctx context.Context) (int64, error) {
		return videodao.FindWorkCountsByAuthorId(uid)
	})
	return int32(n), err
}

// InvalidateWorkCount 新视频发布后删除作品数缓存。
func InvalidateWorkCount(ctx context.Context, uid uint) {
	redis.Invalidate(ctx, workCountK(uid))
}

// ===== 评论数（列表不缓存，直接查库）=====

func GetCommentCount(ctx context.Context, videoId uint) (int32, error) {
	n, err := redis.LoadInt(ctx, commentCountK(videoId), "comment_count", redis.RdbExpireTime, func(ctx context.Context) (int64, error) {
		return videodao.GetCommentCnt(videoId)
	})
	return int32(n), err
}

// InvalidateCommentCount 发/删评论后删除评论数缓存。
func InvalidateCommentCount(ctx context.Context, videoId uint) {
	redis.Invalidate(ctx, commentCountK(videoId))
}

// ===== 点赞 =====

// GetVideoFavoriteCount 视频点赞数。
func GetVideoFavoriteCount(ctx context.Context, videoId uint) (int32, error) {
	n, err := redis.LoadInt(ctx, videoFavK(videoId), "favorite_video_count", redis.ShortTTL(), func(ctx context.Context) (int64, error) {
		return videodao.GetVideoFavoriteCountByVideoId(videoId)
	})
	return int32(n), err
}

// GetUserFavoriteCount 用户点赞数。
func GetUserFavoriteCount(ctx context.Context, uid uint) (int32, error) {
	n, err := redis.LoadInt(ctx, userFavCntK(uid), "favorite_user_count", redis.ShortTTL(), func(ctx context.Context) (int64, error) {
		return videodao.GetUserFavoriteCount(uid)
	})
	return int32(n), err
}

// GetUserTotalFavoritedCount 作者获赞总数。
func GetUserTotalFavoritedCount(ctx context.Context, uid uint) (int32, error) {
	n, err := redis.LoadInt(ctx, userTotalFavK(uid), "favorite_total_count", redis.ShortTTL(), func(ctx context.Context) (int64, error) {
		return videodao.GetUserTotalFavoritedCount(uid)
	})
	return int32(n), err
}

// GetUserFavoriteList 用户点赞列表：ZSet 按点赞时间倒序取前 limit 条，
// 只回源 limit 条记录，不再把用户全部点赞拉进内存。
func GetUserFavoriteList(ctx context.Context, uid uint, limit int) ([]uint, error) {
	return redis.LoadUintZSet(ctx, userFavListK(uid), "favorite_list", redis.RdbExpireTime, limit, func(ctx context.Context) ([]redis.ZItem, error) {
		favorites, err := videodao.GetFavoriteList(uid, limit)
		if err != nil {
			return nil, err
		}
		items := make([]redis.ZItem, 0, len(favorites))
		for _, f := range favorites {
			items = append(items, redis.ZItem{Score: f.CreatedAt, Member: f.VideoId})
		}
		return items, nil
	})
}

// InvalidateFavorite 点赞/取消写库成功后删除点赞列表与相关计数缓存。
// uid=操作者，videoId=视频，authorId=视频作者。
func InvalidateFavorite(ctx context.Context, uid, videoId, authorId uint) {
	redis.InvalidateSet(ctx, userFavListK(uid))
	redis.Invalidate(ctx,
		videoFavK(videoId),
		userFavCntK(uid),
		userTotalFavK(authorId),
	)
}
