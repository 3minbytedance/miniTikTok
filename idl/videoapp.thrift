namespace go videoapp

include "user.thrift"
include "video.thrift"
include "comment.thrift"
include "favorite.thrift"

// VideoAppService 合并原 VideoService、CommentService、FavoriteService：
// 视频 feed/发布 + 评论 + 点赞及全部计数（域内计数查询为进程内调用）
service VideoAppService {
    // 视频域
    video.VideoFeedResponse VideoFeed(1: video.VideoFeedRequest Request),
    video.PublishVideoResponse PublishVideo(1: video.PublishVideoRequest Request),
    video.PublishVideoListResponse GetPublishVideoList(1: video.PublishVideoListRequest Request),
    i32 GetWorkCount(1: i64 user_id),

    // 评论域
    comment.CommentActionResponse CommentAction(1: comment.CommentActionRequest Request),
    comment.CommentListResponse GetCommentList(1: comment.CommentListRequest Request),
    i32 GetCommentCount(1: i64 video_id),

    // 点赞域
    favorite.FavoriteActionResponse FavoriteAction(1: favorite.FavoriteActionRequest Request),
    favorite.FavoriteListResponse GetFavoriteList(1: favorite.FavoriteListRequest Request),
    i32 GetVideoFavoriteCount(1: i64 video_id),
    i32 GetUserFavoriteCount(1: i64 user_id),
    i32 GetUserTotalFavoritedCount(1: i64 user_id),
    bool IsUserFavorite(1: favorite.IsUserFavoriteRequest Request),
}
