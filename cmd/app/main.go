package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/api/handler"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/ingestion"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/repository"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
)

func main() {
	// Carregar configurações
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

	// 1. Configurar dependências para Ingestão
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

	// 2. Configurar o router Chi
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second)) // Timeout para requisições HTTP

	apiHandler := handler.NewAPIHandler(tradeService)

	r.Get("/trades/aggregated", apiHandler.GetAggregatedTrades)

	log.Printf("Servidor API REST rodando na porta %s", cfg.APIPort)
	http.ListenAndServe(fmt.Sprintf(":%s", cfg.APIPort), r)
}
