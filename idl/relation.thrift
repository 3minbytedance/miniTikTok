namespace go relation

include "user.thrift"

struct RelationActionRequest {
    1: i64 user_id, // 当前登录用户
    2: i64 to_user_id, // 对方用户id
    3: i32 action_type, // 1-关注 2-取消关注
}

struct RelationActionResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
}

struct FollowListRequest {
    1: i64 user_id, // 当前登录用户
    2: i64 to_user_id, // 对方用户id
}

struct FollowListResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<user.User> user_list, // 用户信息列表
}


struct FollowerListRequest {
    1: i64 user_id, // 当前登录用户
    2: i64 to_user_id, // 对方用户id
}

//粉丝列表
struct FollowerListResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<user.User> user_list, // 用户列表
}

struct FollowListCountResponse {
    1: i32 count, // 关注数
}

struct FollowerListCountRequest {
    1: i64 user_id, // 用户id
}


struct FriendListRequest {
    1: i64 user_id,     //当前登录用户id
    2: i64 to_user_id,  //对方用户id
}

struct FriendListResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<user.User> user_list, // 用户列表
}

struct IsFollowingRequest {
    1: i64 actor_id, //当前操作id
    2: i64 user_id,  //对方用户id
}

struct IsFriendRequest {
    1: i64 actor_id, //当前操作id
    2: i64 user_id,  //对方用户id
}
