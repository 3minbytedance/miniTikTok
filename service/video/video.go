package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"douyin/common"
	"douyin/constant/biz"
	"douyin/dal/model"
	videodao "douyin/dal/mysql/video"
	"douyin/kitex_gen/user"
	"douyin/kitex_gen/video"
	"douyin/observability"

	"github.com/google/uuid"
)

type VideoAppServiceImpl struct{}

const feedPageSize = 30

// VideoFeed 视频流：从 feed ZSet 取 id，逐条本地补点赞/评论计数，作者信息 RPC 查 social。
func (v *VideoAppServiceImpl) VideoFeed(ctx context.Context, req *video.VideoFeedRequest) (*video.VideoFeedResponse, error) {
	resp := &video.VideoFeedResponse{StatusCode: common.CodeSuccess}

	var latest int64
	if req.LatestTime != nil {
		if t, err := strconv.ParseInt(*req.LatestTime, 10, 64); err == nil {
			latest = t
		}
	}

	ids, err := GetFeedVideoIDs(ctx, latest, feedPageSize)
	if err != nil {
		observability.Logger(ctx).Error("get feed ids failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	// feed 冷启动：从 MySQL 回填后重试一次。
	if len(ids) == 0 {
		if berr := BackfillFeed(ctx); berr == nil {
			ids, _ = GetFeedVideoIDs(ctx, latest, feedPageSize)
		}
	}

	actorId := uint(req.UserId)
	// 批量查询视频，避免逐条 FindVideoByVideoId 的 N+1。
	videos, err := videodao.FindVideosByIDs(ids)
	if err != nil {
		observability.Logger(ctx).Error("batch find videos failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	favorites := prefetchFavorites(ctx, actorId, ids)
	var nextTime int64
	list := make([]*video.Video, 0, len(ids))
	for _, id := range ids {
		mv, ok := videos[id]
		if !ok {
			continue
		}
		list = append(list, v.buildVideo(ctx, &mv, actorId, favorites))
		nextTime = mv.CreatedAt
	}
	if nextTime == 0 {
		nextTime = time.Now().Unix()
	}
	resp.VideoList = list
	resp.NextTime = nextTime
	return resp, nil
}

// PublishVideo 保存上传字节到本地暂存文件，发送 kafka 消息异步完成持久化/截帧/落库。
func (v *VideoAppServiceImpl) PublishVideo(ctx context.Context, req *video.PublishVideoRequest) (*video.PublishVideoResponse, error) {
	resp := &video.PublishVideoResponse{}
	if len(req.Data) == 0 {
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}

	if err := common.CreateDirectoryIfNotExist(); err != nil {
		observability.Logger(ctx).Error("create dir failed", errField(err))
		resp.StatusCode = common.CodeUploadFileError
		resp.StatusMsg = common.MapErrMsg(common.CodeUploadFileError)
		return resp, nil
	}

	fileName := uuid.NewString() + ".mp4"
	localPath := filepath.Join(common.LocalFallbackDir(), fileName)
	if err := os.WriteFile(localPath, req.Data, biz.FileMode); err != nil {
		observability.Logger(ctx).Error("save video temp file failed", errField(err))
		resp.StatusCode = common.CodeUploadFileError
		resp.StatusMsg = common.MapErrMsg(common.CodeUploadFileError)
		return resp, nil
	}

	if err := VideoMQInstance.Produce(&VideoMessage{
		VideoPath:     localPath,
		VideoFileName: fileName,
		UserID:        uint(req.UserId),
		Title:         req.Title,
	}); err != nil {
		observability.Logger(ctx).Error("produce video message failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}

	resp.StatusCode = common.CodeSuccess
	return resp, nil
}

// GetPublishVideoList 用户作品列表。
func (v *VideoAppServiceImpl) GetPublishVideoList(ctx context.Context, req *video.PublishVideoListRequest) (*video.PublishVideoListResponse, error) {
	videos, err := videodao.FindVideosByAuthorId(uint(req.ToUserId))
	if err != nil {
		observability.Logger(ctx).Error("find videos by author failed", errField(err))
		return &video.PublishVideoListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		}, nil
	}
	list := make([]*video.Video, 0, len(videos))
	actorId := uint(req.FromUserId)
	ids := make([]uint, 0, len(videos))
	for i := range videos {
		ids = append(ids, videos[i].ID)
	}
	favorites := prefetchFavorites(ctx, actorId, ids)
	for i := range videos {
		list = append(list, v.buildVideo(ctx, &videos[i], actorId, favorites))
	}
	return &video.PublishVideoListResponse{StatusCode: common.CodeSuccess, VideoList: list}, nil
}

// GetWorkCount 作者作品数（0 值由计数缓存直接缓存，无需额外判否）。
func (v *VideoAppServiceImpl) GetWorkCount(ctx context.Context, userId int64) (int32, error) {
	return GetWorkCount(ctx, uint(userId))
}

// prefetchFavorites 批量预取当前用户对该批视频的点赞状态，把逐条判定压缩成一次 IN 查询。
// actorId 为 0（未登录）或查询失败时返回 nil，等价于全部未点赞。
func prefetchFavorites(ctx context.Context, actorId uint, videoIds []uint) map[uint]bool {
	if actorId == 0 || len(videoIds) == 0 {
		return nil
	}
	favorites, err := videodao.BatchIsFavorite(actorId, videoIds)
	if err != nil {
		observability.Logger(ctx).Error("batch query favorite failed", errField(err))
		return nil
	}
	return favorites
}

// buildVideo 组装视频详情：点赞/评论计数为进程内调用，作者信息为唯一跨域 RPC。
// favorites 由调用方按页预取，nil 表示全部未点赞。
func (v *VideoAppServiceImpl) buildVideo(ctx context.Context, mv *model.Video, actorId uint, favorites map[uint]bool) *video.Video {
	vv := &video.Video{
		Id:       int64(mv.ID),
		Title:    mv.Title,
		PlayUrl:  common.ResolveURL(mv.VideoUrl),
		CoverUrl: common.ResolveURL(mv.CoverUrl),
	}
	if fc, err := GetVideoFavoriteCount(ctx, mv.ID); err == nil {
		vv.FavoriteCount = fc
	}
	if cc, err := GetCommentCount(ctx, mv.ID); err == nil {
		vv.CommentCount = cc
	}
	vv.IsFavorite = favorites[mv.ID]
	if socialClient != nil {
		if ur, err := socialClient.GetUserInfoById(ctx, &user.UserInfoByIdRequest{
			ActorId: int64(actorId),
			UserId:  int64(mv.AuthorId),
		}); err == nil && ur != nil {
			vv.Author = ur.User
		}
	}
	return vv
}
