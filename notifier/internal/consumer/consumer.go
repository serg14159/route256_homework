package consumer

import (
	"context"
	"log"

	"github.com/IBM/sarama"
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
}

// NewKafkaConsumer create new KafkaConsumer instance.
func NewKafkaConsumer(cfg IKafkaCfg, handler IHandlerMessage) (*KafkaConsumer, error) {
	// Consumer config
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.GetBrokers(), cfg.GetConsumerGroup(), config)
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
	defer kc.consumerGroup.Close()

	go func() {
		for err := range kc.consumerGroup.Errors() {
			log.Printf("Error from consumer group: %v", err)
		}
	}()

	log.Printf("Start consuming")

	// Consume messages from specified topics
	for {
		if err := kc.consumerGroup.Consume(ctx, kc.cfg.GetTopics(), kc); err != nil {
			log.Printf("Error during consumption: %v", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

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
	for message := range claim.Messages() {
		if err := kc.handler.HandleMessage(message); err != nil {
			log.Printf("Error handling message: %v", err)
			continue
		}
		sess.MarkMessage(message, "")
	}
	return nil
}
