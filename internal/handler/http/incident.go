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
	logger *zap.Logger
	uc     cases.IncidentUseCase
}

func NewIncidentHandler(logger *zap.Logger, uc cases.IncidentUseCase) *IncidentHandler {
	return &IncidentHandler{
		logger: logger,
		uc:     uc,
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
	var req dtoReq.IncidentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	isValid, msg := h.validateIncidentRequest(req.Name, req.Latitude, req.Longitude, req.Radius)
	if !isValid {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	incident := entity.Incident{
		Name:      req.Name,
		Descr:     req.Descr,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Radius:    req.Radius,
	}

	incidentID, err := h.uc.CreateIncident(r.Context(), incident)
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
// @Param        incident_id  path    string  true  "ID инцидента"
// @Success      200 {object} dtoResp.IncidentResponse
// @Failure      401 {string} string "Не авторизован"
// @Failure      404 {string} string "Инцидент не найден"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents/{incident_id} [get]
func (h *IncidentHandler) IncidentGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "incident_id"))
	if err != nil {
		http.Error(w, "id required/not valid", http.StatusBadRequest)
		return
	}

	incident, err := h.uc.ReadIncident(r.Context(), id)
	if err != nil {
		h.logger.Error("incident get failed",
			zap.Error(err),
			zap.Int("id", id))
		if err == entity.ErrIncidentNotFound {
			http.Error(w, "incident not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
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

	result, err := h.uc.ReadIncidentsWithPagination(r.Context(), page, limit)
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
// @Description  Полное обновление данных существующей опасной зоны (PUT)
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        incident_id    path      int                            true  "ID инцидента"
// @Param        request        body      dtoReq.IncidentUpdateRequest   true  "Полные данные инцидента"
// @Success      200            {string}  string                         "Инцидент обновлен"
// @Failure      400            {string}  string                         "Неверный формат данных"
// @Failure      401            {string}  string                         "Не авторизован"
// @Failure      404            {string}  string                         "Инцидент не найден"
// @Failure      500            {string}  string                         "Внутренняя ошибка сервера"
// @Router       /api/v1/incidents/{incident_id} [put]
func (h *IncidentHandler) IncidentUpdate(w http.ResponseWriter, r *http.Request) {
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

	isValid, msg := h.validateIncidentRequest(req.Name, req.Latitude, req.Longitude, req.Radius)
	if !isValid {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	incident := entity.Incident{
		ID:        id,
		Name:      req.Name,
		Descr:     req.Descr,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Radius:    req.Radius,
		IsActive:  req.IsActive,
	}

	err = h.uc.UpdateIncident(r.Context(), incident)
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
	id, err := strconv.Atoi(chi.URLParam(r, "incident_id"))
	if err != nil {
		http.Error(w, "id required/not valid", http.StatusBadRequest)
		return
	}

	err = h.uc.DeleteIncident(r.Context(), id)
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

func (h *IncidentHandler) validateCoordinates(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

func (h *IncidentHandler) validateIncidentRequest(name string, lat, lng, radius float64) (bool, string) {
	if name == "" {
		return false, "name is required"
	}
	if !h.validateCoordinates(lat, lng) {
		return false, "invalid coordinates"
	}
	if radius <= 0 {
		return false, "radius_m must be > 0"
	}
	return true, ""
}
