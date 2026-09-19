package comment

import (
	"context"
	"net/http"
	"strconv"

	"douyin/common"
	"douyin/kitex_gen/comment"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

func invalid(c *app.RequestContext) {
	c.JSON(http.StatusOK, &comment.CommentActionResponse{
		StatusCode: common.CodeInvalidParam,
		StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
	})
}

func Action(ctx context.Context, c *app.RequestContext) {
	userId, err := common.GetCurrentUserID(c)
	if err != nil {
		invalid(c)
		return
	}
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		invalid(c)
		return
	}
	actionType, err := strconv.Atoi(c.Query("action_type"))
	if err != nil {
		invalid(c)
		return
	}

	req := &comment.CommentActionRequest{
		UserId:     int64(userId),
		VideoId:    videoId,
		ActionType: int32(actionType),
	}
	switch actionType {
	case 1: // 发表评论
		text := c.Query("comment_text")
		if text == "" {
			invalid(c)
			return
		}
		req.CommentText = &text
	case 2: // 删除评论
		commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			invalid(c)
			return
		}
		req.CommentId = &commentId
	default:
		invalid(c)
		return
	}

	resp, err := rpc.Video.CommentAction(ctx, req)
	if err != nil || resp == nil {
		c.JSON(http.StatusInternalServerError, &comment.CommentActionResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func List(ctx context.Context, c *app.RequestContext) {
	userId, _ := common.GetCurrentUserID(c)
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &comment.CommentListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Video.GetCommentList(ctx, &comment.CommentListRequest{
		UserId:  int64(userId),
		VideoId: videoId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &comment.CommentListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
