package product_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"route256/cart/internal/http/client_middleware"
	"route256/cart/internal/models"
)

type IConfig interface {
	GetURI() string
	GetToken() string
}

type Client struct {
	client *http.Client
	cfg    IConfig
}

// Function for creates a new client.
func NewClient(cfg IConfig) *Client {
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Transport: &client_middleware.RetryMiddleware{
				MaxRetries: 3,
				Transport:  http.DefaultTransport,
			},
		},
	}
}

// Function for executes a request to the Product Service using a client with retries.
func (c *Client) GetProduct(SKU models.SKU) (*models.GetProductResponse, error) {
	reqBody := models.GetProductRequest{
		Token: c.cfg.GetToken(),
		SKU:   uint32(SKU),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	uri := c.cfg.GetURI() + "/get_product"

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var product models.GetProductResponse
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &product, nil
}
