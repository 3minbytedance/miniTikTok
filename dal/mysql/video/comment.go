package video

import (
	"douyin/dal/model"
	"go.uber.org/zap"
)

func AddComment(comment *model.Comment) (uint, error) {
	result := DB.Model(model.Comment{}).Create(comment)
	// 判断是否创建成功
	if result.Error != nil {
		zap.L().Error("创建 Comment 失败:", zap.Error(result.Error))
		return 0, result.Error
	} else {
		return comment.ID, nil
	}
}

func FindCommentsByVideoId(videoId uint) ([]model.Comment, error) {
	comments := make([]model.Comment, 0)
	err := DB.Where("video_id = ?", videoId).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func FindCommentById(commentId uint) (model.Comment, error) {
	comment := model.Comment{}
	err := DB.First(&comment, commentId).Error
	return comment, err
}

func DeleteCommentById(commentId uint) error {
	return DB.Delete(&model.Comment{}, commentId).Error
}

func GetCommentCnt(videoId uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Comment{}).Where("video_id = ?", videoId).Count(&cnt).Error
	// 返回评论数和是否查询成功
	return cnt, err
}

// IsCommentBelongsToUser 判断评论是否属于该用户（删除评论前的权限校验）。
func IsCommentBelongsToUser(commentId, userId int64) (bool, error) {
	var cnt int64
	err := DB.Model(&model.Comment{}).Where("id = ? and user_id = ?", commentId, userId).Count(&cnt).Error
	return cnt != 0, err
}
