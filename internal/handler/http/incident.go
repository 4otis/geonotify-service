package http

import (
	"encoding/json"
	"net/http"

	"github.com/4otis/geonotify-service/internal/cases"
	dtoReq "github.com/4otis/geonotify-service/internal/dto/req"
	dtoResp "github.com/4otis/geonotify-service/internal/dto/resp"
	"github.com/4otis/geonotify-service/internal/entity"
	"go.uber.org/zap"
)

type IncidentHandler struct {
	logger       *zap.Logger
	apiKey       string
	incidentCase cases.IncidentUseCase
}

func NewIncidentHandler(logger *zap.Logger, apiKey string,
	incidentCase cases.IncidentUseCase) *IncidentHandler {
	return &IncidentHandler{
		logger:       logger,
		apiKey:       apiKey,
		incidentCase: incidentCase,
	}
}

// @Summary      Создать инцидент (оператор)
// @Description  Создать новую опасную зону (требуется API key)
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header    string  true  "Bearer <api_key>"
// @Param        request        body      dtoReq.IncidentCreateRequest  true  "Данные инцидента"
// @Success      201            {object}  IncidentCreateResponse
// @Failure      400            {object}  ErrorResponse  "Неверный формат данных"
// @Failure      401            {object}  ErrorResponse  "Не авторизован"
// @Failure      500            {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents [post]
func (h *IncidentHandler) IncidentCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dtoReq.IncidentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		http.Error(w, "invalid coordinates", http.StatusBadRequest)
		return
	}
	if req.Radius <= 0 {
		http.Error(w, "radius_m must be > 0", http.StatusBadRequest)
		return
	}

	incident := entity.Incident{
		Name:      req.Name,
		Descr:     req.Descr,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Radius:    req.Radius,
	}

	incidentID, err := h.incidentCase.Create(r.Context(), incident)
	if err != nil {
		h.logger.Error("incident create failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := dtoResp.IncidentCreateResponse{
		IncidentID: incidentID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
