package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/service"
)

type BacktestHandler struct {
	svc *service.BacktestService
}

func NewBacktestHandler(svc *service.BacktestService) *BacktestHandler {
	return &BacktestHandler{svc: svc}
}

func (h *BacktestHandler) RunBacktest(w http.ResponseWriter, r *http.Request) {
	req := dto.BacktestRequest{
		LowLine:      0.16,
		HighLine:     0.28,
		WeeklyAmount: 1000.0,
		Days:         1500,
	}

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeFail(w, http.StatusBadRequest, 400, "请求体 JSON 格式非法")
			return
		}
	}

	rawJSON, err := h.svc.RunBacktestRaw(req)
	if err != nil {
		writeFail(w, http.StatusUnprocessableEntity, 422, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(rawJSON)
}
