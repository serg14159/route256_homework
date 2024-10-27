package server

import (
	"context"
	"net/http"
	"route256/cart/internal/models"
	server_middleware "route256/cart/internal/pkg/mw/server"
	"route256/utils/logger"
)

type IConfig interface {
	GetHost() string
	GetPort() string
}

type ICartService interface {
	AddProduct(ctx context.Context, UID models.UID, SKU models.SKU, Count uint16) error
	DelProduct(ctx context.Context, UID models.UID, SKU models.SKU) error
	DelCart(ctx context.Context, UID models.UID) error
	GetCart(ctx context.Context, UID models.UID) ([]models.CartItemResponse, uint32, error)
	Checkout(ctx context.Context, UID models.UID) (int64, error)
}

type Server struct {
	server      *http.Server
	cfg         IConfig
	cartService ICartService
}

// NewServer function for create new server.
func NewServer(cfg IConfig, cartService ICartService) *Server {
	server := &http.Server{}
	return &Server{
		server:      server,
		cfg:         cfg,
		cartService: cartService,
	}
}

// Run function for running server.
func (s *Server) Run() error {
	// Set address
	address := s.cfg.GetHost() + ":" + s.cfg.GetPort()
	s.server.Addr = address

	// Set handler
	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/{user_id}/cart/{sku_id}", s.AddProduct)
	mux.HandleFunc("DELETE /user/{user_id}/cart/{sku_id}", s.DelProduct)
	mux.HandleFunc("DELETE /user/{user_id}/cart", s.DelCart)
	mux.HandleFunc("GET /user/{user_id}/cart", s.GetCart)
	mux.HandleFunc("POST /user/{user_id}/checkout", s.Checkout)

	s.server.Handler = server_middleware.New(mux)

	// Run goroutine with ListenAndServe
	go func() {
		ctx := context.Background()
		logger.Infow(ctx, "Server is running", "address", address)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorw(ctx, "Failed to listen", "error", err)
		}
	}()

	return nil
}

// Shutdown stop server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
