package repo

import (
	"context"

	"github.com/4otis/geonotify-service/internal/entity"
)

type IncidentRepo interface {
	Create(ctx context.Context, incident entity.Incident) (incidentID int, err error)
	Read(ctx context.Context, incID int) (i *entity.Incident, err error)
	ReadWithPagination(ctx context.Context, page, limit int) ([]*entity.Incident, int, error)
	Update(ctx context.Context, incident entity.Incident) error
	Delete(ctx context.Context, incID int) error
}
