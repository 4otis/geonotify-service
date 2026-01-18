package cases

import (
	"context"
	"math"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"go.uber.org/zap"
)

var _ IncidentUseCase = (*IncidentUseCaseImpl)(nil)

type IncidentUseCase interface {
	CreateIncident(ctx context.Context, incident entity.Incident) (incID int, err error)
	ReadIncident(ctx context.Context, incId int) (*entity.Incident, error)
	ReadIncidentsWithPagination(ctx context.Context, page, limit int) (IncidentsWithPagination, error)
	UpdateIncident(ctx context.Context, incident entity.Incident) error
	DeleteIncident(ctx context.Context, incID int) error
}

type IncidentUseCaseImpl struct {
	repo         repo.IncidentRepo
	locationCase LocationUseCase
	logger       *zap.Logger
}

func NewIncidentUseCase(repo repo.IncidentRepo,
	locationCase LocationUseCase, logger *zap.Logger) *IncidentUseCaseImpl {
	return &IncidentUseCaseImpl{
		repo:         repo,
		locationCase: locationCase,
		logger:       logger,
	}
}

func (uc *IncidentUseCaseImpl) CreateIncident(ctx context.Context, incident entity.Incident) (incID int, err error) {
	incID, err = uc.repo.Create(ctx, incident)
	if err != nil {
		return 0, err
	}

	if err := uc.locationCase.InvalidateIncidentsCache(ctx); err != nil {
		uc.logger.Warn("failed to invalidate cache after creating incident",
			zap.Error(err))
	}

	return incID, nil
}

func (uc *IncidentUseCaseImpl) ReadIncident(ctx context.Context, incId int) (*entity.Incident, error) {
	return uc.repo.Read(ctx, incId)
}

func (uc *IncidentUseCaseImpl) ReadIncidentsWithPagination(ctx context.Context, page, limit int) (IncidentsWithPagination, error) {

	if page < 1 {
		page = 1
	}

	incidents, totalCount, err := uc.repo.ReadWithPagination(ctx, page, limit)
	if err != nil {
		return IncidentsWithPagination{}, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	return IncidentsWithPagination{
		Incidents:  incidents,
		TotalPages: totalPages,
	}, nil
}

func (uc *IncidentUseCaseImpl) UpdateIncident(ctx context.Context, incident entity.Incident) error {
	err := uc.repo.Update(ctx, incident)
	if err != nil {
		return err
	}

	if err := uc.locationCase.InvalidateIncidentsCache(ctx); err != nil {
		uc.logger.Warn("failed to invalidate cache after updating incident",
			zap.Error(err))
	}

	return nil

}

func (uc *IncidentUseCaseImpl) DeleteIncident(ctx context.Context, incID int) error {
	err := uc.repo.Delete(ctx, incID)
	if err != nil {
		return err
	}

	if err := uc.locationCase.InvalidateIncidentsCache(ctx); err != nil {
		uc.logger.Warn("failed to invalidate cache after deleting incident",
			zap.Error(err))
	}

	return nil
}

type IncidentsWithPagination struct {
	Incidents  []*entity.Incident
	TotalPages int
}
