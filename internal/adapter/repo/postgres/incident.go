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

func (r *IncidentRepo) Read(ctx context.Context, incID int) (*entity.Incident, error) {
	query := `
	SELECT (
		id, name, descr, latitude, longitude,
		radius_m, is_active, created_at, updated_at
	) FROM incidents
	WHERE id=$1;	
	`

	i := &entity.Incident{}

	err := r.pool.QueryRow(ctx, query, incID).Scan(
		&i.ID,
		&i.Name,
		&i.Descr,
		&i.Latitude,
		&i.Longitude,
		&i.Radius,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	if err != nil {
		return &entity.Incident{},
			fmt.Errorf("failed to select incident (by id=%v): %w", incID, err)
	}
	return i, nil
}
