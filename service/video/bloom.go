package main

import (
	"log"
	"strconv"

	"douyin/dal/model"
	"douyin/dal/mysql"

	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
)

// video 域 Bloom 过滤器：评论视频 id、作品数作者 id、点赞关系防穿透。
var (
	bloomCommentFilter         *bloom.BloomFilter
	bloomWorkCountFilter       *bloom.BloomFilter
	bloomIsFavoriteFilter      *bloom.BloomFilter
	bloomFavoriteVideoIdFilter *bloom.BloomFilter
)

func InitCommentBloomFilter() {
	// 初始化评论布隆过滤器
	bloomCommentFilter = bloom.NewWithEstimates(100000, 0.01)
}

func InitWorkCountFilter() {
	// 初始化作品数布隆过滤器
	bloomWorkCountFilter = bloom.NewWithEstimates(100000, 0.01)
}

func InitIsFavoriteFilter() {
	// 初始化用户点赞数布隆过滤器
	bloomIsFavoriteFilter = bloom.NewWithEstimates(100000, 0.01)
}

func InitFavoriteVideoIdFilter() {
	// 初始化视频点赞数布隆过滤器
	bloomFavoriteVideoIdFilter = bloom.NewWithEstimates(100000, 0.01)
}

func AddToCommentBloom(data string) {
	bloomCommentFilter.Add([]byte(data))
}

func TestCommentBloom(data string) bool {
	return bloomCommentFilter.Test([]byte(data))
}

// LoadCommentVideoIdToBloomFilter 启动时从 video 库预载有评论的视频 id。
func LoadCommentVideoIdToBloomFilter() {
	var videoIdList []string
	if err := mysql.VideoDB.Model(&model.Comment{}).Distinct().Pluck("video_id", &videoIdList).Error; err != nil {
		log.Fatal("Failed to retrieve comment video ids from database:", err)
	}
	for _, videoId := range videoIdList {
		AddToCommentBloom(videoId)
	}
	zap.L().Info("Loaded comments to the bloom filter.", zap.Int("size", len(videoIdList)))
}

func AddToWorkCountBloom(data string) {
	bloomWorkCountFilter.Add([]byte(data))
}

func TestWorkCountBloom(data string) bool {
	return bloomWorkCountFilter.Test([]byte(data))
}

// LoadWorkCountToBloomFilter 启动时从 video 库预载有作品的作者 id。
func LoadWorkCountToBloomFilter() {
	var authorIdList []string
	if err := mysql.VideoDB.Model(&model.Video{}).Distinct().Pluck("author_id", &authorIdList).Error; err != nil {
		log.Fatal("Failed to retrieve author ids from database:", err)
	}
	for _, authorId := range authorIdList {
		AddToWorkCountBloom(authorId)
	}
	zap.L().Info("Loaded authors from video to the bloom filter.", zap.Int("size", len(authorIdList)))
}

func AddToIsFavoriteBloom(userId, videoId uint) {
	data := strconv.Itoa(int(userId)) + strconv.Itoa(int(videoId))
	bloomIsFavoriteFilter.Add([]byte(data))
}

func TestIsFavoriteBloom(userId, videoId uint) bool {
	data := strconv.Itoa(int(userId)) + strconv.Itoa(int(videoId))
	return bloomIsFavoriteFilter.Test([]byte(data))
}

// LoadIsFavoriteToBloomFilter 启动时从 video 库预载全部点赞关系。
func LoadIsFavoriteToBloomFilter() {
	var favoriteList []model.Favorite
	if err := mysql.VideoDB.Model(&model.Favorite{}).Find(&favoriteList).Error; err != nil {
		log.Fatal("Failed to retrieve favorites from database:", err)
	}
	for _, favorite := range favoriteList {
		AddToIsFavoriteBloom(favorite.UserId, favorite.VideoId)
	}
	zap.L().Info("Loaded favorites to the bloom filter.", zap.Int("size", len(favoriteList)))
}

func AddToFavoriteVideoIdBloom(data string) {
	bloomFavoriteVideoIdFilter.Add([]byte(data))
}

func TestFavoriteVideoIdBloom(data string) bool {
	return bloomFavoriteVideoIdFilter.Test([]byte(data))
}

// LoadFavoriteVideoIdToBloomFilter 启动时从 video 库预载被点赞的视频 id。
func LoadFavoriteVideoIdToBloomFilter() {
	var videoIdList []string
	if err := mysql.VideoDB.Model(&model.Favorite{}).Distinct().Pluck("video_id", &videoIdList).Error; err != nil {
		log.Fatal("Failed to retrieve favorite video ids from database:", err)
	}
	for _, videoId := range videoIdList {
		AddToFavoriteVideoIdBloom(videoId)
	}
	zap.L().Info("Loaded video from favorite to the bloom filter.", zap.Int("size", len(videoIdList)))
}
