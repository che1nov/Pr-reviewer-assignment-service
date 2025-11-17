package httpcontroller

import (
	"log/slog"
	"net/http"

	"github.com/che1nov/Pr-reviewer-assignment-service/internal/usecases"
)

type StatsHandler struct {
	logger          *slog.Logger
	getStatsUseCase *usecases.GetStatsUseCase
}

func NewStatsHandler(
	logger *slog.Logger,
	getStatsUseCase *usecases.GetStatsUseCase,
) *StatsHandler {
	return &StatsHandler{
		logger:          logger,
		getStatsUseCase: getStatsUseCase,
	}
}

// GetStats возвращает статистику по назначениям
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.getStatsUseCase.GetStats(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "ошибка получения статистики", "error", err)
		respondError(h.logger, w, http.StatusInternalServerError, ErrCodeInternal, ErrMsgInternalError)
		return
	}

	respondJSON(h.logger, w, http.StatusOK, stats)
}
