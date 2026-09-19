// Package conn 提供 MySQL 建库、连接、连接池、迁移与读写分离的公共实现。
// 位于 internal 下，仅允许 dal/mysql 下的域包复用，服务层不得直接依赖。
package conn

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"douyin/config"

	_ "github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

const (
	maxOpenConns    = 100
	maxIdleConns    = 20
	connMaxLifetime = time.Hour
	connMaxIdleTime = 10 * time.Minute
)

// Conf 按运行模式返回对应的 MySQL 配置。
func Conf(appConfig *config.AppConfig) *config.MySQLConfig {
	if appConfig.Mode == config.LocalMode {
		return appConfig.Local.MySQLConfig
	}
	return appConfig.Remote.MySQLConfig
}

// Open 目标库不存在则创建，随后建立连接并设置连接池。
func Open(conf *config.MySQLConfig, database string) (*gorm.DB, error) {
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

	db, err := gorm.Open(gormMysql.Open(dsn(conf, conf.Address, conf.Port, database)), &gorm.Config{
		Logger:         mysqlLog,
		PrepareStmt:    true,
		TranslateError: true, // 暴露 gorm.ErrDuplicatedKey 等语义化错误，便于区分唯一键冲突与真实故障
	})
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed: %w", err)
	}

	// 连接池上限与空闲回收：避免请求高峰无限制建连，也避免长期空闲连接被服务端断开后仍被复用。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
	return db, nil
}

// Migrate 幂等增量迁移，domain 仅用于错误信息定位。
func Migrate(db *gorm.DB, domain string, models ...any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("%s auto migrate failed: %w", domain, err)
	}
	return nil
}

// AttachReplicas 注册读写分离：写/事务走主库，读在副本间随机负载均衡。
// 副本列表为空时不注册 resolver，保持单库行为。
func AttachReplicas(db *gorm.DB, conf *config.MySQLConfig, database string) error {
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

func dsn(conf *config.MySQLConfig, address string, port int, database string) string {
	timeout := conf.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		conf.Username,
		conf.Password,
		address,
		port,
		database,
		timeout,
		timeout,
		timeout,
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
