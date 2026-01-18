package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/4otis/geonotify-service/internal/cases"
	dtoResp "github.com/4otis/geonotify-service/internal/dto/resp"
	"github.com/4otis/geonotify-service/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger *zap.Logger
	dbPool *pgxpool.Pool
	redis  *redis.Client
	uc     cases.StatsUseCase
}

func NewHealthHandler(logger *zap.Logger, dbPool *pgxpool.Pool, redis *redis.Client, uc cases.StatsUseCase) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		dbPool: dbPool,
		redis:  redis,
		uc:     uc,
	}
}

// HealthCheck обрабатывает GET /api/v1/system/health
// @Summary      Health check
// @Description  Проверка состояния системы
// @Tags         system
// @Produce      json
// @Success      200 {object} dtoResp.HealthResponse
// @Failure      500 {object} dtoResp.HealthResponse
// @Router       /api/v1/system/health [get]
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := "healthy"
	httpStatus := http.StatusOK

	dbHealthy := true
	if err := h.dbPool.Ping(ctx); err != nil {
		h.logger.Error("database health check failed", zap.Error(err))
		dbHealthy = false
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	redisHealthy := true
	if err := h.redis.HealthCheck(); err != nil {
		h.logger.Error("redis health check failed", zap.Error(err))
		redisHealthy = false
		if status == "healthy" {
			status = "degraded"
			httpStatus = http.StatusPartialContent
		}
	}

	var activeIncidents, inProgressWebhooks int
	var metricsErr error

	if dbHealthy {
		activeIncidents, metricsErr = h.uc.GetActiveIncidentsCount(ctx)
		if metricsErr != nil {
			h.logger.Warn("failed to get active incidents count", zap.Error(metricsErr))
		}

		inProgressWebhooks, metricsErr = h.uc.GetPendingWebhooksCount(ctx)
		if metricsErr != nil {
			h.logger.Warn("failed to get in-progress webhooks count", zap.Error(metricsErr))
		}
	}

	response := dtoResp.HealthResponse{
		Status:          status,
		Timestamp:       time.Now().UTC(),
		ActiveIncidents: activeIncidents,
		PendingWebhooks: inProgressWebhooks,
	}

	responseWithDetails := struct {
		dtoResp.HealthResponse
		Components struct {
			Database string `json:"database"`
			Redis    string `json:"redis"`
		} `json:"components,omitempty"`
	}{
		HealthResponse: response,
	}

	responseWithDetails.Components.Database = "healthy"
	if !dbHealthy {
		responseWithDetails.Components.Database = "unhealthy"
	}

	responseWithDetails.Components.Redis = "healthy"
	if !redisHealthy {
		responseWithDetails.Components.Redis = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(responseWithDetails); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
		w.Write([]byte(`{"status":"error","message":"failed to encode response"}`))
	}
}
