package handler

import (
	"net/http"

	"cyclecompass/api/internal/service"
)

type CycleHandler struct {
	svc *service.CycleService
}

func NewCycleHandler(svc *service.CycleService) *CycleHandler {
	return &CycleHandler{svc: svc}
}

func (h *CycleHandler) GetCycleOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=10")
	state := h.svc.GetCycleOverview(r.Context())
	writeSuccess(w, state)
}
