package favorite

import (
	"context"
	"net/http"
	"strconv"

	"douyin/common"
	"douyin/kitex_gen/favorite"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

func Action(ctx context.Context, c *app.RequestContext) {
	userId, err := common.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, &favorite.FavoriteActionResponse{
			StatusCode: common.CodeInvalidToken,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidToken),
		})
		return
	}
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, &favorite.FavoriteActionResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	actionType, err := strconv.Atoi(c.Query("action_type"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &favorite.FavoriteActionResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Video.FavoriteAction(ctx, &favorite.FavoriteActionRequest{
		UserId:     int64(userId),
		VideoId:    videoId,
		ActionType: int32(actionType),
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &favorite.FavoriteActionResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func List(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, &favorite.FavoriteListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Video.GetFavoriteList(ctx, &favorite.FavoriteListRequest{
		ActionId: int64(actorId),
		UserId:   userId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &favorite.FavoriteListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
