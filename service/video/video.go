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
	"douyin/dal/mysql"
	"douyin/kitex_gen/user"
	"douyin/kitex_gen/video"
	"douyin/mw/kafka"
	"douyin/mw/redis"
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

	ids, err := redis.GetFeedVideoIDs(ctx, latest, feedPageSize)
	if err != nil {
		observability.Logger(ctx).Error("get feed ids failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	// feed 冷启动：从 MySQL 回填后重试一次。
	if len(ids) == 0 {
		if berr := redis.BackfillFeed(ctx); berr == nil {
			ids, _ = redis.GetFeedVideoIDs(ctx, latest, feedPageSize)
		}
	}

	actorId := uint(req.UserId)
	var nextTime int64
	list := make([]*video.Video, 0, len(ids))
	for _, id := range ids {
		mv, ok := mysql.FindVideoByVideoId(id)
		if !ok {
			continue
		}
		list = append(list, v.buildVideo(ctx, &mv, actorId))
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

	if err := kafka.VideoMQInstance.Produce(&kafka.VideoMessage{
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
	videos, _ := mysql.FindVideosByAuthorId(uint(req.ToUserId))
	list := make([]*video.Video, 0, len(videos))
	actorId := uint(req.FromUserId)
	for i := range videos {
		list = append(list, v.buildVideo(ctx, &videos[i], actorId))
	}
	return &video.PublishVideoListResponse{StatusCode: common.CodeSuccess, VideoList: list}, nil
}

func (v *VideoAppServiceImpl) GetWorkCount(ctx context.Context, userId int64) (int32, error) {
	// Bloom 判否：作者无作品直接返回 0。
	if !common.TestWorkCountBloom(strconv.FormatInt(userId, 10)) {
		return 0, nil
	}
	return redis.GetWorkCount(ctx, uint(userId))
}

// buildVideo 组装视频详情：点赞/评论计数为进程内调用，作者信息为唯一跨域 RPC。
func (v *VideoAppServiceImpl) buildVideo(ctx context.Context, mv *model.Video, actorId uint) *video.Video {
	vv := &video.Video{
		Id:       int64(mv.ID),
		Title:    mv.Title,
		PlayUrl:  common.ResolveURL(mv.VideoUrl),
		CoverUrl: common.ResolveURL(mv.CoverUrl),
	}
	if fc, err := redis.GetVideoFavoriteCount(ctx, mv.ID); err == nil {
		vv.FavoriteCount = fc
	}
	if cc, err := redis.GetCommentCount(ctx, mv.ID); err == nil {
		vv.CommentCount = cc
	}
	if actorId != 0 {
		if fav, err := v.isFavorite(ctx, actorId, mv.ID); err == nil {
			vv.IsFavorite = fav
		}
	}
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
