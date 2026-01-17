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
// @Security     ApiKeyAuth
// @Param        request        body      dtoReq.IncidentCreateRequest  true  "Данные инцидента"
// @Success      201            {object}  dtoResp.IncidentCreateResponse
// @Failure      400            {string}  string  "Неверный формат данных"
// @Failure      401            {string}  string  "Не авторизован"
// @Failure      500            {string}  string  "Внутренняя ошибка сервера"
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
// @Security     ApiKeyAuth
// @Param        id             path    string  true  "ID инцидента"
// @Success      200 {object} dtoResp.IncidentResponse
// @Failure      401 {string} string "Не авторизован"
// @Failure      404 {string} string "Инцидент не найден"
// @Router       /api/v1/incidents/{incident_id} [get]
func (h *IncidentHandler) IncidentGet(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "incident_id"))
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
		UpdatedAt:  incident.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// @Summary      Получить список инцидентов с пагинацией (оператор)
// @Description  Получить все инциденты с поддержкой пагинации
// @Tags         incidents
// @Produce      json
// @Security     ApiKeyAuth
// @Param        page           query     int     false  "Номер страницы (по умолчанию 1)"
// @Param        limit          query     int     false  "Лимит на страницу (по умолчанию 10, максимум 100)"
// @Success      200            {object}  dtoResp.IncidentsListResponse
// @Failure      400            {string}  string  "Неверные параметры пагинации"
// @Failure      401            {string}  string  "Не авторизован"
// @Failure      500            {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents [get]
func (h *IncidentHandler) IncidentList(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			http.Error(w, "invalid page parameter (must be >= 1)", http.StatusBadRequest)
			return
		}
		page = p
	}

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 {
			http.Error(w, "invalid limit parameter (must be >= 1)", http.StatusBadRequest)
			return
		}
		limit = l
	}

	result, err := h.incidentCase.ReadIncidentsWithPagination(r.Context(), page, limit)
	if err != nil {
		h.logger.Error("incident list failed",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit))

		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	incidents := make([]dtoResp.IncidentResponse, len(result.Incidents))
	for i, inc := range result.Incidents {
		incidents[i] = dtoResp.IncidentResponse{
			IncidentID: inc.ID,
			Name:       inc.Name,
			Descr:      inc.Descr,
			Latitude:   inc.Latitude,
			Longitude:  inc.Longitude,
			Radius:     inc.Radius,
			IsActive:   inc.IsActive,
			CreatedAt:  inc.CreatedAt,
			UpdatedAt:  inc.UpdatedAt,
		}
	}

	response := dtoResp.IncidentsListResponse{
		Incidents:  incidents,
		Page:       page,
		Limit:      limit,
		TotalPages: result.TotalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// @Summary      Обновить инцидент (оператор)
// @Description  Обновить данные существующей опасной зоны
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        incident_id    path      int                            true  "ID инцидента"
// @Param        request        body      dtoReq.IncidentUpdateRequest   true  "Обновленные данные инцидента"
// @Success      200            {string}  string                         "Инцидент обновлен"
// @Failure      400            {string}  string                         "Неверный формат данных"
// @Failure      401            {string}  string                         "Не авторизован"
// @Failure      404            {string}  string                         "Инцидент не найден"
// @Failure      500            {string}  string                         "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents/{incident_id} [put]
func (h *IncidentHandler) IncidentUpdate(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "incident_id"))
	if err != nil {
		http.Error(w, "id required/not valid", http.StatusBadRequest)
		return
	}

	var req dtoReq.IncidentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90) {
		http.Error(w, "invalid latitude", http.StatusBadRequest)
		return
	}
	if req.Longitude != nil && (*req.Longitude < -180 || *req.Longitude > 180) {
		http.Error(w, "invalid longitude", http.StatusBadRequest)
		return
	}
	if req.Radius != nil && *req.Radius <= 0 {
		http.Error(w, "radius_m must be > 0", http.StatusBadRequest)
		return
	}

	currentIncident, err := h.incidentCase.ReadIncident(r.Context(), id)
	if err != nil {
		h.logger.Error("incident not found for update",
			zap.Error(err),
			zap.Int("id", id))

		http.Error(w, "incident not found", http.StatusNotFound)
		return
	}

	incident := entity.Incident{
		ID: id,
	}

	if req.Name != nil {
		if *req.Name == "" {
			http.Error(w, "name cannot be empty", http.StatusBadRequest)
			return
		}
		incident.Name = *req.Name
	} else {
		incident.Name = currentIncident.Name
	}

	if req.Descr != nil {
		incident.Descr = *req.Descr
	} else {
		incident.Descr = currentIncident.Descr
	}

	if req.Latitude != nil {
		incident.Latitude = *req.Latitude
	} else {
		incident.Latitude = currentIncident.Latitude
	}

	if req.Longitude != nil {
		incident.Longitude = *req.Longitude
	} else {
		incident.Longitude = currentIncident.Longitude
	}

	if req.Radius != nil {
		incident.Radius = *req.Radius
	} else {
		incident.Radius = currentIncident.Radius
	}

	if req.IsActive != nil {
		incident.IsActive = *req.IsActive
	} else {
		incident.IsActive = currentIncident.IsActive
	}

	err = h.incidentCase.UpdateIncident(r.Context(), incident)
	if err != nil {
		h.logger.Error("incident update failed",
			zap.Error(err),
			zap.Int("id", id))

		if err == entity.ErrIncidentNotFound {
			http.Error(w, "incident not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "incident updated"}`))
}

// @Summary      Удалить инцидент (оператор)
// @Description  Мягкое удаление опасной зоны
// @Tags         incidents
// @Produce      json
// @Security     ApiKeyAuth
// @Param        incident_id    path      int     true  "ID инцидента"
// @Success      200            {string}  string  "Инцидент удален"
// @Failure      400            {string}  string  "Неверный ID"
// @Failure      401            {string}  string  "Не авторизован"
// @Failure      404            {string}  string  "Инцидент не найден"
// @Failure      500            {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents/{incident_id} [delete]
func (h *IncidentHandler) IncidentDelete(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+h.apiKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "incident_id"))
	if err != nil {
		http.Error(w, "id required/not valid", http.StatusBadRequest)
		return
	}

	err = h.incidentCase.DeleteIncident(r.Context(), id)
	if err != nil {
		h.logger.Error("incident delete failed",
			zap.Error(err),
			zap.Int("id", id))

		if err == entity.ErrIncidentNotFound {
			http.Error(w, "incident not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "incident deleted"}`))
}
