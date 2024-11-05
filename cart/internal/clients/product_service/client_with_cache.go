package product_service

import (
	"context"
	"route256/cart/internal/models"

	"go.opentelemetry.io/otel"
)

type ICacher interface {
	Get(key models.SKU) (*models.GetProductResponse, bool)
	Set(key models.SKU, value *models.GetProductResponse)
}

// ClientWithCache represents a ProductService client with cache.
type ClientWithCache struct {
	client *Client
	cacher ICacher
}

// NewClientWithCache creates a new ClientWithCache.
func NewClientWithCache(client *Client, cacher ICacher) *ClientWithCache {
	return &ClientWithCache{
		client: client,
		cacher: cacher,
	}
}

// GetProduct retrieves product information, using cache if available.
func (c *ClientWithCache) GetProduct(ctx context.Context, SKU models.SKU) (*models.GetProductResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("ProductServiceClientWithCache").Start(ctx, "GetProduct")
	defer span.End()

	// Check cache
	if value, ok := c.cacher.Get(SKU); ok {
		return value, nil
	}

	// If not in cache, call client
	product, err := c.client.GetProduct(ctx, SKU)
	if err != nil {
		return nil, err
	}

	// Save in cache
	c.cacher.Set(SKU, product)

	return product, nil
}
