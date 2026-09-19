package common

import (
	"douyin/dal/model"
	"douyin/dal/mysql"
	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
	"log"
	"strconv"
)

var bloomUserFilter *bloom.BloomFilter
var bloomCommentFilter *bloom.BloomFilter
var bloomWorkCountFilter *bloom.BloomFilter
var bloomIsFavoriteFilter *bloom.BloomFilter
var bloomFavoriteVideoIdFilter *bloom.BloomFilter
var bloomRelationFollowIdFilter *bloom.BloomFilter
var bloomRelationFollowerIdFilter *bloom.BloomFilter

func InitUserBloomFilter() {
	// 初始化用户名布隆过滤器
	bloomUserFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitCommentBloomFilter() {
	// 初始化评论布隆过滤器
	bloomCommentFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitWorkCountFilter() {
	// 初始化作品数布隆过滤器
	bloomWorkCountFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitIsFavoriteFilter() {
	// 初始化用户点赞数布隆过滤器
	bloomIsFavoriteFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitFavoriteVideoIdFilter() {
	// 初始化视频点赞数布隆过滤器
	bloomFavoriteVideoIdFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitRelationFollowIdFilter() {
	// 初始化关注布隆过滤器
	bloomRelationFollowIdFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func InitRelationFollowerIdFilter() {
	// 初始化粉丝布隆过滤器
	bloomRelationFollowerIdFilter = bloom.NewWithEstimates(100000, 0.01) // 假设预期元素数量为 100000，误判率为 0.01
}

func AddToUserBloom(data string) {
	bloomUserFilter.Add([]byte(data))
}

func TestUserBloom(data string) bool {
	return bloomUserFilter.Test([]byte(data))
}

func LoadUsernamesToBloomFilter() {
	var usernames []string
	err := mysql.DB.Model(&model.User{}).Pluck("name", &usernames).Error
	if err != nil {
		log.Fatal("Failed to retrieve usernames from database:", err)
	}

	for _, username := range usernames {
		AddToUserBloom(username)
	}

	zap.L().Info("Loaded usernames to the bloom filter.", zap.Int("size", len(usernames)))
}

func AddToCommentBloom(data string) {
	bloomCommentFilter.Add([]byte(data))
}

func TestCommentBloom(data string) bool {
	return bloomCommentFilter.Test([]byte(data))
}

func LoadCommentVideoIdToBloomFilter() {
	var videoIdList []string
	if err := mysql.DB.Model(&model.Comment{}).Distinct().Pluck("video_id", &videoIdList).Error; err != nil {
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

func LoadWorkCountToBloomFilter() {
	var authorIdList []string
	if err := mysql.DB.Model(&model.Video{}).Distinct().Pluck("author_id", &authorIdList).Error; err != nil {
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

func LoadIsFavoriteToBloomFilter() {
	var favoriteList []model.Favorite
	if err := mysql.DB.Model(&model.Favorite{}).Find(&favoriteList).Error; err != nil {
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

func LoadFavoriteVideoIdToBloomFilter() {
	var videoIdList []string
	if err := mysql.DB.Model(&model.Favorite{}).Distinct().Pluck("video_id", &videoIdList).Error; err != nil {
		log.Fatal("Failed to retrieve favorite video ids from database:", err)
	}
	for _, videoId := range videoIdList {
		AddToFavoriteVideoIdBloom(videoId)
	}
	zap.L().Info("Loaded video from favorite to the bloom filter.", zap.Int("size", len(videoIdList)))
}

func AddToRelationFollowIdBloom(data string) {
	bloomRelationFollowIdFilter.Add([]byte(data))
}

func TestRelationFollowIdBloom(data string) bool {
	return bloomRelationFollowIdFilter.Test([]byte(data))
}

func LoadRelationFollowIdToBloomFilter() {
	var followIdList []string
	if err := mysql.DB.Model(&model.UserFollow{}).Distinct().Pluck("user_id", &followIdList).Error; err != nil {
		log.Fatal("Failed to retrieve follow ids from database:", err)
	}
	for _, followId := range followIdList {
		AddToRelationFollowIdBloom(followId)
	}
	zap.L().Info("Loaded followId from follow to the bloom filter.", zap.Int("size", len(followIdList)))
}

func AddToRelationFollowerIdBloom(data string) {
	bloomRelationFollowerIdFilter.Add([]byte(data))
}

func TestRelationFollowerIdBloom(data string) bool {
	return bloomRelationFollowerIdFilter.Test([]byte(data))
}

func LoadRelationFollowerIdToBloomFilter() {
	var followerIdList []string
	if err := mysql.DB.Model(&model.UserFollow{}).Distinct().Pluck("follow_id", &followerIdList).Error; err != nil {
		log.Fatal("Failed to retrieve follower ids from database:", err)
	}
	for _, followerId := range followerIdList {
		AddToRelationFollowerIdBloom(followerId)
	}
	zap.L().Info("Loaded follower from follow to the bloom filter.", zap.Int("size", len(followerIdList)))
}
