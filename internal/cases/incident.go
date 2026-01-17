package cases

import (
	"context"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
)

type IncidentUseCase interface {
	CreateIncident(ctx context.Context, incident entity.Incident) (incID int, err error)
	ReadIncident(ctx context.Context, incId int) (*entity.Incident, error)
	// ReadIncidentsWithPagination(ctx context.Context, page, limit int) (IncidentsWithPagination, error)
	// UpdateIncident(ctx context.Context) error
	// DeleteIncident(ctx context.Context, incID int) error
}

type IncidentUseCaseImpl struct {
	repo repo.IncidentRepo
}

func NewIncidentUseCase(repo repo.IncidentRepo) *IncidentUseCaseImpl {
	return &IncidentUseCaseImpl{
		repo: repo,
	}
}

func (uc *IncidentUseCaseImpl) CreateIncident(ctx context.Context, incident entity.Incident) (incID int, err error) {
	incID, err = uc.repo.Create(ctx, incident)
	if err != nil {
		return incID, err
	}

	return incID, nil
}

func (uc *IncidentUseCaseImpl) ReadIncident(ctx context.Context, incId int) (*entity.Incident, error) {
	incident, err := uc.repo.Read(ctx, incId)
	if err != nil {
		return nil, err
	}

	return incident, nil
}

type IncidentsWithPagination struct {
	Incidents  []*entity.Incident
	TotalPages int
}
