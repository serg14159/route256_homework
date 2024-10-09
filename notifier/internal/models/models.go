package models

import "time"

// User ID.
type UID = int64

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPayed           OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

type OrderEvent struct {
	OrderID    UID         `json:"order_id"`
	Status     OrderStatus `json:"status"`
	Time       time.Time   `json:"time"`
	Additional string      `json:"additional"`
}
