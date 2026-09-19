package video

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"douyin/common"
	"douyin/kitex_gen/video"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	maxFileSize = 50 * 1024 * 1024
	minFileSize = 1 * 1024 * 1024
)

func FeedList(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	latestTime := c.Query("latest_time")
	if latestTime == "" {
		latestTime = strconv.FormatInt(time.Now().Unix(), 10)
	} else if t, err := strconv.Atoi(latestTime); err != nil || t < 0 {
		c.JSON(http.StatusOK, &video.VideoFeedResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Video.VideoFeed(ctx, &video.VideoFeedRequest{
		UserId:     int64(actorId),
		LatestTime: &latestTime,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &video.VideoFeedResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPublishList(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &video.PublishVideoListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Video.GetPublishVideoList(ctx, &video.PublishVideoListRequest{
		FromUserId: int64(actorId),
		ToUserId:   userId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &video.PublishVideoListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func Publish(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	title := c.PostForm("title")
	file, err := c.FormFile("data")
	if err != nil || title == "" || file.Size == 0 {
		c.JSON(http.StatusOK, &video.PublishVideoResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidFileType(ext) {
		c.JSON(http.StatusOK, &video.PublishVideoResponse{
			StatusCode: common.CodeInvalidFileType,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidFileType),
		})
		return
	}
	if file.Size > maxFileSize || file.Size < minFileSize {
		c.JSON(http.StatusOK, &video.PublishVideoResponse{
			StatusCode: common.CodeInvalidFileSize,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidFileSize),
		})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, &video.PublishVideoResponse{
			StatusCode: common.CodeUploadFileError,
			StatusMsg:  common.MapErrMsg(common.CodeUploadFileError),
		})
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &video.PublishVideoResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	resp, err := rpc.Video.PublishVideo(ctx, &video.PublishVideoRequest{
		UserId: int64(actorId),
		Title:  title,
		Data:   data,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &video.PublishVideoResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func isValidFileType(ext string) bool {
	for _, e := range []string{".mp4", ".avi", ".mov"} {
		if ext == e {
			return true
		}
	}
	return false
}
