package handler

import (
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"
	"github.com/go-chi/chi/v5"
)

func RegisterTradeAPIHandlers(r *chi.Mux, tradeService service.TradeService) {

	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/trades", func(r chi.Router) {
			r.Get("/aggregated", GetAggregatedTradesHandler(tradeService))

		})
	})
}
