package kafka

import (
	"context"
	"douyin/config"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"strconv"
	"time"
)

type Manager struct {
	Brokers []string
}

var kafkaManager *Manager

type MQ struct {
	Topic    string
	GroupId  string
	Producer *kafka.Writer
	Consumer *kafka.Reader
}

func Init(appConfig *config.AppConfig) (err error) {
	var conf *config.KafkaConfig
	if appConfig.Mode == config.LocalMode {
		conf = appConfig.Local.KafkaConfig
	} else {
		conf = appConfig.Remote.KafkaConfig
	}
	brokerUrl := conf.Address + ":" + strconv.Itoa(conf.Port)
	// 初始化 Kafka Manager
	brokers := []string{brokerUrl}
	kafkaManager = NewKafkaManager(brokers)

	return nil
}

// Brokers 返回当前 broker 地址列表。
func Brokers() []string {
	return kafkaManager.Brokers
}

// NewProducer 创建指定 topic 的生产者。
func NewProducer(topic string) *kafka.Writer {
	return kafkaManager.NewProducer(topic)
}

// NewConsumer 创建指定 topic/group 的消费者。
func NewConsumer(topic, groupId string) *kafka.Reader {
	return kafkaManager.NewConsumer(topic, groupId)
}

// ProduceMessage 向 Kafka 写入消息的公共函数, 由于不同业务的消息格式不同, 所以使用 interface{} 代替
func ProduceMessage(producer *kafka.Writer, message interface{}) error {
	return kafkaManager.ProduceMessage(producer, message)
}

func NewKafkaManager(brokers []string) *Manager {
	return &Manager{
		Brokers: brokers,
	}
}

func (m *Manager) NewProducer(topic string) *kafka.Writer {
	// TODO writer 优雅关闭
	return &kafka.Writer{
		Addr:                   kafka.TCP(m.Brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{}, // 使用Hash算法按照key将消息均匀分布到不同的partition上
		WriteTimeout:           1 * time.Second,
		RequiredAcks:           kafka.RequireAll, // 需要确保Leader和所有Follower都写入成功才可以发送下一条消息, 确保消息成功写入, 不丢失
		AllowAutoTopicCreation: true,             // Topic不存在时自动创建。生产环境中一般设为false，由运维管理员创建Topic并配置partition数目
	}
}

func (m *Manager) NewConsumer(topic, groupId string) *kafka.Reader {
	// TODO reader 优雅关闭
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: m.Brokers,
		Topic:   topic,
		GroupID: groupId,
		// CommitInterval: 1 * time.Second, // 不配置此项, 默认每次读取都会自动提交offset
		StartOffset: kafka.FirstOffset, //当一个特定的partition没有commited offset时(比如第一次读一个partition，之前没有commit过)，通过StartOffset指定从第一个还是最后一个位置开始消费。StartOffset的取值要么是FirstOffset要么是LastOffset，LastOffset表示Consumer启动之前生成的老数据不管了。仅当指定了GroupID时，StartOffset才生效
	})
}

// ProduceMessage 向 Kafka 写入消息的公共函数, 由于不同业务的消息格式不同, 所以使用 interface{} 代替
func (m *Manager) ProduceMessage(producer *kafka.Writer, message interface{}) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return producer.WriteMessages(context.Background(), kafka.Message{
		Value: messageBytes,
	})
}
