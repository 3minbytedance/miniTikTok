.3namespace go social

include "user.thrift"
include "relation.thrift"

// SocialService 合并原 UserService 与 RelationService：
// 用户注册/登录/信息 + 关注/粉丝/好友关系
service SocialService {
    // 用户域
    user.UserRegisterResponse Register(1: user.UserRegisterRequest Request),
    user.UserLoginResponse Login(1: user.UserLoginRequest Request),
    user.UserInfoByIdResponse GetUserInfoById(1: user.UserInfoByIdRequest Request),

    // 关系域
    relation.RelationActionResponse RelationAction(1: relation.RelationActionRequest Request),
    relation.FollowListResponse GetFollowList(1: relation.FollowListRequest Request),
    relation.FollowerListResponse GetFollowerList(1: relation.FollowerListRequest Request),
    relation.FriendListResponse GetFriendList(1: relation.FriendListRequest Request),
    i32 GetFollowListCount(1: i64 user_id),
    i32 GetFollowerListCount(1: i64 user_id),
    bool IsFollowing(1: relation.IsFollowingRequest Request),
    bool IsFriend(1: relation.IsFriendRequest Request),
}
