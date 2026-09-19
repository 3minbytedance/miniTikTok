namespace go comment

include "user.thrift"

struct Comment {
    1: i64 id, // 视频评论id
    2: user.User user, // 评论用户信息
    3: string content, // 评论内容
    4: string create_date, // 评论发布日期，格式 mm-dd
}

struct CommentActionRequest {
    1: i64 user_id, // 用户鉴权token
    2: i64 video_id, // 视频id
    3: i32 action_type, // 1-发布评论，2-删除评论
    4: optional string comment_text, // 用户填写的评论内容，在action_type=1的时候使用
    5: optional i64 comment_id, // 要删除的评论id，在action_type=2的时候使用
}

struct CommentActionResponse {
    1: i32 status_code,
    2: string status_msg,
    3: optional Comment comment, // 评论成功返回评论内容，不需要重新拉取整个列表
}

struct CommentListRequest {
    1: i64 user_id, // 用户id
    2: i64 video_id, // 视频id
}

struct CommentListResponse {
    1: i32 status_code, // 状态码，0-成功，其他值-失败
    2: string status_msg, // 返回状态描述
    3: list<Comment> comment_list, // 评论列表
}
