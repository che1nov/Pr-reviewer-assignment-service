package httpcontroller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/dto"
)

func respondJSON(logger *slog.Logger, w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("не удалось закодировать ответ", "error", err)
	}
}

func respondError(logger *slog.Logger, w http.ResponseWriter, status int, code, message string) {
	respondJSON(logger, w, status, dto.ErrorResponse{
		Error: dto.ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func respondBadRequest(logger *slog.Logger, r *http.Request, w http.ResponseWriter, code, message string, err error) {
	if err != nil {
		logger.WarnContext(r.Context(), "ошибка декодирования запроса", "error", err)
	} else {
		logger.WarnContext(r.Context(), "ошибка валидации запроса", "message", message)
	}
	respondError(logger, w, http.StatusBadRequest, code, message)
}
