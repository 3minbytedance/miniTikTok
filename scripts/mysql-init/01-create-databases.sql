-- 为 social / video 微服务分别创建独立数据库（schema 隔离，同一 MySQL 实例）
CREATE DATABASE IF NOT EXISTS douyin_social DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS douyin_video DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 授权业务账号（业务表由各服务启动时的 GORM AutoMigrate 自动创建）
GRANT ALL PRIVILEGES ON douyin_social.* TO 'doushen'@'%';
GRANT ALL PRIVILEGES ON douyin_video.* TO 'doushen'@'%';

-- 复制专用账号（GTID 自动定位，从库 CHANGE REPLICATION SOURCE 使用）
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED WITH caching_sha2_password BY 'repl123' REQUIRE NONE;
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl'@'%';

FLUSH PRIVILEGES;
