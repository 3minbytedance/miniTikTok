package mongo

import (
	"context"
	"douyin/config"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

var Mongo *mongo.Database
var Ctx context.Context

func Init(appConfig *config.AppConfig) (err error) {
	var conf *config.MongoConfig
	if appConfig.Mode == config.LocalMode {
		conf = appConfig.Local.MongoConfig
	} else {
		conf = appConfig.Remote.MongoConfig
	}
	Ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoUrl := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", conf.Username, conf.Password, conf.Address, conf.Port, conf.DB)
	client, err := mongo.Connect(Ctx, options.Client().ApplyURI(mongoUrl))

	if err != nil {
		zap.L().Error("Connection MongoDB Error:", zap.Error(err))
		return
	}

	// 检查连接
	err = client.Ping(Ctx, nil)
	if err != nil {
		zap.L().Error("Connection MongoDB Error:", zap.Error(err))
		return
	}

	Mongo = client.Database(conf.DB)
	return
}
