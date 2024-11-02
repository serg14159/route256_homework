package cacher

import (
	"context"
	"encoding/json"
	"route256/cart/internal/models"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCacher
type RedisCacher struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacher initializes and returns a new RedisCacher instance.
func NewRedisCacher(client *redis.Client, ttl time.Duration) *RedisCacher {
	return &RedisCacher{
		client: client,
		ttl:    ttl,
	}
}

// Get retrieves a cached product response by SKU key.
func (c *RedisCacher) Get(ctx context.Context, key models.SKU) (*models.GetProductResponse, bool, error) {
	strKey := c.buildKey(key)
	val, err := c.client.Get(ctx, strKey).Result()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var product models.GetProductResponse
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return nil, false, err
	}

	return &product, true, nil
}

// Set stores a product response in Redis with the specified SKU key and TTL.
func (c *RedisCacher) Set(ctx context.Context, key models.SKU, value *models.GetProductResponse) error {
	strKey := c.buildKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, strKey, data, c.ttl).Err()
}

// buildKey generates a unique cache key string from the SKU.
func (c *RedisCacher) buildKey(key models.SKU) string {
	return "product:" + strconv.FormatInt(int64(key), 10)
}
