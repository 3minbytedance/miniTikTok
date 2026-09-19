package mysql

import (
	"douyin/config"
	"douyin/dal/model"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

func Init(appConfig *config.AppConfig) (err error) {
	var conf *config.MySQLConfig
	if appConfig.Mode == config.LocalMode {
		conf = appConfig.Local.MySQLConfig
	} else {
		conf = appConfig.Remote.MySQLConfig
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Username,
		conf.Password,
		conf.Address,
		conf.Port,
		conf.Database,
	)

	mysqlLog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Error,
			Colorful:                  false,
		})

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: mysqlLog, PrepareStmt: true})
	if err != nil {
		return fmt.Errorf("connect to mysql failed: %w", err)
	}
	// 启动时自动建表/补列（幂等，仅增量迁移）
	if err := DB.AutoMigrate(
		&model.User{},
		&model.UserInfo{},
		&model.Video{},
		&model.Comment{},
		&model.Favorite{},
		&model.UserFollow{},
	); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}
	return nil
}
