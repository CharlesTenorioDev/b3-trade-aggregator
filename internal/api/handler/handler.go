package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
)

func GetAggregatedTradesHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instrumentCode := r.URL.Query().Get("ticker")
		if instrumentCode == "" {
			http.Error(w, "Parâmetro 'ticker' é obrigatório.", http.StatusBadRequest)
			return
		}

		startDateStr := r.URL.Query().Get("data_inicio")

		// Chama o serviço para obter os dados agregados
		aggregatedData, err := tradeService.RetrieveAggregatedData(r.Context(), instrumentCode, startDateStr)
		if err != nil {

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

func CreateTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

func GetTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

func UpdateTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}

func DeleteTradeHandler(tradeService service.TradeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "Endpoint não implementado", http.StatusNotImplemented)
	}
}
