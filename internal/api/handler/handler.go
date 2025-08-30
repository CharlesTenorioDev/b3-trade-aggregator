package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
)

// APIHandler provê métodos para lidar com requisições HTTP da API.
type APIHandler struct {
	tradeService service.TradeService
}

// NewAPIHandler cria uma nova instância de APIHandler.
func NewAPIHandler(svc service.TradeService) *APIHandler {
	return &APIHandler{tradeService: svc}
}

// GetAggregatedTrades lida com a requisição de busca de dados agregados de negociações.
func (h *APIHandler) GetAggregatedTrades(w http.ResponseWriter, r *http.Request) {
	instrumentCode := r.URL.Query().Get("ticker") // Parâmetro "ticker" do requisito
	if instrumentCode == "" {
		http.Error(w, "Parâmetro 'ticker' é obrigatório.", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("data_inicio") // Parâmetro "data_inicio" do requisito

	// Chama o serviço para obter os dados agregados
	aggregatedData, err := h.tradeService.RetrieveAggregatedData(r.Context(), instrumentCode, startDateStr)
	if err != nil {
		// Verifica se é um erro de "não encontrado" baseado na mensagem
		if strings.Contains(err.Error(), "dados não encontrados") {
			http.Error(w, "Dados não encontrados para o ticker e período especificados.", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Erro interno ao consultar dados: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aggregatedData)
}
