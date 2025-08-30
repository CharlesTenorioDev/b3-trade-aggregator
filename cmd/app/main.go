package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/api/handler"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/ingestion"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/repository"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/pkg/server"
)

var (
	VERSION = "0.1.0-dev"
	COMMIT  = "ABCDEFG-dev"
)

func main() {
	log.Println("Starting B3 Trade Aggregator application")

	// Carrega configurações
	cfg := config.LoadConfig()

	// Inicializar pool de conexões com o banco de dados usando pgx
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Falha ao criar pool de conexões: %v", err)
	}
	defer pool.Close()

	// Testa a conexão
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalf("Falha ao pingar o banco de dados: %v", err)
	}
	log.Println("Pool de conexões PostgreSQL estabelecido com sucesso!")

	// Configurar dependências para Ingestão
	tradeReader := ingestion.NewTradeStreamReader()
	tradeRepo := repository.NewPostgresTradeRepository(pool)
	tradeService := service.NewTradeService(tradeReader, tradeRepo)

	// Exemplo de como você chamaria a ingestão.
	// Em um ambiente real, isso poderia ser um comando CLI separado ou um cron job.
	go func() {
		ingestionCtx, cancel := context.WithTimeout(context.Background(), 14*time.Minute) // Timeout de 14 minutos
		defer cancel()
		log.Println("Iniciando processo de ingestão...")
		// O caminho do arquivo deve vir de uma configuração ou argumento CLI
		filePath := "/path/to/your/29-08-2025_NEGOCIOSAVISTA.txt" // <-- AJUSTE ESTE CAMINHO
		if err := tradeService.ProcessIngestion(ingestionCtx, filePath); err != nil {
			log.Printf("Erro durante a ingestão: %v", err)
		} else {
			log.Println("Ingestão concluída com sucesso!")
		}
	}()

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
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server Run on [Port: %s], [Mode: %s], [Version: %s], [Commit: %s]", cfg.Port, cfg.Mode, VERSION, COMMIT)

	select {}
}
