package handler

import (
	"net/http"
	"time"

	"cyclecompass/api/internal/config"
	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

type RootHandler struct {
	cfg      *config.Config
	provider *market.OnlineProvider
}

func NewRootHandler(cfg *config.Config, provider *market.OnlineProvider) *RootHandler {
	return &RootHandler{
		cfg:      cfg,
		provider: provider,
	}
}

func (h *RootHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	resp := dto.RootResponse{
		Name:    h.cfg.AppName,
		Version: h.cfg.AppVersion,
		Docs:    "/docs",
		Health:  "/api/v1/health",
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *RootHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dQuote := h.provider.GetQuote(ctx, "930955")
	gQuote := h.provider.GetQuote(ctx, "399673")

	data := dto.HealthData{
		Status:    "healthy",
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000000"),
		Services: map[string]string{
			"api":                "ok",
			"strategy_engine":    "ok",
			"backtesting_engine": "ok",
			"market_data":        gQuote.Meta.Status,
		},
		MonitoredSources: map[string]dto.HealthSourceMeta{
			"dividend_930955": {
				Source: dQuote.Meta.SourceName,
				AsOf:   dQuote.Meta.AsOf,
				Status: dQuote.Meta.Status,
			},
			"growth_399673": {
				Source: gQuote.Meta.SourceName,
				AsOf:   gQuote.Meta.AsOf,
				Status: gQuote.Meta.Status,
			},
		},
	}
	writeSuccess(w, data)
}
