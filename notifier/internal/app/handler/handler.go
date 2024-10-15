package handler

import (
	"encoding/json"

	"route256/notifier/internal/models"

	"github.com/IBM/sarama"
)

type INotifierService interface {
	ProcessOrderEvent(event *models.OrderEvent) error
}

type MessageHandler struct {
	notifierService INotifierService
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(service INotifierService) *MessageHandler {
	return &MessageHandler{
		notifierService: service,
	}
}

// HandleMessage processes the incoming Kafka message.
func (h *MessageHandler) HandleMessage(msg *sarama.ConsumerMessage) error {
	var event models.OrderEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	if err := h.notifierService.ProcessOrderEvent(&event); err != nil {
		return err
	}

	return nil
}
