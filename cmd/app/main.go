package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/api/handler"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config/logger"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/repository"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/pkg/server"
	"go.uber.org/zap"
)

var (
	VERSION = "0.1.0-dev"
	COMMIT  = "ABCDEFG-dev"
)

func main() {
	logger.Info("Starting B3 Trade Aggregator Web Application",
		zap.String("version", VERSION),
		zap.String("commit", COMMIT))

	// Carrega configurações
	cfg := config.LoadConfig()

	// Inicializar pool de conexões com o banco de dados usando pgx
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("Falha ao criar pool de conexões", err, zap.String("database_url", cfg.DatabaseURL))
		return
	}
	defer pool.Close()

	// Testa a conexão
	if err = pool.Ping(context.Background()); err != nil {
		logger.Error("Falha ao pingar o banco de dados", err)
		return
	}
	logger.Info("Pool de conexões PostgreSQL estabelecido com sucesso!")

	// Configurar dependências para consultas (sem ingestão)
	tradeRepo := repository.NewPostgresTradeRepository(pool)
	tradeService := service.NewTradeService(nil, tradeRepo) // No reader needed for queries only

	// Criação do router com Chi
	r := chi.NewRouter()

	// Middleware básico
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Configure CORS
	corsMiddleware := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(corsMiddleware)

	// Healthcheck básico
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"MSG":"Server Ok","codigo":200}`)
	})

	// Registra handlers do módulo trade
	handler.RegisterTradeAPIHandlers(r, tradeService)

	// Create an HTTP server
	srv := server.NewHTTPServer(r, cfg)

	// Start the server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", err)
			return
		}
	}()

	logger.Info("Server started successfully",
		zap.String("port", cfg.Port),
		zap.String("mode", cfg.Mode),
		zap.String("version", VERSION),
		zap.String("commit", COMMIT))

	select {}
}
