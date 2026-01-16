package cases

import (
	"context"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
)

type IncidentUseCase interface {
	Create(ctx context.Context, incident entity.Incident) (incID int, err error)
}

type IncidentUseCaseImpl struct {
	repo repo.IncidentRepo
}

func NewIncidentUseCase(repo repo.IncidentRepo) *IncidentUseCaseImpl {
	return &IncidentUseCaseImpl{
		repo: repo,
	}
}

func (uc *IncidentUseCaseImpl) Create(ctx context.Context, incident entity.Incident) (incID int, err error) {
	incID, err = uc.repo.Create(ctx, incident)
	if err != nil {
		return incID, err
	}

	return incID, nil
}
