package cases

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/4otis/geonotify-service/pkg/redis"
	"go.uber.org/zap"
)

type LocationUseCase interface {
	CheckLocation(ctx context.Context, userID string, lat, lng float64) (bool, []entity.Incident, error)
	InvalidateIncidentsCache(ctx context.Context) error
}

type LocationUseCaseImpl struct {
	incidentRepo repo.IncidentRepo
	checkRepo    repo.CheckRepo
	webhookRepo  repo.WebhookRepo
	redis        *redis.Client
	logger       *zap.Logger
	cacheTTL     time.Duration
}

func NewLocationUseCase(
	incidentRepo repo.IncidentRepo,
	checkRepo repo.CheckRepo,
	webhookRepo repo.WebhookRepo,
	redis *redis.Client,
	logger *zap.Logger,
	cacheTTLMinutes int,
) *LocationUseCaseImpl {
	return &LocationUseCaseImpl{
		incidentRepo: incidentRepo,
		checkRepo:    checkRepo,
		webhookRepo:  webhookRepo,
		redis:        redis,
		logger:       logger,
		cacheTTL:     time.Duration(cacheTTLMinutes) * time.Minute,
	}
}

func (uc *LocationUseCaseImpl) CheckLocation(ctx context.Context, userID string, lat, lng float64) (bool, []*entity.Incident, error) {
	if strings.TrimSpace(userID) == "" {
		return false, nil, entity.ErrUserIDRequired
	}

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return false, nil, entity.ErrInvalidCoordinates
	}

	uc.logger.Debug("checking location",
		zap.String("user_id", userID),
		zap.Float64("lat", lat),
		zap.Float64("lng", lng))

	activeIncidents, err := uc.getActiveIncidents(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get active incidents: %w", err)
	}

	matchingIncidents := uc.findMatchingIncidents(lat, lng, activeIncidents)
	hasAlert := len(matchingIncidents) > 0

	uc.logger.Debug("mathcingIncidents",
		zap.Int("amount", len(matchingIncidents)),
		zap.String("user_id", userID),
	)

	checkID, err := uc.saveCheck(ctx, userID, lat, lng, hasAlert)
	if err != nil {
		return false, nil, fmt.Errorf("failed to save check: %w", err)
	}

	if hasAlert {
		if err := uc.createWebhook(ctx, checkID, matchingIncidents); err != nil {
			uc.logger.Error("failed to create webhook",
				zap.Error(err),
				zap.Int("check_id", checkID))
		}
	}

	return hasAlert, matchingIncidents, nil
}

func (uc *LocationUseCaseImpl) getActiveIncidents(ctx context.Context) ([]*entity.Incident, error) {
	cacheKey := "active_incidents:v1"

	var cachedIncidents []*entity.Incident
	if err := uc.redis.Get(cacheKey, &cachedIncidents); err == nil {
		uc.logger.Debug("retrieved active incidents from cache",
			zap.Int("count", len(cachedIncidents)))
		return cachedIncidents, nil
	}

	uc.logger.Debug("failed to get active incidents from cache")

	incidents, err := uc.incidentRepo.ReadAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active incidents from DB: %w", err)
	}

	uc.logger.Debug("retrieved active incidents from DB",
		zap.Int("count", len(incidents)))

	if err := uc.redis.Set(cacheKey, incidents, uc.cacheTTL); err != nil {
		uc.logger.Debug("failed to cache incidents",
			zap.Error(err))
	}

	uc.logger.Debug("successfully cached incidents",
		zap.Int("count", len(incidents)))

	return incidents, nil
}

func (uc *LocationUseCaseImpl) findMatchingIncidents(lat, lng float64, incidents []*entity.Incident) []*entity.Incident {
	var matching []*entity.Incident

	for _, incident := range incidents {
		if isPointInRadius(lat, lng, incident.Latitude, incident.Longitude, incident.Radius) {
			matching = append(matching, incident)
		}
	}

	return matching
}

func (uc *LocationUseCaseImpl) saveCheck(ctx context.Context, userID string, lat, lng float64, hasAlert bool) (int, error) {
	check := entity.Check{
		UserID:    userID,
		Latitude:  lat,
		Longitude: lng,
		HasAlert:  hasAlert,
	}

	checkID, err := uc.checkRepo.Create(ctx, check)
	if err != nil {
		return 0, fmt.Errorf("failed to create check record: %w", err)
	}

	uc.logger.Debug("check saved",
		zap.Int("check_id", checkID),
		zap.Bool("has_alert", hasAlert))

	return checkID, nil
}

func (uc *LocationUseCaseImpl) createWebhook(ctx context.Context, checkID int, incidents []*entity.Incident) error {
	payload := map[string]interface{}{
		"check_id":  checkID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"incidents": incidents,
	}
	// json.Marshal разыменовывает указатели,
	// []*entity.Incident обработается корректно
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	webhook := entity.Webhook{
		CheckID:     checkID,
		State:       "in progress",
		RetryCnt:    0,
		Payload:     payloadBytes,
		ScheduledAt: time.Now(),
	}

	webhookID, err := uc.webhookRepo.Create(ctx, webhook)
	if err != nil {
		return fmt.Errorf("failed to create webhook record: %w", err)
	}

	queueTask := map[string]interface{}{
		"webhook_id": webhookID,
		"check_id":   checkID,
		"payload":    string(payloadBytes),
	}

	if err := uc.redis.LPush("webhooks:queue", queueTask); err != nil {
		uc.logger.Error("failed to push webhook to queue",
			zap.Error(err),
			zap.Int("webhook_id", webhookID))
	}

	uc.logger.Info("webhook created",
		zap.Int("webhook_id", webhookID),
		zap.Int("check_id", checkID),
		zap.Int("incidents_count", len(incidents)))

	return nil
}

func (uc *LocationUseCaseImpl) InvalidateIncidentsCache(ctx context.Context) error {
	cacheKey := "active_incidents:v1"
	if err := uc.redis.Delete(cacheKey); err != nil && err != redis.ErrNotFound {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	uc.logger.Debug("incidents cache invalidated")
	return nil
}

func isPointInRadius(lat1, lon1, lat2, lon2, radius float64) bool {
	const earthRadius_m = 6371000

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Формула гаверсинусов
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius_m * c
	return distance <= radius
}
