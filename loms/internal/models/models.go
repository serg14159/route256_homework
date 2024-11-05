package models

import "time"

// User ID.
type UID = int64

// Stock keeping unit.
type SKU = uint32

// Order ID.
type OID = int64

// Item represents single item in an order.
type Item struct {
	SKU   SKU    `validate:"gt=0"`
	Count uint16 `validate:"gt=0"`
}

// Stock represents inventory information for a specific product (SKU).
type Stock struct {
	SKU        SKU    `json:"sku"`
	TotalCount uint64 `json:"total_count"`
	Reserved   uint64 `json:"reserved"`
}

// OrderStatus represents status of order.
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPayed           OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// Order represents order containing items.
type Order struct {
	OrderID OID
	Status  OrderStatus
	UserID  int64
	Items   []Item
}

// OrderCreateRequest represents a request to create an order.
type OrderCreateRequest struct {
	User  UID    `validate:"gt=0"`
	Items []Item `validate:"required,dive"`
}

// OrderCreateResponse represents a response after creating an order.
type OrderCreateResponse struct {
	OrderID OID `json:"orderID"`
}

// OrderInfoRequest represents a request for information about an order.
type OrderInfoRequest struct {
	OrderID OID `validate:"gt=0"`
}

// OrderInfoResponse represents a response containing order information.
type OrderInfoResponse struct {
	Status OrderStatus `json:"status"`
	User   UID         `validate:"gt=0"`
	Items  []Item      `json:"items"`
}

// OrderPayRequest represents a request to pay for an order.
type OrderPayRequest struct {
	OrderID OID `validate:"gt=0"`
}

// OrderPayResponse represents a response after paying for an order.
type OrderPayResponse struct{}

// OrderCancelRequest represents a request to cancel an order.
type OrderCancelRequest struct {
	OrderID OID `validate:"gt=0"`
}

// OrderCancelResponse represents a response after canceling an order.
type OrderCancelResponse struct{}

// StocksInfoRequest represents a request for stock information.
type StocksInfoRequest struct {
	SKU SKU `validate:"gt=0"`
}

// StocksInfoResponse represents a response containing stock information.
type StocksInfoResponse struct {
	Count uint64 `json:"count"`
}

// OrderEvent represents kafka message.
type OrderEvent struct {
	OrderID    UID         `json:"order_id"`
	Status     OrderStatus `json:"status"`
	Time       time.Time   `json:"time"`
	Additional string      `json:"additional"`
}

// OutboxEvent represents event stored in the outbox table.
type OutboxEvent struct {
	ID        int64  `json:"id"`
	EventType string `json:"event_type"`
	Payload   string `json:"payload"`
}
