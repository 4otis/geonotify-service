package cases

import (
	"context"
	"math"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
)

type IncidentUseCase interface {
	CreateIncident(ctx context.Context, incident entity.Incident) (incID int, err error)
	ReadIncident(ctx context.Context, incId int) (*entity.Incident, error)
	ReadIncidentsWithPagination(ctx context.Context, page, limit int) (IncidentsWithPagination, error)
	UpdateIncident(ctx context.Context, incident entity.Incident) error
	DeleteIncident(ctx context.Context, incID int) error
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
	return uc.repo.Update(ctx, incident)
}

func (uc *IncidentUseCaseImpl) DeleteIncident(ctx context.Context, incID int) error {
	return uc.repo.Delete(ctx, incID)
}

type IncidentsWithPagination struct {
	Incidents  []*entity.Incident
	TotalPages int
}
