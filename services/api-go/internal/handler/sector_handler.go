package handler

import (
	"net/http"

	"cyclecompass/api/internal/service"
)

type SectorHandler struct {
	svc *service.SectorService
}

func NewSectorHandler(svc *service.SectorService) *SectorHandler {
	return &SectorHandler{svc: svc}
}

func (h *SectorHandler) GetIndustryMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=15")
	metrics := h.svc.GetIndustryMetrics()
	writeSuccess(w, metrics)
}
