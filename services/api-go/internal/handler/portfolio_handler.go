package handler

import (
	"net/http"

	"cyclecompass/api/internal/service"
)

type PortfolioHandler struct {
	svc *service.PortfolioService
}

func NewPortfolioHandler(svc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

func (h *PortfolioHandler) GetPortfolioOverview(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		accountID = "ACC_DEMO_001"
	}

	overview, err := h.svc.GetPortfolioOverview(r.Context(), accountID)
	if err != nil {
		writeFail(w, http.StatusInternalServerError, 500, "获取账户持仓失败: "+err.Error())
		return
	}
	writeSuccess(w, overview)
}
