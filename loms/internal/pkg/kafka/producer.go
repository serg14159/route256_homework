package kafka

import (
	"context"
	"fmt"
	"route256/loms/internal/models"
	"route256/utils/logger"
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

// NewKafkaProducer creates a new KafkaProducer instance.
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

// SendOutboxEvent send message containing OutboxEvent to Kafka.
func (kp *KafkaProducer) SendOutboxEvent(ctx context.Context, event *models.OutboxEvent) error {
	key := fmt.Sprintf("%d", event.ID)

	msg := &sarama.ProducerMessage{
		Topic:     kp.cfg.GetTopic(),
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(event.Payload),
		Timestamp: time.Now(),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	logger.Infow(ctx, fmt.Sprintf("Message sent to partition %d at offset %d\n", partition, offset))

	return nil
}

// Close gracefully close Kafka producer connection.
func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}
