package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
)

// GetAggregatedTradesHandler returns an http.Handler for getting aggregated trade data
func GetAggregatedTradesHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instrumentCode := r.URL.Query().Get("ticker") // Parâmetro "ticker" do requisito
		if instrumentCode == "" {
			http.Error(w, "Parâmetro 'ticker' é obrigatório.", http.StatusBadRequest)
			return
		}

		startDateStr := r.URL.Query().Get("data_inicio") // Parâmetro "data_inicio" do requisito

		// Chama o serviço para obter os dados agregados
		aggregatedData, err := tradeService.RetrieveAggregatedData(r.Context(), instrumentCode, startDateStr)
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
		err = json.NewEncoder(w).Encode(aggregatedData)
		if err != nil {
			http.Error(w, "Erro ao converter resposta para JSON", http.StatusInternalServerError)
			return
		}
	}
}

// CreateTradeHandler returns an http.Handler for creating new trades
func CreateTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement trade creation logic
		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

// GetTradeHandler returns an http.Handler for getting a specific trade
func GetTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement trade retrieval logic
		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

// UpdateTradeHandler returns an http.Handler for updating a specific trade
func UpdateTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement trade update logic
		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

// DeleteTradeHandler returns an http.Handler for deleting a specific trade
func DeleteTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement trade deletion logic
		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}
