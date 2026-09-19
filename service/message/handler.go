package main

import (
	"context"
	"time"

	"douyin/common"
	"douyin/dal/model"
	"douyin/dal/mongo"
	"douyin/kitex_gen/message"
	"douyin/kitex_gen/relation"
	"douyin/observability"
)

type MessageServiceImpl struct{}

// MessageChat 校验好友关系后，从 MongoDB 拉取历史消息。
func (m *MessageServiceImpl) MessageChat(ctx context.Context, req *message.MessageChatRequest) (*message.MessageChatResponse, error) {
	resp := &message.MessageChatResponse{}

	isFriend, err := socialClient.IsFriend(ctx, &relation.IsFriendRequest{
		ActorId: req.FromUserId,
		UserId:  req.ToUserId,
	})
	if err != nil {
		observability.Logger(ctx).Error("check friend failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	if !isFriend {
		resp.StatusCode = common.CodeNotFriend
		resp.StatusMsg = common.MapErrMsg(common.CodeNotFriend)
		return resp, nil
	}

	msgs, err := mongo.GetMessageList(uint(req.FromUserId), uint(req.ToUserId), req.PreMsgTime)
	if err != nil {
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}

	list := make([]*message.Message, 0, len(msgs))
	for _, msg := range msgs {
		list = append(list, &message.Message{
			Id:         msg.ID,
			FromUserId: int64(msg.FromUserId),
			ToUserId:   int64(msg.ToUserId),
			Content:    msg.Content,
			CreateTime: msg.CreateTime,
		})
	}
	resp.StatusCode = common.CodeSuccess
	resp.MessageList = list
	return resp, nil
}

// MessageAction 发送消息：校验好友后直接写入 MongoDB。
func (m *MessageServiceImpl) MessageAction(ctx context.Context, req *message.MessageActionRequest) (*message.MessageActionResponse, error) {
	resp := &message.MessageActionResponse{}

	isFriend, err := socialClient.IsFriend(ctx, &relation.IsFriendRequest{
		ActorId: req.FromUserId,
		UserId:  req.ToUserId,
	})
	if err != nil {
		observability.Logger(ctx).Error("check friend failed", errField(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	if !isFriend {
		resp.StatusCode = common.CodeNotFriend
		resp.StatusMsg = common.MapErrMsg(common.CodeNotFriend)
		return resp, nil
	}

	if req.Content == "" {
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}

	msg := &model.Message{
		FromUserId: uint(req.FromUserId),
		ToUserId:   uint(req.ToUserId),
		Content:    req.Content,
		CreateTime: time.Now().UnixNano() / 1e6, // 毫秒时间戳
	}
	if err := mongo.InsertMessage(msg); err != nil {
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}
	resp.StatusCode = common.CodeSuccess
	return resp, nil
}
