package models

// User ID.
type UID = int64

// Stock keeping unit.
type SKU = uint32

// Order ID.
type OID = int64

// Item represents a single item in an order.
type Item struct {
	SKU   SKU    `validate:"gt=0"`
	Count uint16 `validate:"gt=0"`
}

// Stock represents the inventory information for a specific product (SKU).
type Stock struct {
	SKU        SKU    `json:"sku"`
	TotalCount uint64 `json:"total_count"`
	Reserved   uint64 `json:"reserved"`
}

// OrderStatus represents the status of an order.
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
	Status OrderStatus
	UserID int64
	Items  []Item
}

// OrderCreateRequest represents a request to create an order.
type OrderCreateRequest struct {
	User  UID    `validate:"gt=0"`
	Items []Item `validate:"required,dive"`
}

// OrderCreateResponse represents a response after creating an order.
type OrderCreateResponse struct {
	OrderID int64 `json:"orderID"`
}

// OrderInfoRequest represents a request for information about an order.
type OrderInfoRequest struct {
	OrderID int64 `validate:"gt=0"`
}

// OrderInfoResponse represents a response containing order information.
type OrderInfoResponse struct {
	Status string `json:"status"`
	User   UID    `validate:"gt=0"`
	Items  []Item `json:"items"`
}

// OrderPayRequest represents a request to pay for an order.
type OrderPayRequest struct {
	OrderID int64 `validate:"gt=0"`
}

// OrderPayResponse represents a response after paying for an order.
type OrderPayResponse struct{}

// OrderCancelRequest represents a request to cancel an order.
type OrderCancelRequest struct {
	OrderID int64 `validate:"gt=0"`
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
