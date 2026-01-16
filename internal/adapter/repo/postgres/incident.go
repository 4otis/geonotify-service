package postgres

import (
	"context"
	"fmt"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/4otis/geonotify-service/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.IncidentRepo = (*IncidentRepo)(nil)

type IncidentRepo struct {
	pool *pgxpool.Pool
}

func (r *IncidentRepo) Create(ctx context.Context, incident entity.Incident) (incidentID int, err error) {
	query := `
	INSERT INTO incidents (
		name, descr, latitude, longitude, radius_m, is_active
	) VALUES (
		@name, @descr, @latitude, @longitude, @radius_m, @is_active
	) RETURNING id;
	`
	args := map[string]interface{}{
		"name":      incident.Name,
		"descr":     incident.Descr,
		"latitude":  incident.Latitude,
		"longitude": incident.Longitude,
		"radius_m":  incident.Radius,
		"is_active": true,
	}

	err = postgres.QueryRowNamed(ctx, r.pool, query, args).Scan(&incident)
	if err != nil {
		return 0, fmt.Errorf("failed to create incident: %w", err)
	}

	return incidentID, nil
}
