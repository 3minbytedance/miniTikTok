package mysql

import (
	"douyin/dal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AddComment(comment *model.Comment) (uint, error) {
	result := VideoDB.Model(model.Comment{}).Create(comment)
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
	result := VideoDB.Where("video_id = ?", videoId).Order("created_at desc").Find(&comments)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	return comments, nil
}

func FindCommentById(commentId uint) (model.Comment, error) {
	comment := model.Comment{}
	result := VideoDB.Find(&comment, commentId)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return comment, result.Error
	}
	return comment, nil
}

func DeleteCommentById(commentId uint) error {
	result := VideoDB.Delete(&model.Comment{}, commentId)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return result.Error
	}
	return nil
}

func GetCommentCnt(videoId uint) (int64, error) {
	var cnt int64
	err := VideoDB.Model(&model.Comment{}).Where("video_id = ?", videoId).Count(&cnt).Error
	// 返回评论数和是否查询成功
	return cnt, err
}

func IsCommentBelongsToUser(commentId *int64, userId int64) (bool, error) {
	var cnt int64
	err := VideoDB.Model(&model.Comment{}).Where("id = ? and user_id = ?", commentId, userId).Count(&cnt).Error
	return cnt != 0, err
}
