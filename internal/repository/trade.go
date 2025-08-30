// internal/repository/trade.go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/entity"
)

// TradeRepository define a interface para operações de banco de dados relacionadas a trades.
type TradeRepository interface {
	SaveTrades(ctx context.Context, trades []entity.Trade) error
	GetAggregatedData(ctx context.Context, instrumentCode string, startDate time.Time) (*entity.AggregatedData, error)
}

type postgresTradeRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTradeRepository cria uma nova instância de postgresTradeRepository.
func NewPostgresTradeRepository(pool *pgxpool.Pool) TradeRepository {
	return &postgresTradeRepository{pool: pool}
}

// SaveTrades persiste um lote de negociações no banco de dados usando COPY FROM do pgx.
// Esta implementação é muito mais performática para ingestão em massa (565MB).
func (r *postgresTradeRepository) SaveTrades(ctx context.Context, trades []entity.Trade) error {
	if len(trades) == 0 {
		return nil
	}

	// Inicia uma transação
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: falha ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepara o COPY FROM
	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"trades"},
		[]string{"trade_date", "instrument_code", "negotiated_price", "negotiated_quantity", "closing_time"},
		pgx.CopyFromRows(r.tradesToRows(trades)),
	)
	if err != nil {
		return fmt.Errorf("repository: falha ao executar COPY FROM: %w", err)
	}

	// Verifica se o número de linhas copiadas corresponde ao esperado
	if copyCount != int64(len(trades)) {
		return fmt.Errorf("repository: número de linhas copiadas (%d) não corresponde ao esperado (%d)", copyCount, len(trades))
	}

	// Commit da transação
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository: falha ao fazer commit da transação: %w", err)
	}

	return nil
}

// tradesToRows converte um slice de Trade para um slice de []interface{} para o COPY FROM.
func (r *postgresTradeRepository) tradesToRows(trades []entity.Trade) [][]interface{} {
	rows := make([][]interface{}, len(trades))
	for i, trade := range trades {
		rows[i] = []interface{}{
			trade.TradeDate,
			trade.InstrumentCode,
			trade.NegotiatedPrice,
			trade.NegotiatedQuantity,
			trade.ClosingTime,
		}
	}
	return rows
}

// GetAggregatedData busca dados agregados para um ticker e período.
func (r *postgresTradeRepository) GetAggregatedData(ctx context.Context, instrumentCode string, startDate time.Time) (*entity.AggregatedData, error) {
	query := `
        SELECT
            MAX(t.negotiated_price) AS max_range_value,
            MAX(daily_volume.total_volume) AS max_daily_volume,
            $1 AS instrument_code
        FROM
            trades t
        JOIN (
            SELECT
                trade_date,
                SUM(negotiated_quantity) AS total_volume
            FROM
                trades
            WHERE
                instrument_code = $1 AND trade_date >= $2
            GROUP BY
                trade_date
        ) AS daily_volume ON t.trade_date = daily_volume.trade_date AND t.instrument_code = $1
        WHERE
            t.instrument_code = $1 AND t.trade_date >= $2;
    `

	var result entity.AggregatedData
	err := r.pool.QueryRow(ctx, query, instrumentCode, startDate).Scan(
		&result.MaxRangeValue,
		&result.MaxDailyVolume,
		&result.InstrumentCode,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("repository: dados não encontrados")
		}
		return nil, fmt.Errorf("repository: falha ao obter dados agregados: %w", err)
	}

	return &result, nil
}
