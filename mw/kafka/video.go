package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"douyin/common"
	"douyin/dal/model"
	"douyin/dal/mysql"
	"douyin/mw/redis"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type VideoMessage struct {
	VideoPath     string // API 暂存的本地视频文件路径
	VideoFileName string // 对象名（相对路径）
	UserID        uint
	Title         string
}

type VideoMQ struct {
	MQ
}

var VideoMQInstance *VideoMQ

// InitVideoKafka 初始化视频发布的生产者与消费者。
func InitVideoKafka() {
	VideoMQInstance = &VideoMQ{
		MQ{
			Topic:   "videos",
			GroupId: "video_group",
		},
	}
	// 显式确保 topic 存在：消费组先于 topic 启动时，分组订阅不会自动感知后建的 topic
	ensureTopic(VideoMQInstance.Topic)
	VideoMQInstance.Producer = kafkaManager.NewProducer(VideoMQInstance.Topic)
	VideoMQInstance.Consumer = kafkaManager.NewConsumer(VideoMQInstance.Topic, VideoMQInstance.GroupId)

	go VideoMQInstance.Consume()
}

// ensureTopic 若 topic 不存在则以 1 分区创建。
func ensureTopic(topic string) {
	client := &kafka.Client{Addr: kafka.TCP(kafkaManager.Brokers...)}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
	})
	if err != nil {
		zap.L().Warn("[VideoMQ] ensure topic failed (broker may auto-create)", zap.String("topic", topic), zap.Error(err))
		return
	}
	for topic, terr := range resp.Errors {
		if terr != nil && !errors.Is(terr, kafka.TopicAlreadyExists) {
			zap.L().Warn("[VideoMQ] create topic error", zap.String("topic", topic), zap.Error(terr))
		}
	}
}

// Produce 发布视频处理消息。
func (m *VideoMQ) Produce(message *VideoMessage) error {
	return kafkaManager.ProduceMessage(m.Producer, message)
}

// Consume 消费视频消息：持久化文件（OSS/本地）→ 截帧 → MySQL 落库 → 维护 feed ZSet 与作品数缓存。
func (m *VideoMQ) Consume() {
	for {
		msg, err := m.Consumer.ReadMessage(context.Background())
		if err != nil {
			zap.L().Error("[VideoMQ] read message failed", zap.Error(err))
			continue
		}

		videoMsg := new(VideoMessage)
		if err := json.Unmarshal(msg.Value, videoMsg); err != nil {
			zap.L().Error("[VideoMQ] unmarshal message failed", zap.Error(err))
			continue
		}
		m.handle(videoMsg)
	}
}

func (m *VideoMQ) handle(videoMsg *VideoMessage) {
	ctx := context.Background()

	// 1. 持久化视频文件：OSS 已配置则上传，否则保留在本地目录。
	if err := common.PersistFile(videoMsg.VideoPath, videoMsg.VideoFileName); err != nil {
		zap.L().Error("[VideoMQ] persist video failed", zap.Error(err))
		return
	}

	// 2. 生成封面（OSS 数据万象 / 本地 ffmpeg）。
	coverName, err := common.GenerateCover(videoMsg.VideoPath, videoMsg.VideoFileName)
	if err != nil {
		zap.L().Error("[VideoMQ] generate cover failed", zap.Error(err))
		return
	}

	// 3. 视频信息落库 MySQL（权威数据源）。
	video := &model.Video{
		AuthorId:  videoMsg.UserID,
		VideoUrl:  videoMsg.VideoFileName,
		CoverUrl:  coverName,
		Title:     videoMsg.Title,
		CreatedAt: time.Now().Unix(),
	}
	if !mysql.InsertVideo(video) {
		zap.L().Error("[VideoMQ] insert video to mysql failed")
		return
	}

	// 4. 维护 feed ZSet、失效作品数缓存、登记 Bloom。
	if err := redis.AddVideoToFeed(ctx, video.ID, video.CreatedAt); err != nil {
		zap.L().Error("[VideoMQ] add video to feed failed", zap.Error(err))
	}
	redis.InvalidateWorkCount(ctx, videoMsg.UserID)
	common.AddToWorkCountBloom(strconv.FormatUint(uint64(videoMsg.UserID), 10))

	// 5. OSS 模式下删除本地暂存文件；本地降级模式保留文件用于 /static 访问。
	if common.OSSConfigured() {
		if err := os.Remove(videoMsg.VideoPath); err != nil && !os.IsNotExist(err) {
			zap.L().Warn("[VideoMQ] remove temp video failed", zap.Error(err))
		}
	}

	zap.L().Info("[VideoMQ] video processed", zap.Uint64("video_id", uint64(video.ID)))
}
