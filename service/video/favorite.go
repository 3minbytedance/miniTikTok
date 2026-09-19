package main

import (
	"context"
	"errors"

	"douyin/common"
	videodao "douyin/dal/mysql/video"
	"douyin/kitex_gen/favorite"
	"douyin/kitex_gen/video"
	"douyin/observability"
)

// FavoriteAction 点赞/取消：以写库结果为准（唯一索引兜底重复、影响行数兜底未点赞），
// 省掉写前预检；成功后失效点赞列表与相关计数缓存。
func (v *VideoAppServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	resp := &favorite.FavoriteActionResponse{}
	uid, vid := uint(req.UserId), uint(req.VideoId)

	authorId, found := videodao.GetAuthorIdByVideoId(vid)
	if !found {
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}

	switch req.ActionType {
	case 1: // 点赞：直接插入，重复点赞由唯一索引 uk_user_video 兜底
		if err := videodao.AddUserFavorite(uid, vid); err != nil {
			if errors.Is(err, videodao.ErrFavoriteExists) { // 已经点过，含并发下另一个请求刚成功
				resp.StatusCode = common.CodeFavoriteRepeat
				resp.StatusMsg = common.MapErrMsg(common.CodeFavoriteRepeat)
				return resp, nil
			}
			observability.Logger(ctx).Error("add favorite failed", errField(err))
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
	case 2: // 取消点赞：删除结果即判定依据，无受影响行说明本就没点赞
		if err := videodao.DeleteUserFavorite(uid, vid); err != nil {
			if errors.Is(err, videodao.ErrFavoriteNotExists) {
				resp.StatusCode = common.CodeFavoriteRepeat
				resp.StatusMsg = common.MapErrMsg(common.CodeFavoriteRepeat)
				return resp, nil
			}
			observability.Logger(ctx).Error("delete favorite failed", errField(err))
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
	default:
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}

	InvalidateFavorite(ctx, uid, vid, authorId)
	resp.StatusCode = common.CodeSuccess
	return resp, nil
}

// favoriteListSize 点赞列表返回条数。
const favoriteListSize = 30

// GetFavoriteList 用户点赞视频列表（按点赞时间倒序）。
func (v *VideoAppServiceImpl) GetFavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (*favorite.FavoriteListResponse, error) {
	resp := &favorite.FavoriteListResponse{StatusCode: common.CodeSuccess}
	ids, err := GetUserFavoriteList(ctx, uint(req.UserId), favoriteListSize)
	if err != nil {
		observability.Logger(ctx).Error("get favorite list failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	actorId := uint(req.ActionId)
	// 批量查询视频，避免逐条 FindVideoByVideoId 的 N+1。
	videos, err := videodao.FindVideosByIDs(ids)
	if err != nil {
		observability.Logger(ctx).Error("batch find videos failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	favorites := prefetchFavorites(ctx, actorId, ids)
	list := make([]*video.Video, 0, len(ids))
	for _, id := range ids {
		mv, ok := videos[id]
		if !ok {
			continue
		}
		list = append(list, v.buildVideo(ctx, &mv, actorId, favorites))
	}
	resp.VideoList = list
	return resp, nil
}

func (v *VideoAppServiceImpl) GetVideoFavoriteCount(ctx context.Context, videoId int64) (int32, error) {
	return GetVideoFavoriteCount(ctx, uint(videoId))
}

func (v *VideoAppServiceImpl) GetUserFavoriteCount(ctx context.Context, userId int64) (int32, error) {
	return GetUserFavoriteCount(ctx, uint(userId))
}

func (v *VideoAppServiceImpl) GetUserTotalFavoritedCount(ctx context.Context, userId int64) (int32, error) {
	return GetUserTotalFavoritedCount(ctx, uint(userId))
}

func (v *VideoAppServiceImpl) IsUserFavorite(ctx context.Context, req *favorite.IsUserFavoriteRequest) (bool, error) {
	return videodao.IsFavorite(uint(req.UserId), uint(req.VideoId))
}
