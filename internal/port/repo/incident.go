package repo

import (
	"context"

	"github.com/4otis/geonotify-service/internal/entity"
)

type IncidentRepo interface {
	Create(ctx context.Context, incident entity.Incident) (incidentID int, err error)
	// Read()
	// ReadAllWithPagination()
	// Update()
	// Delete() error
}
