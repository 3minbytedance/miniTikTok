// Package social 提供用户与关系链域（user_login / user_info / user_follow）的 MySQL 访问，
// 仅供 social 服务使用。
package social

import (
	"douyin/config"
	"douyin/dal/model"
	"douyin/dal/mysql/internal/conn"

	"gorm.io/gorm"
)

// DB 用户/关系域数据库句柄，由 social 服务独占。
var DB *gorm.DB

// Init 初始化 social 域数据库：库不存在则自动创建，启动时做幂等增量迁移；
// 配置了读副本时启用读写分离。
func Init(appConfig *config.AppConfig) (err error) {
	conf := conn.Conf(appConfig)
	db, err := conn.Open(conf, conf.SocialDatabase)
	if err != nil {
		return err
	}
	if err := conn.Migrate(db, "social", &model.User{}, &model.UserInfo{}, &model.UserFollow{}); err != nil {
		return err
	}
	if err := conn.AttachReplicas(db, conf, conf.SocialDatabase); err != nil {
		return err
	}
	DB = db
	return nil
}
