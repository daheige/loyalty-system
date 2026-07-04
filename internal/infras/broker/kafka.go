package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	godBroker "github.com/go-god/broker"
	"github.com/go-god/broker/gkafka"
)

// TargetTopic 定义业务类型与 Topic 的映射关系
type TargetTopic struct {
	Type  string `mapstructure:"type"`
	Topic string `mapstructure:"topic"`
}

// Event 定义忠诚度事件
type Event struct {
	Type      string          `json:"type"`
	ShopID    string          `json:"shop_id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// Message 封装订阅到的消息
type Message struct {
	Topic   string
	Payload []byte
}

// EventType 定义事件类型
const (
	EventTypeOrderPaid      = "shopify.order.paid"
	EventTypeReviewCreated  = "review.created"
	EventTypeMemberCheckin  = "member.checkin"
	EventTypePointsEarned   = "points.earned"
	EventTypePointsSpent    = "points.spent"
	EventTypeTierUpgraded   = "tier.upgraded"
	EventTypeTierDowngraded = "tier.downgraded"
	EventTypePointsExpired  = "points.expired"
)

// eventCategories 定义事件类型对应的业务分类
var eventCategories = map[string]string{
	EventTypeOrderPaid:      "events",
	EventTypeReviewCreated:  "events",
	EventTypeMemberCheckin:  "events",
	EventTypePointsEarned:   "points",
	EventTypePointsSpent:    "points",
	EventTypePointsExpired:  "points",
	EventTypeTierUpgraded:   "tiers",
	EventTypeTierDowngraded: "tiers",
}

// defaultTargetTopics 默认 Topic 映射
func defaultTargetTopics() map[string]string {
	return map[string]string{
		"events": "loyalty.events",
		"points": "loyalty.points",
		"tiers":  "loyalty.tiers",
	}
}

// Broker 封装消息队列
type Broker struct {
	broker       godBroker.Broker
	groupID      string
	targetTopics map[string]string
}

// NewBroker 创建 Broker 实例
func NewBroker(brokers []string, groupID string, targetTopics []TargetTopic) (*Broker, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers is empty")
	}

	b := gkafka.New(godBroker.WithBrokerAddress(brokers...))

	topicMap := make(map[string]string)
	for _, t := range targetTopics {
		if t.Type == "" || t.Topic == "" {
			continue
		}
		topicMap[t.Type] = t.Topic
	}

	// 默认 Topic 映射，确保未配置时系统仍可运行
	if len(topicMap) == 0 {
		topicMap = defaultTargetTopics()
	}

	return &Broker{
		broker:       b,
		groupID:      groupID,
		targetTopics: topicMap,
	}, nil
}

// ResolveTopic 根据事件类型从配置中解析目标 Topic
func (b *Broker) ResolveTopic(eventType string) string {
	category, ok := eventCategories[eventType]
	if !ok {
		category = "events"
	}

	if topic, ok := b.targetTopics[category]; ok {
		return topic
	}

	if topic, ok := b.targetTopics["events"]; ok {
		return topic
	}

	return "loyalty.events"
}

// Publish 发布事件到指定 Topic
func (b *Broker) Publish(ctx context.Context, topic string, eventType string, shopID string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	event := Event{
		Type:      eventType,
		ShopID:    shopID,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	opts := []godBroker.PubOption{}
	if shopID != "" {
		opts = append(opts, godBroker.WithPublishName(shopID))
	}

	return b.broker.Publish(ctx, topic, eventBytes, opts...)
}

// Subscribe 订阅指定 Topic 的事件
func (b *Broker) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, msg Message) error) error {
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}

	h := func(ctx context.Context, value []byte) error {
		return handler(ctx, Message{Topic: topic, Payload: value})
	}

	return b.broker.Subscribe(ctx, topic, b.groupID, h)
}

// Topics 返回配置中所有需要去重的 Topic 列表
func (b *Broker) Topics() []string {
	var topics []string
	seen := make(map[string]struct{})
	for _, topic := range b.targetTopics {
		if _, ok := seen[topic]; ok {
			continue
		}
		seen[topic] = struct{}{}
		topics = append(topics, topic)
	}
	return topics
}

// Close 关闭连接
func (b *Broker) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := b.broker.Shutdown(ctx); err != nil {
		log.Printf("close broker error: %v", err)
	}
	return nil
}
