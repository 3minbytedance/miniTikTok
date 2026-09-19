namespace go favorite

include "video.thrift"

struct FavoriteActionRequest {
    1: i64 user_id, // 用户id
    2: i64 video_id, // 视频id
    3: i32 action_type, // 1-点赞，2-取消点赞
}

struct FavoriteActionResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
}

struct FavoriteListRequest {
    1: i64 action_id, // 当前操作用户的用户id
    2: i64 user_id,   //列出user_id点赞的视频
}

struct FavoriteListResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<video.Video> video_list, // 用户点赞视频列表
}

struct IsUserFavoriteRequest{
    1: i64 user_id,  // 当前操作用户的用户id
    2: i64 video_id, // 视频id
}
