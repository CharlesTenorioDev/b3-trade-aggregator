package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/entity"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/ingestion"
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/repository"
)

// TradeService define a interface para as operações de negócio de negociações.
type TradeService interface {
	ProcessIngestion(ctx context.Context, filePath string) error
	RetrieveAggregatedData(ctx context.Context, instrumentCode string, startDateStr string) (*entity.AggregatedData, error)
}

type tradeServiceImpl struct {
	tradeReader ingestion.TradeReader
	tradeRepo   repository.TradeRepository
}

func NewTradeService(reader ingestion.TradeReader, repo repository.TradeRepository) TradeService {
	return &tradeServiceImpl{
		tradeReader: reader,
		tradeRepo:   repo,
	}
}

// ProcessIngestion orquestra o pipeline de leitura, parsing e persistência.
func (s *tradeServiceImpl) ProcessIngestion(ctx context.Context, filePath string) error {
	const (
		numWorkers = 4    // Número de goroutines para processar e salvar (ajustável)
		batchSize  = 1000 // Tamanho do lote para inserção no banco de dados
	)

	tradeCh := s.tradeReader.Read(ctx, filePath) // Inicia a leitura do arquivo

	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers) // Canal para coletar erros dos workers

	// Iniciar workers consumidores
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			batch := make([]entity.Trade, 0, batchSize)
			for trade := range tradeCh {
				batch = append(batch, trade)
				if len(batch) >= batchSize {
					// Salvar lote no banco de dados
					err := s.tradeRepo.SaveTrades(ctx, batch)
					if err != nil {
						errCh <- fmt.Errorf("worker %d: falha ao salvar lote: %w", workerID, err)
						return // Termina o worker em caso de erro fatal de DB
					}
					batch = make([]entity.Trade, 0, batchSize) // Reseta o lote
				}
			}
			// Salvar qualquer lote restante no final do canal
			if len(batch) > 0 {
				err := s.tradeRepo.SaveTrades(ctx, batch)
				if err != nil {
					errCh <- fmt.Errorf("worker %d: falha ao salvar lote final: %w", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()    // Espera todos os workers terminarem
	close(errCh) // Fecha o canal de erros após todos os workers terminarem

	// Verifica se houve algum erro nos workers
	for err := range errCh {
		return err // Retorna o primeiro erro encontrado
	}

	return nil
}

// RetrieveAggregatedData obtém dados agregados e aplica a lógica de cálculo da data.
func (s *tradeServiceImpl) RetrieveAggregatedData(ctx context.Context, instrumentCode string, startDateStr string) (*entity.AggregatedData, error) {
	var startDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr) // Formato ISO-8601
		if err != nil {
			return nil, fmt.Errorf("service: formato de 'data_inicio' inválido. Use YYYY-MM-DD: %w", err)
		}
	} else {
		// Se data_inicio for omitido, a consulta deve abranger os últimos 7 dias úteis,
		// tendo como último dia do período de análise a data anterior à atual.
		// Exemplo simplificado: hoje - 8 dias para cobrir 7 dias úteis "completos".
		// Uma implementação mais robusta envolveria um calendário de dias úteis da B3.
		startDate = time.Now().AddDate(0, 0, -8) // Ajuste conforme a regra exata de "dias úteis"
	}

	data, err := s.tradeRepo.GetAggregatedData(ctx, instrumentCode, startDate)
	if err != nil {
		return nil, fmt.Errorf("service: falha ao buscar dados agregados: %w", err)
	}
	return data, nil
}
