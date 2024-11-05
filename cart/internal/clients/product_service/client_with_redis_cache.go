package product_service

import (
	"context"
	"route256/cart/internal/models"
	"route256/cart/internal/pkg/metrics"
	"strconv"
	"time"

	"route256/utils/logger"

	"go.opentelemetry.io/otel"
	"golang.org/x/sync/singleflight"
)

// IRedisCacher
type IRedisCacher interface {
	Get(ctx context.Context, key models.SKU) (*models.GetProductResponse, bool, error)
	Set(ctx context.Context, key models.SKU, value *models.GetProductResponse) error
}

// ClientWithRedisCache
type ClientWithRedisCache struct {
	client *Client
	cacher IRedisCacher
	group  singleflight.Group
}

// NewClientWithRedisCache initializes a new ClientWithRedisCache instance.
func NewClientWithRedisCache(client *Client, cacher IRedisCacher) *ClientWithRedisCache {
	return &ClientWithRedisCache{
		client: client,
		cacher: cacher,
	}
}

// GetProduct retrieves a product by SKU, using the cache if available.
func (c *ClientWithRedisCache) GetProduct(ctx context.Context, SKU models.SKU) (*models.GetProductResponse, error) {
	// Tracer
	ctx, span := otel.Tracer("ProductServiceClientWithRedisCache").Start(ctx, "GetProduct")
	defer span.End()

	start := time.Now()

	// Check cache
	product, found, err := c.cacher.Get(ctx, SKU)
	if err != nil {
		logger.Errorw(ctx, "Cache Get error", "error", err)
	}

	if found {
		metrics.IncCacheHitCounter()
		metrics.ObserveCacheResponseTime("hit", time.Since(start))
		return product, nil
	}

	metrics.IncCacheMissCounter()

	v, err, _ := c.group.Do(strconv.FormatInt(int64(SKU), 10), func() (interface{}, error) {
		product, err := c.client.GetProduct(ctx, SKU)
		if err != nil {
			return nil, err
		}

		if err := c.cacher.Set(ctx, SKU, product); err != nil {
			logger.Errorw(ctx, "Cache Set error", "error", err)
		}

		metrics.ObserveCacheResponseTime("miss", time.Since(start))

		return product, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*models.GetProductResponse), nil
}
