package relation

import (
	"context"
	"net/http"
	"strconv"

	"douyin/common"
	"douyin/kitex_gen/relation"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

func Action(ctx context.Context, c *app.RequestContext) {
	actionId, err := common.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusOK, &relation.RelationActionResponse{
			StatusCode: common.CodeInvalidToken,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidToken),
		})
		return
	}
	toUserId, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &relation.RelationActionResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	if int64(actionId) == toUserId {
		c.JSON(http.StatusOK, &relation.RelationActionResponse{
			StatusCode: common.CodeFollowMyself,
			StatusMsg:  common.MapErrMsg(common.CodeFollowMyself),
		})
		return
	}
	actionType, err := strconv.Atoi(c.Query("action_type"))
	if err != nil || (actionType != 1 && actionType != 2) {
		c.JSON(http.StatusOK, &relation.RelationActionResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.RelationAction(ctx, &relation.RelationActionRequest{
		UserId:     int64(actionId),
		ToUserId:   toUserId,
		ActionType: int32(actionType),
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &relation.RelationActionResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseTargetUser(c *app.RequestContext) (int64, bool) {
	id, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func FollowList(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	toUserId, ok := parseTargetUser(c)
	if !ok {
		c.JSON(http.StatusOK, &relation.FollowListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.GetFollowList(ctx, &relation.FollowListRequest{
		UserId:   int64(actorId),
		ToUserId: toUserId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusInternalServerError, &relation.FollowListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func FollowerList(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	toUserId, ok := parseTargetUser(c)
	if !ok {
		c.JSON(http.StatusOK, &relation.FollowerListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.GetFollowerList(ctx, &relation.FollowerListRequest{
		UserId:   int64(actorId),
		ToUserId: toUserId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusInternalServerError, &relation.FollowerListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func FriendList(ctx context.Context, c *app.RequestContext) {
	actorId, err := common.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusOK, &relation.FriendListResponse{
			StatusCode: common.CodeInvalidToken,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidToken),
		})
		return
	}
	toUserId, ok := parseTargetUser(c)
	if !ok {
		c.JSON(http.StatusOK, &relation.FriendListResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.GetFriendList(ctx, &relation.FriendListRequest{
		UserId:   int64(actorId),
		ToUserId: toUserId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusInternalServerError, &relation.FriendListResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
