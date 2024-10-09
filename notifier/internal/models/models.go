package models

import "time"

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPayed           OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

type OrderEvent struct {
	OrderID    string      `json:"order_id"`
	Status     OrderStatus `json:"status"`
	Time       time.Time   `json:"time"`
	Additional string      `json:"additional"`
}
