package handler

import (
	"net/http"
	"strconv"

	"cyclecompass/api/internal/service"
)

type StrategyHandler struct {
	svc *service.StrategyService
}

func NewStrategyHandler(svc *service.StrategyService) *StrategyHandler {
	return &StrategyHandler{svc: svc}
}

func (h *StrategyHandler) GetRotationSignal(w http.ResponseWriter, r *http.Request) {
	sig := h.svc.GetRotationSignal(r.Context())
	writeSuccess(w, sig)
}

func (h *StrategyHandler) GetScoringDetail(w http.ResponseWriter, r *http.Request) {
	baseAmount := 200.0
	if raw := r.URL.Query().Get("base_amount"); raw != "" {
		if val, err := strconv.ParseFloat(raw, 64); err == nil && val > 0 {
			baseAmount = val
		}
	}

	detail := h.svc.GetScoringDetail(baseAmount)
	writeSuccess(w, detail)
}

func (h *StrategyHandler) GetCombinedSignal(w http.ResponseWriter, r *http.Request) {
	comb := h.svc.GetCombinedSignal(r.Context())
	writeSuccess(w, comb)
}
