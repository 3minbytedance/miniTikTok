package redis

import (
	"context"
	"douyin/config"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var Ctx = context.Background()

// Rdb Comment模块Rdb
var Rdb *redis.Client

// RdbExpireTime key的过期时间
var RdbExpireTime time.Duration

func Init(appConfig *config.AppConfig) (err error) {
	var conf *config.RedisConfig
	if appConfig.Mode == config.LocalMode {
		conf = appConfig.Local.RedisConfig
	} else {
		conf = appConfig.Remote.RedisConfig
	}
	// 获取conf中的过期时间, 单位为s
	RdbExpireTime = time.Duration(conf.ExpireTime) * time.Second

	Rdb = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", conf.Address, conf.Port),
		Password:     conf.Password, // 密码
		DB:           conf.DB,       // 数据库
		PoolSize:     conf.PoolSize, // 连接池大小
		MinIdleConns: conf.MinIdleConns,
	})
	if err = Rdb.Ping(Ctx).Err(); err != nil {
		return err
	}
	return
}
