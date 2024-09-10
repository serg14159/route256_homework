package suite

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"route256/cart/internal/app/server"
	"route256/cart/internal/models"
	"route256/cart/internal/pkg/cart/repository"
	"route256/cart/internal/pkg/cart/service"
	"route256/cart/internal/pkg/clients/product_service"
	internal_errors "route256/cart/internal/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Struct Config for product service client.
type Config struct{}

func (c *Config) GetHost() string {
	return "localhost"
}

func (c *Config) GetPort() string {
	return "8082"
}

func (c *Config) GetURI() string {
	return "http://route256.pavl.uk:8080"
}

func (c *Config) GetToken() string {
	return "testtoken"
}

func (c *Config) GetMaxRetries() int {
	return 3
}

// Struct TSuite for tests.
type TSuite struct {
	suite.Suite
	repo           *repository.Repository
	service        *service.CartService
	server         *server.Server
	productService *product_service.Client
	serverURL      string
	cancelFunc     context.CancelFunc
}

// Function SetupTest.
func (s *TSuite) SetupTest() {
	// Repository
	s.repo = repository.NewCartRepository()

	// Product service
	clientCfg := &Config{}
	s.productService = product_service.NewClient(clientCfg)

	// Cart service.
	s.service = service.NewService(s.repo, s.productService)

	// Server configuration
	cfg := &Config{}

	// Server
	s.server = server.NewServer(cfg, s.service)

	// Create context for cancel
	_, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Run server
	go func() {
		err := s.server.Run()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait
	time.Sleep(100 * time.Millisecond)

	// Create URL
	s.serverURL = "http://" + cfg.GetHost() + ":" + cfg.GetPort()
}

// Function TearDownTest stop server after test.
func (s *TSuite) TearDownTest() {
	// Stop server.
	s.cancelFunc()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	require.NoError(s.T(), err)
}

// Function TestDeleteProductFromCart.
func (s *TSuite) TestDeleteProductFromCart() {
	ctx := context.Background()
	// Add item to cart
	UID := models.UID(31337)
	SKU := models.SKU(1076963)
	err := s.repo.AddItem(ctx, UID, models.CartItem{SKU: SKU, Count: 1})
	require.NoError(s.T(), err)

	// Send http request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, s.serverURL+"/user/"+strconv.FormatInt(int64(UID), 10)+"/cart/"+strconv.FormatInt(int64(SKU), 10), nil)
	require.NoError(s.T(), err)

	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)

	// Check delete item
	_, err = s.repo.GetItemsByUserID(ctx, UID)

	if err != nil {
		if errors.Is(err, internal_errors.ErrNotFound) {
			return
		}
		require.NoError(s.T(), err)
	}
}

// Function TestGetCart
func (s *TSuite) TestGetCart() {
	// Add item to cart
	UID := models.UID(31337)
	firstSKU := models.SKU(1076963)
	secondSKU := models.SKU(1148162)
	err := s.repo.AddItem(context.TODO(), UID, models.CartItem{SKU: firstSKU, Count: 2})
	require.NoError(s.T(), err)
	err = s.repo.AddItem(context.TODO(), UID, models.CartItem{SKU: secondSKU, Count: 3})
	require.NoError(s.T(), err)

	// Send http request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, s.serverURL+"/user/"+strconv.FormatInt(int64(UID), 10)+"/cart", nil)
	require.NoError(s.T(), err)

	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Check get item
	var res models.GetCartResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	require.NoError(s.T(), err)
	assert.Len(s.T(), res.Items, 2)
	assert.Equal(s.T(), uint32(15551), res.TotalPrice) // 2*3379 + 3*2931 = 15551
}
