package server

import (
	"context"
	"net/http"
	"route256/cart/internal/http/server_middleware"
	"route256/cart/internal/models"

	"log"

	"github.com/go-playground/validator"
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
}

type Server struct {
	server      *http.Server
	cfg         IConfig
	cartService ICartService
}

// Function for create new server.
func NewServer(cfg IConfig, cartService ICartService) *Server {
	server := &http.Server{}
	return &Server{
		server:      server,
		cfg:         cfg,
		cartService: cartService,
	}
}

// Function for running server.
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

	logMux := server_middleware.NewLogMux(mux)

	s.server.Handler = logMux

	// Run goroutine with ListenAndServe
	go func() {
		log.Printf("Server is running on %s", address)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed listen: %v", err)
		}
	}()

	return nil
}

// Shutdown stop server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}
