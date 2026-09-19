// Package video 提供视频/评论/点赞域（video / comments / favorite）的 MySQL 访问，
// 仅供 video 服务使用。
package video

import (
	"douyin/config"
	"douyin/dal/model"
	"douyin/dal/mysql/internal/conn"

	"gorm.io/gorm"
)

// DB 视频/评论/点赞域数据库句柄，由 video 服务独占。
var DB *gorm.DB

// Init 初始化 video 域数据库：库不存在则自动创建，启动时做幂等增量迁移；
// 配置了读副本时启用读写分离。
func Init(appConfig *config.AppConfig) (err error) {
	conf := conn.Conf(appConfig)
	db, err := conn.Open(conf, conf.VideoDatabase)
	if err != nil {
		return err
	}
	if err := conn.Migrate(db, "video", &model.Video{}, &model.Comment{}, &model.Favorite{}); err != nil {
		return err
	}
	if err := conn.AttachReplicas(db, conf, conf.VideoDatabase); err != nil {
		return err
	}
	DB = db
	return nil
}
