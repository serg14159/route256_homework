package product_service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"route256/cart/internal/models"
	"route256/cart/internal/pkg/metrics"
	client_middleware "route256/cart/internal/pkg/mw/client"
	"sync"
	"time"

	internal_errors "route256/cart/internal/pkg/errors"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
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

	transport := otelhttp.NewTransport(&http.Transport{})

	return &Client{
		cfg: cfg,
		client: &http.Client{
			Transport: &client_middleware.RetryMiddleware{
				MaxRetries: cfg.GetMaxRetries(),
				Transport:  transport,
			},
		},
		rateLimiter: rateLimiter,
	}
}

// GetProduct function for executes a request to the Product Service using a client with retries.
func (c *Client) GetProduct(ctx context.Context, SKU models.SKU) (product *models.GetProductResponse, err error) {
	// Tracer
	ctx, span := otel.Tracer("ProductServiceClient").Start(ctx, "GetProduct")
	defer span.End()

	if err = c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	var req *http.Request
	if req, err = c.prepareRequest(ctx, SKU); err != nil {
		return nil, err
	}

	// Start time for metrics
	start := time.Now()
	defer metrics.LogExternalRequest(req.URL.Path, start, &err)

	// Call client
	var resp *http.Response
	if resp, err = c.client.Do(req); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("request canceled: %w", err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	product, err = c.handleResponse(resp)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// prepareRequest
func (c *Client) prepareRequest(ctx context.Context, SKU models.SKU) (*http.Request, error) {
	reqBody := models.GetProductRequest{
		Token: c.cfg.GetToken(),
		SKU:   uint32(SKU),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", internal_errors.ErrInternalServerError)
	}

	uri := c.cfg.GetURI() + "/get_product"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", internal_errors.ErrInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// handleResponse
func (c *Client) handleResponse(resp *http.Response) (*models.GetProductResponse, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status code %d: %w", resp.StatusCode, internal_errors.ErrPreconditionFailed)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", internal_errors.ErrInternalServerError)
	}

	var product models.GetProductResponse
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", internal_errors.ErrInternalServerError)
	}

	return &product, nil
}
