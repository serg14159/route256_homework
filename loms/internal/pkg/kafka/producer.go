package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"route256/loms/internal/models"
	"time"

	"github.com/IBM/sarama"
)

type IKafkaCfg interface {
	GetBrokers() []string
	GetTopic() string
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	cfg      IKafkaCfg
}

func NewKafkaProducer(cfg IKafkaCfg) (*KafkaProducer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(cfg.GetBrokers(), kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	kp := &KafkaProducer{
		producer: producer,
		cfg:      cfg,
	}

	return kp, nil
}

func (kp *KafkaProducer) SendOrderEvent(ctx context.Context, event *models.OrderEvent) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := fmt.Sprintf("%d", event.OrderID)

	msg := &sarama.ProducerMessage{
		Topic:     kp.cfg.GetTopic(),
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(eventBytes),
		Timestamp: time.Now(),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)

	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}
