package consumer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

const (
	MaxMetadataRetry     = 5
	MetadataRetryBackoff = time.Second * 10
	ConsumerRetryBackoff = time.Second * 5
	MaxAttempts          = 5
	RetryDelay           = time.Second * 5
)

type IKafkaCfg interface {
	GetBrokers() []string
	GetConsumerGroup() string
	GetTopics() []string
}

type IHandlerMessage interface {
	HandleMessage(msg *sarama.ConsumerMessage) error
}

type KafkaConsumer struct {
	cfg           IKafkaCfg
	handler       IHandlerMessage
	consumerGroup sarama.ConsumerGroup
	wg            sync.WaitGroup
}

// NewKafkaConsumer create new KafkaConsumer instance.
func NewKafkaConsumer(cfg IKafkaCfg, handler IHandlerMessage) (*KafkaConsumer, error) {
	// Consumer config
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Retries config
	config.Metadata.Retry.Max = MaxMetadataRetry
	config.Metadata.Retry.Backoff = MetadataRetryBackoff
	config.Consumer.Retry.Backoff = ConsumerRetryBackoff

	// Create consumer group
	var consumerGroup sarama.ConsumerGroup
	var err error

	for attempts := 0; attempts < MaxAttempts; attempts++ {
		consumerGroup, err = sarama.NewConsumerGroup(cfg.GetBrokers(), cfg.GetConsumerGroup(), config)
		if err == nil {
			break
		}
		log.Printf("Failed to create consumer group (attempt %d/%d): %v", attempts+1, MaxAttempts, err)
		time.Sleep(RetryDelay)
	}
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumerGroup: consumerGroup,
		cfg:           cfg,
		handler:       handler,
	}, nil
}

// Start begins consuming messages from Kafka.
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	log.Printf("Start consuming...")
	kc.wg.Add(1)
	defer kc.wg.Done()
	defer closeConsumerGroup(kc)

	go kc.handleConsumerErrors()

	// Consume messages from specified topics
	return kc.consumeLoop(ctx)
}

// closeConsumerGroup close consumer group.
func closeConsumerGroup(kc *KafkaConsumer) {
	if err := kc.consumerGroup.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
}

// handleConsumerErrors listens for errors from the consumer group.
func (kc *KafkaConsumer) handleConsumerErrors() {
	for err := range kc.consumerGroup.Errors() {
		log.Printf("Error from consumer group: %v", err)
	}
}

// consumeLoop continuously consumes messages.
func (kc *KafkaConsumer) consumeLoop(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := kc.consumerGroup.Consume(ctx, kc.cfg.GetTopics(), kc)
		if err != nil {
			log.Printf("Error during consumption: %v", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}
	}
}

// Wait blocks until KafkaConsumer finished work.
func (kc *KafkaConsumer) Wait() {
	kc.wg.Wait()
}

// Setup is a required method for sarama.ConsumerGroupHandler interface.
func (kc *KafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is a required method for sarama.ConsumerGroupHandler interface.
func (kc *KafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from Kafka partition claim.
func (kc *KafkaConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := kc.handler.HandleMessage(message); err != nil {
				log.Printf("Error handling message: %v", err)
				continue
			}
			sess.MarkMessage(message, "")
		case <-sess.Context().Done():
			return nil
		}
	}
}
