package mysql

import (
	"douyin/dal/model"
	"strconv"
	"time"
)

func FindVideoByVideoId(videoId uint) (model.Video, bool) {
	video := model.Video{}
	return video, VideoDB.Where("id = ?", videoId).First(&video).RowsAffected != 0
}

// FindVideosByAuthorId 返回查询到的列表及是否出错
// 若未找到，返回空列表
func FindVideosByAuthorId(authorId uint) ([]model.Video, bool) {
	videos := make([]model.Video, 0)
	return videos, VideoDB.Where("author_id = ?", authorId).Find(&videos).RowsAffected != 0
}

func FindWorkCountsByAuthorId(authorId uint) int64 {
	var count int64
	VideoDB.Model(&model.Video{}).Where("author_id = ?", authorId).Count(&count)
	return count
}

// InsertVideo return 是否插入成功
func InsertVideo(video *model.Video) bool {
	result := VideoDB.Model(model.Video{}).Create(video)
	return result.RowsAffected != 0
}

func GetLatestVideos(latestTime string) []model.Video {
	videos := make([]model.Video, 0, 30)

	VideoDB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Limit(30).Find(&videos)
	if len(videos) == 0 {
		//如果视频都看完，重置时间戳
		latestTime = strconv.FormatInt(time.Now().Unix(), 10)
		VideoDB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Limit(30).Find(&videos)
	}
	return videos
}

func GetAllVideos(latestTime string) []model.Video {
	videos := make([]model.Video, 0, 30)
	VideoDB.Model(&model.Video{}).Where("created_at < ?", latestTime).Order("created_at DESC").Find(&videos)
	return videos
}

func GetAuthorIdByVideoId(videoId uint) (uint, bool) {
	var authorId uint
	find := VideoDB.Model(&model.Video{}).
		Select("author_id").
		Where("id = ?", videoId).
		Find(&authorId).RowsAffected != 0
	return authorId, find
}
