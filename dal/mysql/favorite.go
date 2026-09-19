package mysql

import (
	"douyin/dal/model"

	"gorm.io/gorm"
)

// GetUserFavoriteCount 从数据库中根据id用户喜欢数
func GetUserFavoriteCount(id uint) (int64, error) {
	var cnt int64
	err := VideoDB.Model(&model.Favorite{}).Where("user_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// GetUserTotalFavoritedCount 统计某用户发布的所有视频获得的点赞总数。
func GetUserTotalFavoritedCount(authorId uint) (int64, error) {
	var cnt int64
	err := VideoDB.Model(&model.Favorite{}).
		Where("video_id IN (?)", VideoDB.Model(&model.Video{}).Select("id").Where("author_id = ?", authorId)).
		Count(&cnt).Error
	return cnt, err
}

func GetVideoFavoriteCountByVideoId(id uint) (int64, error) {
	var cnt int64
	err := VideoDB.Model(&model.Favorite{}).Where("video_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// AddUserFavorite 添加喜欢关系
func AddUserFavorite(userId, videoId uint) bool {
	favorite := model.Favorite{UserId: userId, VideoId: videoId}
	result := VideoDB.Model(&model.Favorite{}).Create(&favorite)
	return result.RowsAffected != 0
}

// BatchCreateUserFavorite 批量添加喜欢关系
func BatchCreateUserFavorite(favorites []model.Favorite) bool {
	result := VideoDB.Model(&model.Favorite{}).Create(&favorites)
	return result.RowsAffected != 0
}

// BatchDeleteUserFavorite 批量删除喜欢关系
func BatchDeleteUserFavorite(favorites []model.Favorite) bool {
	result := VideoDB.Model(&model.Favorite{}).Delete(&favorites)
	return result.RowsAffected != 0
}

// DeleteUserFavorite 删除喜欢关系
func DeleteUserFavorite(userId, videoId uint) error {
	favorite := model.Favorite{UserId: userId, VideoId: videoId}
	result := VideoDB.Delete(&model.Favorite{}, favorite)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return result.Error
	}
	return nil
}

func IsFavorite(userId, videoId uint) bool {
	var count int64
	VideoDB.Model(&model.Favorite{}).
		Where("user_id = ? AND video_id = ?", userId, videoId).
		Count(&count)
	return count != 0
}

func FindFavoriteByVideoId(userId, videoId uint) (uint, bool) {
	var id uint
	found := VideoDB.Model(&model.Favorite{}).
		Select("id").
		Where("user_id = ? AND video_id = ?", userId, videoId).
		First(&id).
		RowsAffected != 0
	return id, found
}

// GetFavoritesById 从数据库中获取点赞列表
func GetFavoritesById(id uint) []uint {
	var videoList []uint
	VideoDB.Model(&model.Favorite{}).
		Limit(30).
		Select("video_id").
		Where("user_id = ?", id).
		Order("id desc").
		Find(&videoList)
	return videoList
}
