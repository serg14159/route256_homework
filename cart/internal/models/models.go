package models

// User ID.
type UID = int64

// Stock keeping unit.
type SKU = int64

// Struct for cart item.
type CartItem struct {
	SKU   SKU    `json:"sku_id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
	Count uint16 `json:"count"`
}

// Add product in user cart by SKU.
type AddProductRequest struct {
	UID   UID    `json:"user_id"`
	SKU   SKU    `json:"sku_id"`
	Count uint16 `json:"count"`
}

type AddProductResponse struct {
}

// Del product from user cart by SKU.
type DelProductRequest struct {
	UID UID `json:"user_id"`
	SKU SKU `json:"sku_id"`
}

type DelProductResponse struct {
}

// Del user cart.
type DelCartRequest struct {
	UID UID `json:"user_id"`
}

type DelCartResponse struct {
}

// Get user cart.
type GetCartRequest struct {
	UID UID `json:"user_id"`
}

type GetCartResponse struct {
	Items      []CartItem `json:"items"`
	TotalPrice uint32     `json:"total_price"`
}
