package message

import (
	"context"
	"net/http"
	"strconv"

	"douyin/common"
	"douyin/kitex_gen/message"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

func Action(ctx context.Context, c *app.RequestContext) {
	fromUserId, err := common.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusOK, &message.MessageActionResponse{
			StatusCode: common.CodeInvalidToken,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidToken),
		})
		return
	}
	toUserId, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &message.MessageActionResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Message.MessageAction(ctx, &message.MessageActionRequest{
		FromUserId: int64(fromUserId),
		ToUserId:   toUserId,
		ActionType: 1,
		Content:    c.Query("content"),
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &message.MessageActionResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func Chat(ctx context.Context, c *app.RequestContext) {
	fromUserId, err := common.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusOK, &message.MessageChatResponse{
			StatusCode: common.CodeInvalidToken,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidToken),
		})
		return
	}
	toUserId, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &message.MessageChatResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	preMsgTime, err := strconv.ParseInt(c.Query("pre_msg_time"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &message.MessageChatResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Message.MessageChat(ctx, &message.MessageChatRequest{
		FromUserId: int64(fromUserId),
		ToUserId:   toUserId,
		PreMsgTime: preMsgTime,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &message.MessageChatResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
