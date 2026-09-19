-- 从库初始化：配置复制源（GTID 自动定位）并启动复制线程。
-- 主库建库/建表/授权语句均已写入 binlog，由复制自动同步，无需在从库重复执行。
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST = 'mysql',
  SOURCE_PORT = 3306,
  SOURCE_USER = 'repl',
  SOURCE_PASSWORD = 'repl123',
  SOURCE_AUTO_POSITION = 1,
  GET_SOURCE_PUBLIC_KEY = 1;

START REPLICA;
