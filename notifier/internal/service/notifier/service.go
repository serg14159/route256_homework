package service

import (
	"log"
	"route256/notifier/internal/models"
)

type NotifierService struct {
}

// NewNotifierService create new instance of NotifierService.
func NewNotifierService() *NotifierService {
	return &NotifierService{}
}

// ProcessOrderEvent function for processing OrderEvent.
func (s *NotifierService) ProcessOrderEvent(event *models.OrderEvent) error {
	log.Printf("Order event: %v", event)
	return nil
}
