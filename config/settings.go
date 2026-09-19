package config

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	Conf      = new(AppConfig)
	LocalMode = "local"
)

type AppConfig struct {
	Name       string `mapstructure:"name"`
	Node       string `mapstructure:"node"`
	Mode       string `mapstructure:"mode"`
	Port       int    `mapstructure:"port"`
	Version    string `mapstructure:"version"`
	*LogConfig `mapstructure:"log"`
	*OSSConfig `mapstructure:"oss"`
	*ObsConfig `mapstructure:"observability"`

	Local struct {
		*MySQLConfig `mapstructure:"mysql"`
		*RedisConfig `mapstructure:"redis"`
		*KafkaConfig `mapstructure:"kafka"`
		*MongoConfig `mapstructure:"mongo"`
	} `mapstructure:"local"`

	Remote struct {
		*MySQLConfig `mapstructure:"mysql"`
		*RedisConfig `mapstructure:"redis"`
		*KafkaConfig `mapstructure:"kafka"`
		*MongoConfig `mapstructure:"mongo"`
	} `mapstructure:"remote"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Username       string         `mapstructure:"username"`
	Password       string         `mapstructure:"password"`
	Address        string         `mapstructure:"address"`
	Port           int            `mapstructure:"port"`
	SocialDatabase string         `mapstructure:"social_database"` // social 服务库（用户/关系域）
	VideoDatabase  string         `mapstructure:"video_database"`  // video 服务库（视频/评论/点赞域）
	Replicas       []MySQLReplica `mapstructure:"replicas"`        // 读副本列表，为空则不启用读写分离（全部走主库）
	Timeout        int            `mapstructure:"timeout"`
}

// MySQLReplica 读副本连接信息，账号密码复用主库配置。
type MySQLReplica struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type RedisConfig struct {
	Address      string `mapstructure:"address"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	ExpireTime   int64  `mapstructure:"expire_time"`
}

type KafkaConfig struct {
	Address  string `mapstructure:"address"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type MongoConfig struct {
	Address  string `mapstructure:"address"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       string `mapstructure:"db"`
}

// OSSConfig 对象存储配置。Enabled=false（或关键字段为空）时自动降级到本地文件存储。
type OSSConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	BucketURL         string `mapstructure:"bucket_url"`    // COS Bucket 访问地址
	CIBucketURL       string `mapstructure:"ci_bucket_url"` // COS 数据万象地址（视频截帧）
	SecretID          string `mapstructure:"secret_id"`
	SecretKey         string `mapstructure:"secret_key"`
	SessionToken      string `mapstructure:"session_token"`
	Region            string `mapstructure:"region"`
	Bucket            string `mapstructure:"bucket"`
	LocalFallbackPath string `mapstructure:"local_fallback_path"` // 本地降级存储目录
	StaticURLPrefix   string `mapstructure:"static_url_prefix"`   // 本地文件对外访问的 URL 前缀
}

// ObsConfig 可观测性配置（OpenTelemetry / Prometheus）。
type ObsConfig struct {
	ServiceName    string `mapstructure:"service_name"`
	CollectorAddr  string `mapstructure:"collector_addr"` // otel collector 地址，为空则只导出日志/metrics
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	TraceEnabled   bool   `mapstructure:"trace_enabled"`
}

func Init() (err error) {
	// 可通过 DOUYIN_CONFIG 环境变量指定配置文件（容器部署时指向 app.docker.yaml），默认本地配置。
	configFile := os.Getenv("DOUYIN_CONFIG")
	if configFile == "" {
		configFile = "config/app.yaml"
	}
	viper.SetConfigFile(configFile) // 指定配置文件路径
	err = viper.ReadInConfig()      // 读取配置信息
	if err != nil {                 // 读取配置信息失败¬
		log.Fatalf("Read app.yaml failed: %s \n", err)
	}

	// 读取到的配置信息 反序列化到 Conf 里面
	if err := viper.Unmarshal(Conf); err != nil {
		log.Printf("Viper unmarshal failed: %v\n", err)
	}

	// 监控配置文件变化, 实时更新Conf
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("配置发生变化了...")
		if err := viper.Unmarshal(Conf); err != nil {
			log.Printf("Viper unmarshal failed, err: %v\n", err)
		}
	})

	return
}
