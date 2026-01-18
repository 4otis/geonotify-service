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

type LocationHandler struct {
	logger *zap.Logger
	uc     cases.LocationUseCase
}

func NewLocationHandler(logger *zap.Logger, uc cases.LocationUseCase) *LocationHandler {
	return &LocationHandler{
		logger: logger,
		uc:     uc,
	}
}

// LocationCheck обрабатывает POST /api/v1/location/check
// @Summary      Проверить координаты
// @Description  Проверить, попадает ли точка в опасную зону (публичный эндпоинт)
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        request body dtoReq.LocationCheckRequest true "Координаты для проверки"
// @Success      200 {object} dtoResp.LocationCheckResponse
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/location/check [post]
func (h *LocationHandler) LocationCheck(w http.ResponseWriter, r *http.Request) {
	var req dtoReq.LocationCheckRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.respondWithError(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	if req.UserID == "" {
		h.respondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		h.respondWithError(w, http.StatusBadRequest, "invalid coordinates")
		return
	}

	hasAlert, incidents, err := h.uc.CheckLocation(r.Context(), req.UserID, req.Latitude, req.Longitude)
	if err != nil {
		h.logger.Error("location check failed",
			zap.Error(err),
			zap.String("user_id", req.UserID))

		switch err {
		case entity.ErrUserIDRequired, entity.ErrInvalidCoordinates:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	incidentResponses := make([]dtoResp.IncidentResponse, len(incidents))
	for i, inc := range incidents {
		if inc != nil {
			incidentResponses[i] = dtoResp.IncidentResponse{
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
	}

	response := dtoResp.LocationCheckResponse{
		HasAlert:  hasAlert,
		Incidents: incidentResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "failed to encode response"}`))
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (h *LocationHandler) respondWithError(w http.ResponseWriter, code int, message string) {
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
