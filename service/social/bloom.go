package main

import (
	"log"

	"douyin/dal/model"
	"douyin/dal/mysql"

	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
)

// social 域 Bloom 过滤器：用户名判重、关注/粉丝 id 防穿透。
var (
	bloomUserFilter               *bloom.BloomFilter
	bloomRelationFollowIdFilter   *bloom.BloomFilter
	bloomRelationFollowerIdFilter *bloom.BloomFilter
)

func InitUserBloomFilter() {
	// 假设预期元素数量为 100000，误判率为 0.01
	bloomUserFilter = bloom.NewWithEstimates(100000, 0.01)
}

func InitRelationFollowIdFilter() {
	// 初始化关注布隆过滤器
	bloomRelationFollowIdFilter = bloom.NewWithEstimates(100000, 0.01)
}

func InitRelationFollowerIdFilter() {
	// 初始化粉丝布隆过滤器
	bloomRelationFollowerIdFilter = bloom.NewWithEstimates(100000, 0.01)
}

func AddToUserBloom(data string) {
	bloomUserFilter.Add([]byte(data))
}

func TestUserBloom(data string) bool {
	return bloomUserFilter.Test([]byte(data))
}

// LoadUsernamesToBloomFilter 启动时从 social 库预载全部用户名。
func LoadUsernamesToBloomFilter() {
	var usernames []string
	if err := mysql.SocialDB.Model(&model.User{}).Pluck("name", &usernames).Error; err != nil {
		log.Fatal("Failed to retrieve usernames from database:", err)
	}

	for _, username := range usernames {
		AddToUserBloom(username)
	}

	zap.L().Info("Loaded usernames to the bloom filter.", zap.Int("size", len(usernames)))
}

func AddToRelationFollowIdBloom(data string) {
	bloomRelationFollowIdFilter.Add([]byte(data))
}

func TestRelationFollowIdBloom(data string) bool {
	return bloomRelationFollowIdFilter.Test([]byte(data))
}

// LoadRelationFollowIdToBloomFilter 启动时从 social 库预载全部关注者 id。
func LoadRelationFollowIdToBloomFilter() {
	var followIdList []string
	if err := mysql.SocialDB.Model(&model.UserFollow{}).Distinct().Pluck("user_id", &followIdList).Error; err != nil {
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

// LoadRelationFollowerIdToBloomFilter 启动时从 social 库预载全部粉丝 id。
func LoadRelationFollowerIdToBloomFilter() {
	var followerIdList []string
	if err := mysql.SocialDB.Model(&model.UserFollow{}).Distinct().Pluck("follow_id", &followerIdList).Error; err != nil {
		log.Fatal("Failed to retrieve follower ids from database:", err)
	}
	for _, followerId := range followerIdList {
		AddToRelationFollowerIdBloom(followerId)
	}
	zap.L().Info("Loaded follower from follow to the bloom filter.", zap.Int("size", len(followerIdList)))
}
