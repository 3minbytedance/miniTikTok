namespace go video

include "user.thrift"

struct Video {
    1: i64 id, // 视频唯一标识
    2: user.User author, // 视频作者信息
    3: string play_url, // 视频播放地址
    4: string cover_url, // 视频封面地址
    5: i32 favorite_count, // 视频的点赞总数
    6: i32 comment_count, // 视频的评论总数
    7: bool is_favorite, // true-已点赞，false-未点赞
    8: string title, // 视频标题
}

struct VideoFeedRequest {
    1: optional string latest_time, // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
    2: i64 user_id, // 可选参数，登录用户设置
}

struct VideoFeedResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<Video> video_list, // 视频列表
    4: i64 next_time, // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
}

struct PublishVideoRequest {
    1: i64 user_id, // 用户id
    2: binary data, // 视频数据
    3: string title, // 视频标题
}

struct PublishVideoResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
}

struct PublishVideoListRequest {
    1: i64 from_user_id,
    2: i64 to_user_id,
}

struct PublishVideoListResponse {
    1: i32 status_code,
    2: string status_msg,
    3: list<Video> video_list,
}

struct GetWorkCountRequest{
    1: i64 user_id,
}

struct GetWorkCountResponse{
    1: i32 work_count, //作品数数量
}
