package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
)

type IConfig interface {
	GetHost() string
	GetPort() string
}

type ICartService interface {
}

type Server struct {
	server      *http.Server
	cfg         IConfig
	cartService ICartService
}

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
	// Set host and port
	address := s.cfg.GetHost() + ":" + s.cfg.GetPort()
	s.server.Addr = address

	// Set handler

	// TO DO

	// Run goroutine with ListenAndServe
	go func() {
		log.Info().Msgf("Server is running on %s", address)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed listen")
		}
	}()

	return nil
}

// Shutdown stop server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
