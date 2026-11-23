package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"cyclecompass/api/internal/service"
)

type MarketHandler struct {
	svc *service.MarketService
}

func NewMarketHandler(svc *service.MarketService) *MarketHandler {
	return &MarketHandler{svc: svc}
}

func (h *MarketHandler) GetQuotes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	symbols := r.URL.Query().Get("symbols")

	quotes := h.svc.GetQuotes(r.Context(), category, symbols)
	writeSuccess(w, quotes)
}

func (h *MarketHandler) GetSingleQuote(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		writeFail(w, http.StatusBadRequest, 400, "缺少标的代码")
		return
	}

	quote := h.svc.GetSingleQuote(r.Context(), symbol)
	writeSuccess(w, quote)
}
