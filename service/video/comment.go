package main

import (
	"context"
	"strconv"
	"time"

	"douyin/common"
	"douyin/dal/model"
	"douyin/dal/mysql"
	"douyin/kitex_gen/comment"
	"douyin/kitex_gen/user"
	"douyin/mw/redis"
)

// CommentAction 发/删评论：先写 MySQL，成功后失效评论数缓存。
func (v *VideoAppServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (*comment.CommentActionResponse, error) {
	resp := &comment.CommentActionResponse{}
	uid, vid := uint(req.UserId), uint(req.VideoId)

	switch req.ActionType {
	case 1: // 发表评论
		if req.CommentText == nil || *req.CommentText == "" {
			resp.StatusCode = common.CodeInvalidParam
			resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
			return resp, nil
		}
		content := common.ReplaceWord(*req.CommentText)
		c := &model.Comment{VideoId: vid, UserId: uid, Content: content, CreatedAt: time.Now()}
		id, err := mysql.AddComment(c)
		if err != nil {
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
		redis.InvalidateCommentCount(ctx, vid)
		common.AddToCommentBloom(itoa64(int64(vid)))
		resp.StatusCode = common.CodeSuccess
		resp.Comment = v.buildComment(ctx, c, id)
		return resp, nil

	case 2: // 删除评论
		if req.CommentId == nil {
			resp.StatusCode = common.CodeInvalidParam
			resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
			return resp, nil
		}
		ok, err := mysql.IsCommentBelongsToUser(req.CommentId, req.UserId)
		if err != nil || !ok {
			resp.StatusCode = common.CodeInvalidCommentAction
			resp.StatusMsg = common.MapErrMsg(common.CodeInvalidCommentAction)
			return resp, nil
		}
		if err := mysql.DeleteCommentById(uint(*req.CommentId)); err != nil {
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
		redis.InvalidateCommentCount(ctx, vid)
		resp.StatusCode = common.CodeSuccess
		return resp, nil

	default:
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}
}

// GetCommentList 评论列表直接查 MySQL（实时性优先），逐条补用户信息。
func (v *VideoAppServiceImpl) GetCommentList(ctx context.Context, req *comment.CommentListRequest) (*comment.CommentListResponse, error) {
	comments, err := mysql.FindCommentsByVideoId(uint(req.VideoId))
	if err != nil {
		return &comment.CommentListResponse{
			StatusCode: common.CodeDBError,
			StatusMsg:  common.MapErrMsg(common.CodeDBError),
		}, nil
	}
	list := make([]*comment.Comment, 0, len(comments))
	for i := range comments {
		list = append(list, v.buildComment(ctx, &comments[i], comments[i].ID))
	}
	return &comment.CommentListResponse{StatusCode: common.CodeSuccess, CommentList: list}, nil
}

func (v *VideoAppServiceImpl) GetCommentCount(ctx context.Context, videoId int64) (int32, error) {
	if !common.TestCommentBloom(itoa64(videoId)) {
		return 0, nil
	}
	return redis.GetCommentCount(ctx, uint(videoId))
}

// buildComment 组装评论（用户信息跨域 RPC 查 social）。
func (v *VideoAppServiceImpl) buildComment(ctx context.Context, c *model.Comment, id uint) *comment.Comment {
	cc := &comment.Comment{
		Id:         int64(id),
		Content:    c.Content,
		CreateDate: c.CreatedAt.Format("01-02"),
	}
	if socialClient != nil {
		if ur, err := socialClient.GetUserInfoById(ctx, &user.UserInfoByIdRequest{
			ActorId: 0,
			UserId:  int64(c.UserId),
		}); err == nil && ur != nil {
			cc.User = ur.User
		}
	}
	return cc
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
