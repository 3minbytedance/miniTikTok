package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"douyin/config"
	"douyin/dal/model"

	_ "github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	// SocialDB 用户/关系域数据库句柄，由 social 服务独占。
	SocialDB *gorm.DB
	// VideoDB 视频/评论/点赞域数据库句柄，由 video 服务独占。
	VideoDB *gorm.DB
)

// InitSocial 初始化 social 域数据库（user_login / user_info / user_follow），
// 库不存在则自动创建，启动时做幂等增量迁移；配置了读副本时启用读写分离。
func InitSocial(appConfig *config.AppConfig) (err error) {
	conf := mysqlConf(appConfig)
	db, err := open(conf, conf.SocialDatabase)
	if err != nil {
		return err
	}
	if err := migrate(db, "social", &model.User{}, &model.UserInfo{}, &model.UserFollow{}); err != nil {
		return err
	}
	if err := attachReplicas(db, conf, conf.SocialDatabase); err != nil {
		return err
	}
	SocialDB = db
	return nil
}

// InitVideo 初始化 video 域数据库（video / comments / favorite），
// 库不存在则自动创建，启动时做幂等增量迁移；配置了读副本时启用读写分离。
func InitVideo(appConfig *config.AppConfig) (err error) {
	conf := mysqlConf(appConfig)
	db, err := open(conf, conf.VideoDatabase)
	if err != nil {
		return err
	}
	if err := migrate(db, "video", &model.Video{}, &model.Comment{}, &model.Favorite{}); err != nil {
		return err
	}
	if err := attachReplicas(db, conf, conf.VideoDatabase); err != nil {
		return err
	}
	VideoDB = db
	return nil
}

func mysqlConf(appConfig *config.AppConfig) *config.MySQLConfig {
	if appConfig.Mode == config.LocalMode {
		return appConfig.Local.MySQLConfig
	}
	return appConfig.Remote.MySQLConfig
}

func migrate(db *gorm.DB, domain string, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("%s auto migrate failed: %w", domain, err)
	}
	return nil
}

// attachReplicas 注册读写分离：写/事务走主库，读在副本间随机负载均衡。
// 副本列表为空时不注册 resolver，保持单库行为。
func attachReplicas(db *gorm.DB, conf *config.MySQLConfig, database string) error {
	if len(conf.Replicas) == 0 {
		return nil
	}
	replicas := make([]gorm.Dialector, 0, len(conf.Replicas))
	for _, r := range conf.Replicas {
		replicas = append(replicas, gormMysql.Open(dsn(conf, r.Address, r.Port, database)))
	}
	return db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{gormMysql.Open(dsn(conf, conf.Address, conf.Port, database))},
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}))
}

func open(conf *config.MySQLConfig, database string) (*gorm.DB, error) {
	if err := ensureDatabase(conf, database); err != nil {
		return nil, err
	}

	mysqlLog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Error,
			Colorful:                  false,
		})

	db, err := gorm.Open(gormMysql.Open(dsn(conf, conf.Address, conf.Port, database)), &gorm.Config{Logger: mysqlLog, PrepareStmt: true})
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed: %w", err)
	}
	return db, nil
}

func dsn(conf *config.MySQLConfig, address string, port int, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Username,
		conf.Password,
		address,
		port,
		database,
	)
}

// ensureDatabase 目标库不存在则创建（幂等）。
// 连接账号无建库权限时，回退检查库是否已存在，已存在则视为成功。
func ensureDatabase(conf *config.MySQLConfig, database string) error {
	raw, err := sql.Open("mysql", dsn(conf, conf.Address, conf.Port, ""))
	if err != nil {
		return fmt.Errorf("open mysql without database failed: %w", err)
	}
	defer raw.Close()

	if _, err := raw.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", database,
	)); err != nil {
		var cnt int
		if qerr := raw.QueryRow(
			"SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", database,
		).Scan(&cnt); qerr != nil || cnt == 0 {
			return fmt.Errorf("ensure database %s failed: %w", database, err)
		}
	}
	return nil
}
