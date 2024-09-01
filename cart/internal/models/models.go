package models

// User ID.
type UID = int64

// Stock keeping unit.
type SKU = int64

// Struct for cart item.
type CartItem struct {
	SKU   SKU    `json:"sku_id"`
	Count uint16 `json:"count"`
}

// Struct for cart item response.
type CartItemResponse struct {
	SKU   SKU    `json:"sku_id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
	Count uint16 `json:"count"`
}

// Add product in user cart by SKU.
type AddProductRequest struct {
	Count uint16 `json:"count" validate:"required,min=1"`
}

type AddProductResponse struct {
}

// Del product from user cart by SKU.
type DelProductRequest struct {
}

type DelProductResponse struct {
}

// Del user cart.
type DelCartRequest struct {
}

type DelCartResponse struct {
}

// Get user cart.
type GetCartRequest struct {
}

type GetCartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice uint32             `json:"total_price"`
}

// Product service.
type GetProductRequest struct {
	Token string `json:"token"`
	SKU   uint32 `json:"sku"`
}

type GetProductResponse struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}
