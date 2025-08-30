package ingestion

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/entity" // Caminho do módulo ajustado
)

// TradeReader define a interface para leitura de stream de negociações.
type TradeReader interface {
	Read(ctx context.Context, path string) <-chan entity.Trade
}

// TradeStreamReader implementa TradeReader para arquivos de texto.
type TradeStreamReader struct{}

// NewTradeStreamReader cria uma nova instância de TradeStreamReader.
func NewTradeStreamReader() TradeReader {
	return &TradeStreamReader{}
}

// Read abre o arquivo em streaming, parseia cada linha em uma Trade
// e envia para um canal. Erros de parsing são logados.
func (c *TradeStreamReader) Read(ctx context.Context, path string) <-chan entity.Trade {
	tradeCh := make(chan entity.Trade)

	go func() {
		defer close(tradeCh) // Garante que o canal será fechado ao final da goroutine

		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro ao abrir arquivo: %v\n", err)
			return
		}
		defer file.Close()

		logFile, err := os.Create("errors.log") // Arquivo para log de erros de parsing
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro ao criar arquivo de log: %v\n", err)
			return
		}
		defer logFile.Close()

		scanner := bufio.NewScanner(file)

		// Pula o cabeçalho do arquivo, se houver
		if scanner.Scan() {
			// Pode-se logar ou ignorar a linha do cabeçalho
			// fmt.Printf("Cabeçalho: %s\n", scanner.Text())
		}

		for scanner.Scan() {
			select {
			case <-ctx.Done(): // Verifica se o contexto foi cancelado
				fmt.Println("pipeline cancelado pelo contexto")
				return
			default:
				line := scanner.Text()
				trade, err := parseTrade(line) // Tenta parsear a linha
				if err != nil {
					// Salva a linha com erro no log e continua o processamento
					logFile.WriteString(fmt.Sprintf("erro: %v | linha: %s\n", err, line))
					continue
				}
				tradeCh <- trade // Envia a trade parseada para o canal
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "erro ao ler arquivo: %v\n", err)
		}
	}()

	return tradeCh
}

// parseTrade transforma uma linha do arquivo em uma struct Trade.
// Ajustado para o formato exato das primeiras linhas do seu exemplo.
func parseTrade(line string) (entity.Trade, error) {
	parts := strings.Split(line, ";") // Delimitador é `;`
	// A linha de exemplo tem 11 colunas. Precisamos de pelo menos a coluna TradeDate (índice 8).
	if len(parts) < 9 {
		return entity.Trade{}, fmt.Errorf("linha inválida, esperado pelo menos 9 colunas, encontrada %d", len(parts))
	}

	// Mapeamento correto das colunas para a struct Trade
	// A DataNegocio da struct vem do campo 'DataNegocio' do arquivo (índice 8)
	tradeDate, err := time.Parse("2006-01-02", parts[8])
	if err != nil {
		return entity.Trade{}, fmt.Errorf("erro ao parsear TradeDate '%s': %w", parts[8], err)
	}

	// O NegotiatedPrice (PrecoNegocio) está na posição 3 e usa vírgula como separador decimal
	priceStr := strings.Replace(parts[3], ",", ".", 1) // Substitui vírgula por ponto para strconv.ParseFloat
	negotiatedPrice, err := parseFloat(priceStr)
	if err != nil {
		return entity.Trade{}, fmt.Errorf("erro ao parsear NegotiatedPrice '%s': %w", parts[3], err)
	}

	// NegotiatedQuantity (QuantidadeNegociada) está na posição 4
	negotiatedQuantity, err := parseInt(parts[4])
	if err != nil {
		return entity.Trade{}, fmt.Errorf("erro ao parsear NegotiatedQuantity '%s': %w", parts[4], err)
	}

	// ClosingTime (HoraFechamento) está na posição 5
	closingTime := parts[5]

	return entity.Trade{
		TradeDate:          tradeDate,
		InstrumentCode:     parts[1], // CodigoInstrumento está na posição 1
		NegotiatedPrice:    negotiatedPrice,
		NegotiatedQuantity: negotiatedQuantity,
		ClosingTime:        closingTime,
	}, nil
}

// parseFloat converte uma string para float64.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// parseInt converte uma string para int.
func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}
