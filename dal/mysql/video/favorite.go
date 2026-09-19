package video

import (
	"errors"

	"douyin/dal/model"

	"gorm.io/gorm"
)

// ErrFavoriteExists 重复点赞：并发下唯一索引 uk_user_video 冲突，与真实 DB 故障区分开。
var ErrFavoriteExists = errors.New("favorite already exists")

// ErrFavoriteNotExists 取消点赞时记录不存在：并发下已被其他请求取消，与真实 DB 故障区分开。
var ErrFavoriteNotExists = errors.New("favorite not exists")

// GetUserFavoriteCount 从数据库中根据id用户喜欢数
func GetUserFavoriteCount(id uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Favorite{}).Where("user_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// GetUserTotalFavoritedCount 统计某用户发布的所有视频获得的点赞总数。
func GetUserTotalFavoritedCount(authorId uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Favorite{}).
		Where("video_id IN (?)", DB.Model(&model.Video{}).Select("id").Where("author_id = ?", authorId)).
		Count(&cnt).Error
	return cnt, err
}

func GetVideoFavoriteCountByVideoId(id uint) (int64, error) {
	var cnt int64
	err := DB.Model(&model.Favorite{}).Where("video_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// AddUserFavorite 添加喜欢关系；重复点赞返回 ErrFavoriteExists，其余错误原样返回。
func AddUserFavorite(userId, videoId uint) error {
	err := DB.Model(&model.Favorite{}).Create(&model.Favorite{UserId: userId, VideoId: videoId}).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrFavoriteExists
	}
	return err
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

// DeleteUserFavorite 删除喜欢关系；记录不存在返回 ErrFavoriteNotExists，
// 让调用方直接以删除结果为准，无需先查后删。
func DeleteUserFavorite(userId, videoId uint) error {
	res := DB.Where("user_id = ? AND video_id = ?", userId, videoId).Delete(&model.Favorite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrFavoriteNotExists
	}
	return nil
}

// IsFavorite 判断点赞关系是否存在；只需存在性，用 Limit(1) 避免 COUNT 全量扫描。
func IsFavorite(userId, videoId uint) (bool, error) {
	var id uint
	res := DB.Model(&model.Favorite{}).
		Select("id").
		Where("user_id = ? AND video_id = ?", userId, videoId).
		Limit(1).
		Find(&id)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// BatchIsFavorite 批量判断用户对一组视频的点赞状态，返回已点赞的 video_id 集合，
// 把逐条 Limit(1) 查询压缩为一次 IN 查询。未点赞的 id 不出现在返回的 map 中。
func BatchIsFavorite(userId uint, videoIds []uint) (map[uint]bool, error) {
	favorites := make(map[uint]bool, len(videoIds))
	if len(videoIds) == 0 {
		return favorites, nil
	}
	var liked []uint
	err := DB.Model(&model.Favorite{}).
		Select("video_id").
		Where("user_id = ? AND video_id IN ?", userId, videoIds).
		Find(&liked).Error
	if err != nil {
		return nil, err
	}
	for _, id := range liked {
		favorites[id] = true
	}
	return favorites, nil
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

// GetFavoriteList 按点赞顺序倒序取用户点赞记录，供点赞列表缓存回源。
// 用 id 排序而非 created_at：自增 id 与点赞时间同序，且历史行 created_at=0 不影响排序。
func GetFavoriteList(userId uint, limit int) ([]model.Favorite, error) {
	list := make([]model.Favorite, 0, limit)
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}
