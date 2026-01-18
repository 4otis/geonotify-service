package http

import (
	"encoding/json"
	"net/http"

	"github.com/4otis/geonotify-service/internal/cases"
	dtoResp "github.com/4otis/geonotify-service/internal/dto/resp"
	"go.uber.org/zap"
)

type StatsHandler struct {
	logger    *zap.Logger
	uc        cases.StatsUseCase
	windowMin int
}

func NewStatsHandler(logger *zap.Logger, uc cases.StatsUseCase, windowMin int) *StatsHandler {
	return &StatsHandler{
		logger:    logger,
		uc:        uc,
		windowMin: windowMin,
	}
}

// GetStats обрабатывает GET /api/v1/incidents/stats
// @Summary      Статистика по зонам
// @Description  Получить статистику уникальных пользователей за последние N минут
// @Tags         stats
// @Produce      json
// @Success      200 {object} dtoResp.StatsResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/incidents/stats [get]
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userCount, totalChecks, periodStart, err := h.uc.GetStats(r.Context(), h.windowMin)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "failed to retrieve statistics")
		return
	}

	response := dtoResp.StatsResponse{
		UserCount:     userCount,
		TotalChecks:   totalChecks,
		WindowMinutes: h.windowMin,
		PeriodStart:   periodStart,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *StatsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.Error("failed to encode error response", zap.Error(err))
	}
}

func (h *StatsHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "failed to encode response"}`))
	}
}
