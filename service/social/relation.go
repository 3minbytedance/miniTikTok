package main

import (
	"context"

	"douyin/common"
	socialdao "douyin/dal/mysql/social"
	"douyin/kitex_gen/relation"
	"douyin/kitex_gen/user"
)

// RelationAction 关注/取关：先写 MySQL，成功后删除双方关系缓存。
func (s *SocialServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (*relation.RelationActionResponse, error) {
	resp := &relation.RelationActionResponse{}
	uid, toUid := uint(req.UserId), uint(req.ToUserId)

	if uid == toUid {
		resp.StatusCode = common.CodeFollowMyself
		resp.StatusMsg = common.MapErrMsg(common.CodeFollowMyself)
		return resp, nil
	}

	switch req.ActionType {
	case 1: // 关注
		if socialdao.IsFollowing(uid, toUid) {
			resp.StatusCode = common.CodeFollowRepeat
			resp.StatusMsg = common.MapErrMsg(common.CodeFollowRepeat)
			return resp, nil
		}
		if err := socialdao.AddFollow(uid, toUid); err != nil {
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
	case 2: // 取关
		if !socialdao.IsFollowing(uid, toUid) {
			resp.StatusCode = common.CodeCancelFollowRepeat
			resp.StatusMsg = common.MapErrMsg(common.CodeCancelFollowRepeat)
			return resp, nil
		}
		if err := socialdao.DeleteFollowById(uid, toUid); err != nil {
			resp.StatusCode = common.CodeDBError
			resp.StatusMsg = common.MapErrMsg(common.CodeDBError)
			return resp, nil
		}
	default:
		resp.StatusCode = common.CodeInvalidParam
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidParam)
		return resp, nil
	}

	// 写库成功后失效缓存（cache-aside invalidate）。
	InvalidateRelation(ctx, uid, toUid)
	resp.StatusCode = common.CodeSuccess
	return resp, nil
}

func (s *SocialServiceImpl) GetFollowList(ctx context.Context, req *relation.FollowListRequest) (*relation.FollowListResponse, error) {
	ids, err := GetFollowSet(ctx, uint(req.ToUserId))
	if err != nil {
		return &relation.FollowListResponse{StatusCode: common.CodeDBError, StatusMsg: common.MapErrMsg(common.CodeDBError)}, nil
	}
	return &relation.FollowListResponse{
		StatusCode: common.CodeSuccess,
		UserList:   s.buildUserList(ctx, ids, uint(req.UserId)),
	}, nil
}

func (s *SocialServiceImpl) GetFollowerList(ctx context.Context, req *relation.FollowerListRequest) (*relation.FollowerListResponse, error) {
	ids, err := GetFollowerSet(ctx, uint(req.ToUserId))
	if err != nil {
		return &relation.FollowerListResponse{StatusCode: common.CodeDBError, StatusMsg: common.MapErrMsg(common.CodeDBError)}, nil
	}
	return &relation.FollowerListResponse{
		StatusCode: common.CodeSuccess,
		UserList:   s.buildUserList(ctx, ids, uint(req.UserId)),
	}, nil
}

// GetFriendList 好友 = 双向关注，取关注集合与粉丝集合的交集。
func (s *SocialServiceImpl) GetFriendList(ctx context.Context, req *relation.FriendListRequest) (*relation.FriendListResponse, error) {
	follows, err := GetFollowSet(ctx, uint(req.ToUserId))
	if err != nil {
		return &relation.FriendListResponse{StatusCode: common.CodeDBError, StatusMsg: common.MapErrMsg(common.CodeDBError)}, nil
	}
	friends := make([]uint, 0, len(follows))
	for _, id := range follows {
		if ok, _ := IsFollowing(ctx, id, uint(req.ToUserId)); ok {
			friends = append(friends, id)
		}
	}
	return &relation.FriendListResponse{
		StatusCode: common.CodeSuccess,
		UserList:   s.buildUserList(ctx, friends, uint(req.UserId)),
	}, nil
}

func (s *SocialServiceImpl) GetFollowListCount(ctx context.Context, userId int64) (int32, error) {
	c, err := GetFollowCount(ctx, uint(userId))
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (s *SocialServiceImpl) GetFollowerListCount(ctx context.Context, userId int64) (int32, error) {
	c, err := GetFollowerCount(ctx, uint(userId))
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (s *SocialServiceImpl) IsFollowing(ctx context.Context, req *relation.IsFollowingRequest) (bool, error) {
	return IsFollowing(ctx, uint(req.ActorId), uint(req.UserId))
}

func (s *SocialServiceImpl) IsFriend(ctx context.Context, req *relation.IsFriendRequest) (bool, error) {
	return IsFriend(ctx, uint(req.ActorId), uint(req.UserId))
}

// buildUserList 顺序组装用户信息（跨域统计走短 TTL 缓存，避免 goroutine fan-out）。
func (s *SocialServiceImpl) buildUserList(ctx context.Context, ids []uint, actorId uint) []*user.User {
	list := make([]*user.User, 0, len(ids))
	for _, id := range ids {
		u, err := s.buildUser(ctx, id, actorId)
		if err != nil {
			continue
		}
		list = append(list, u)
	}
	return list
}
