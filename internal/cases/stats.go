package cases

import (
	"context"
	"fmt"
	"time"

	"github.com/4otis/geonotify-service/internal/port/repo"
	"go.uber.org/zap"
)

var _ StatsUseCase = (*StatsUseCaseImpl)(nil)

type StatsUseCase interface {
	GetStats(ctx context.Context, windowMinutes int) (userCount, totalChecks int, periodStart time.Time, err error)
	GetActiveIncidentsCount(ctx context.Context) (int, error)
	GetPendingWebhooksCount(ctx context.Context) (int, error)
}

type StatsUseCaseImpl struct {
	incidentRepo repo.IncidentRepo
	checkRepo    repo.CheckRepo
	webhookRepo  repo.WebhookRepo
	logger       *zap.Logger
}

func NewStatsUseCase(
	incidentRepo repo.IncidentRepo,
	checkRepo repo.CheckRepo,
	webhookRepo repo.WebhookRepo,
	logger *zap.Logger,
) *StatsUseCaseImpl {
	return &StatsUseCaseImpl{
		incidentRepo: incidentRepo,
		checkRepo:    checkRepo,
		webhookRepo:  webhookRepo,
		logger:       logger,
	}
}

func (uc *StatsUseCaseImpl) GetStats(ctx context.Context, windowMinutes int) (userCount, totalChecks int, periodStart time.Time, err error) {
	if windowMinutes <= 0 {
		return 0, 0, time.Time{}, fmt.Errorf("window minutes must be positive")
	}

	userCount, totalChecks, periodStart, err = uc.checkRepo.GetStats(ctx, windowMinutes)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to get stats: %w", err)
	}

	uc.logger.Debug("stats retrieved",
		zap.Int("window_minutes", windowMinutes),
		zap.Int("user_count", userCount),
		zap.Int("total_checks", totalChecks))

	return userCount, totalChecks, periodStart, nil
}

func (uc *StatsUseCaseImpl) GetActiveIncidentsCount(ctx context.Context) (int, error) {
	incidents, err := uc.incidentRepo.ReadAllActive(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get active incidents: %w", err)
	}

	return len(incidents), nil
}

func (uc *StatsUseCaseImpl) GetPendingWebhooksCount(ctx context.Context) (int, error) {
	const limit = 1000
	webhooks, err := uc.webhookRepo.ReadInProgress(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending webhooks: %w", err)
	}

	return len(webhooks), nil
}
