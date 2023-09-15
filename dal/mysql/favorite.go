package mysql

import (
	"douyin/dal/model"
	"gorm.io/gorm"
)

// GetUserFavoriteCount 从数据库中根据id用户喜欢数
func GetUserFavoriteCount(id uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Favorite{}).Where("user_id = ?", id).Count(&cnt).Error
	return cnt, err
}

func GetVideoFavoriteCountByVideoId(id uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Favorite{}).Where("video_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// AddUserFavorite 添加喜欢关系
func AddUserFavorite(userId, videoId uint) bool {
	favorite := model.Favorite{UserId: userId, VideoId: videoId}
	result := DB.Model(&model.Favorite{}).Create(&favorite)
	return result.RowsAffected != 0
}

// BatchCreateUserFavorite 批量添加喜欢关系
func BatchCreateUserFavorite(favorites []model.Favorite) bool {
	result := DB.Model(&model.Favorite{}).Create(&favorites)
	return result.RowsAffected != 0
}

// BatchDeleteUserFavorite 批量删除喜欢关系
func BatchDeleteUserFavorite(favorites []model.Favorite) bool {
	result := DB.Model(&model.Favorite{}).Delete(&favorites)
	return result.RowsAffected != 0
}

// DeleteUserFavorite 删除喜欢关系
func DeleteUserFavorite(userId, videoId uint) error {
	favorite := model.Favorite{UserId: userId, VideoId: videoId}
	result := DB.Delete(&model.Favorite{}, favorite)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return result.Error
	}
	return nil
}

func IsFavorite(userId, videoId uint) bool {
	var count int64
	DB.Model(&model.Favorite{}).
		Where("user_id = ? AND video_id = ?", userId, videoId).
		Count(&count)
	return count != 0
}

func FindFavoriteByVideoId(userId, videoId uint) (uint, bool) {
	var id uint
	found := DB.Model(&model.Favorite{}).
		Select("id").
		Where("user_id = ? AND video_id = ?", userId, videoId).
		First(&id).
		RowsAffected != 0
	return id, found
}

// GetFavoritesById 从数据库中获取点赞列表
func GetFavoritesById(id uint) []uint {
	var videoList []uint
	DB.Model(&model.Favorite{}).
		Limit(30).
		Select("video_id").
		Where("user_id = ?", id).
		Order("id desc").
		Find(&videoList)
	return videoList
}

//
//// GetFavoritedUserCount 根据用户id，返回该用户的点赞的视频数（外部使用）
//func GetFavoritedUserCount(userId int64) (int, error) {
//	db := daoMysql.DB
//	rdb := daoRedis.UserFavoriteRDB
//	_, num, err := getFavoritesById(db, rdb, userId, idTypeUser)
//	return num, err
//}
