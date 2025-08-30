package entity

import "time"

// Trade representa uma negociação de ativo na B3.
type Trade struct {
	TradeDate          time.Time // Data em que a negociação ocorreu
	InstrumentCode     string    // Ticker do ativo, ex: PETR4
	NegotiatedPrice    float64   // Valor unitário do ativo. Considerar decimal.Decimal para precisão financeira se for crítico.
	NegotiatedQuantity int       // Quantidade de ativos negociados
	ClosingTime        string    // Formato "HHMMSSmmm"
}

// AggregatedData representa a estrutura de dados agregados de saída da API.
type AggregatedData struct {
	InstrumentCode string  `json:"ticker"`           // Código do instrumento (ticker)
	MaxRangeValue  float64 `json:"max_range_value"`  // Maior preço unitário no período
	MaxDailyVolume int     `json:"max_daily_volume"` // Volume máximo de negociações em um único dia
}
