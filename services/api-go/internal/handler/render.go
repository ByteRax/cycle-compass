package handler

import (
	"encoding/json"
	"net/http"

	"cyclecompass/api/internal/dto"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeSuccess[T any](w http.ResponseWriter, data T) {
	writeJSON(w, http.StatusOK, dto.Success(data))
}

func writeFail(w http.ResponseWriter, status int, code int, message string) {
	writeJSON(w, status, dto.Fail(code, message))
}
