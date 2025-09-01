package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config/logger"
	"github.com/go-chi/chi/v5"
)

// Server define a interface para o servidor HTTP
type Server interface {
	Listen(ctx context.Context, wg *sync.WaitGroup)
}

type HTTPServer struct {
	router     *chi.Mux
	httpServer *http.Server
}

func NewHTTPServer(router *chi.Mux, cfg *config.Config) *HTTPServer {
	srv := &HTTPServer{
		router: router,
	}

	srv.httpServer = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second, // Default timeout
		WriteTimeout: 30 * time.Second, // Default timeout
	}

	return srv
}

func (s *HTTPServer) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *HTTPServer) Listen(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting HTTP server", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		logger.Info("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server forced to shutdown", err)
		}

		logger.Info("HTTP server exiting.")
	}()
}
