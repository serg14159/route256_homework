package product_service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"route256/cart/internal/http/client_middleware"
	"route256/cart/internal/models"
	"sync"
	"time"

	internal_errors "route256/cart/internal/pkg/errors"
)

const (
	RateLimiterTokens   = 10
	RateLimiterInterval = time.Second
)

type RateLimiter struct {
	tokens chan struct{}
	mu     sync.Mutex
}

func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, rate),
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
		loop:
			for i := 0; i < rate; i++ {
				select {
				case rl.tokens <- struct{}{}:
				default:
					break loop
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

type IConfig interface {
	GetURI() string
	GetToken() string
	GetMaxRetries() int
}

type Client struct {
	client      *http.Client
	cfg         IConfig
	rateLimiter *RateLimiter
}

// NewClient function for creates a new client.
func NewClient(cfg IConfig) *Client {
	rateLimiter := NewRateLimiter(RateLimiterTokens, RateLimiterInterval)
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Transport: &client_middleware.RetryMiddleware{
				MaxRetries: cfg.GetMaxRetries(),
				Transport:  http.DefaultTransport,
			},
		},
		rateLimiter: rateLimiter,
	}
}

// GetProduct function for executes a request to the Product Service using a client with retries.
func (c *Client) GetProduct(ctx context.Context, SKU models.SKU) (*models.GetProductResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	reqBody := models.GetProductRequest{
		Token: c.cfg.GetToken(),
		SKU:   uint32(SKU),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w, %w", err, internal_errors.ErrInternalServerError)
	}

	uri := c.cfg.GetURI() + "/get_product"

	req, err := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w, %w", err, internal_errors.ErrInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("request canceled: %w", err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service status code: %d, invalid sku, %w", resp.StatusCode, internal_errors.ErrPreconditionFailed)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w, %w", err, internal_errors.ErrInternalServerError)
	}

	var product models.GetProductResponse
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w, %w", err, internal_errors.ErrInternalServerError)
	}

	return &product, nil
}
