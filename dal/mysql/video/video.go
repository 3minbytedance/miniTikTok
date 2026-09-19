package video

import (
	"strconv"
	"time"

	"douyin/dal/model"
)

func FindVideoByVideoId(videoId uint) (model.Video, bool) {
	video := model.Video{}
	return video, DB.Where("id = ?", videoId).First(&video).RowsAffected != 0
}

// FindVideosByAuthorId 返回查询到的列表及是否出错
// 若未找到，返回空列表
func FindVideosByAuthorId(authorId uint) ([]model.Video, error) {
	videos := make([]model.Video, 0)
	err := DB.Where("author_id = ?", authorId).Find(&videos).Error
	return videos, err
}

// FindVideosByIDs 按 id 批量查询视频，返回 id→视频 映射，避免逐条查询的 N+1。
func FindVideosByIDs(ids []uint) (map[uint]model.Video, error) {
	result := make(map[uint]model.Video, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	videos := make([]model.Video, 0, len(ids))
	if err := DB.Where("id IN ?", ids).Find(&videos).Error; err != nil {
		return nil, err
	}
	for _, v := range videos {
		result[v.ID] = v
	}
	return result, nil
}

func FindWorkCountsByAuthorId(authorId uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Video{}).Where("author_id = ?", authorId).Count(&count).Error
	return count, err
}

// InsertVideo return 是否插入成功
func InsertVideo(video *model.Video) bool {
	result := DB.Model(model.Video{}).Create(video)
	return result.RowsAffected != 0
}

func GetLatestVideos(latestTime string) []model.Video {
	videos := make([]model.Video, 0, 30)

	DB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Limit(30).Find(&videos)
	if len(videos) == 0 {
		//如果视频都看完，重置时间戳
		latestTime = strconv.FormatInt(time.Now().Unix(), 10)
		DB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Limit(30).Find(&videos)
	}
	return videos
}

func GetAllVideos(latestTime string) []model.Video {
	videos := make([]model.Video, 0, 30)
	DB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Find(&videos)
	return videos
}

func GetAuthorIdByVideoId(videoId uint) (uint, bool) {
	var authorId uint
	find := DB.Model(&model.Video{}).
		Select("author_id").
		Where("id = ?", videoId).
		Find(&authorId).RowsAffected != 0
	return authorId, find
}
