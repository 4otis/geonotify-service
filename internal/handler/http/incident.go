package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/4otis/geonotify-service/internal/cases"
	dtoReq "github.com/4otis/geonotify-service/internal/dto/req"
	dtoResp "github.com/4otis/geonotify-service/internal/dto/resp"
	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/go-chi/chi"
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

	incidentID, err := h.incidentCase.CreateIncident(r.Context(), incident)
	if err != nil {
		h.logger.Error("incident create failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := dtoResp.IncidentCreateResponse{
		IncidentID: incidentID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// @Summary      Получить инцидент по ID (оператор)
// @Description  Детали конкретной зоны опасности
// @Tags         incidents
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer <api_key>"
// @Param        id             path    string  true  "ID инцидента"
// @Success      200 {object} IncidentResponse
// @Failure      401 {string} string "Не авторизован"
// @Failure      404 {string} string "Инцидент не найден"
// @Router       /api/v1/incidents/{id} [get]
func (h *IncidentHandler) IncidentGet(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "id required/not valid", http.StatusBadRequest)
		return
	}

	incident, err := h.incidentCase.ReadIncident(r.Context(), id)
	if err != nil {
		h.logger.Error("incident get failed",
			zap.Error(err),
			zap.Int("id", id))

		http.Error(w, "incident not found", http.StatusNotFound)
		return
	}

	response := dtoResp.IncidentResponse{
		IncidentID: incident.ID,
		Name:       incident.Name,
		Descr:      incident.Descr,
		Latitude:   incident.Latitude,
		Longitude:  incident.Longitude,
		Radius:     incident.Radius,
		IsActive:   incident.IsActive,
		CreatedAt:  incident.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
