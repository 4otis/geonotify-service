package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/4otis/geonotify-service/pkg/postgres"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.IncidentRepo = (*IncidentRepo)(nil)

type IncidentRepo struct {
	pool *pgxpool.Pool
}

func NewIncidentRepo(pool *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{
		pool: pool,
	}
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

	err = postgres.QueryRowNamed(ctx, r.pool, query, args).Scan(&incidentID)
	if err != nil {
		return 0, fmt.Errorf("failed to create incident: %w", err)
	}

	return incidentID, nil
}

func (r *IncidentRepo) Read(ctx context.Context, incID int) (*entity.Incident, error) {
	query := `
	SELECT 
		id, name, descr, latitude, longitude,
		radius_m, is_active, created_at, updated_at
	FROM incidents
	WHERE id=$1 AND deleted_at IS NULL;	
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrIncidentNotFound
		}
		return nil, fmt.Errorf("failed to select incident (by id=%v): %w", incID, err)
	}

	return i, nil
}

func (r *IncidentRepo) ReadWithPagination(ctx context.Context, page, limit int) ([]*entity.Incident, int, error) {
	query := `
	SELECT COUNT(*)
	FROM incidents
	WHERE deleted_at IS NULL;
	`
	totalIncidents := 0

	err := r.pool.QueryRow(ctx, query).Scan(&totalIncidents)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	incidents := make([]*entity.Incident, 0, limit)

	if totalIncidents == 0 && page == 1 {
		return incidents, totalIncidents, nil
	}

	query = `
	SELECT 
		id, name, descr, latitude, longitude,
		radius_m, is_active, created_at, updated_at
	FROM incidents
	WHERE deleted_at IS NULL
	ORDER BY updated_at DESC
	LIMIT $1 OFFSET $2;
	`

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query incident: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		i := &entity.Incident{}
		err := rows.Scan(
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
			return nil, 0, fmt.Errorf("failed to scan incident from rows: %w", err)
		}
		incidents = append(incidents, i)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error while iterating incident rows: %w", err)
	}

	return incidents, totalIncidents, nil
}

func (r *IncidentRepo) Update(ctx context.Context, incident entity.Incident) error {
	query := `
	UPDATE incidents 
	SET 
		name = $1,
		descr = $2,
		latitude = $3,
		longitude = $4,
		radius_m = $5,
		is_active = $6,
		updated_at = NOW()
	WHERE id = $7 AND deleted_at IS NULL;
	`

	result, err := r.pool.Exec(ctx, query,
		incident.Name,
		incident.Descr,
		incident.Latitude,
		incident.Longitude,
		incident.Radius,
		incident.IsActive,
		incident.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update incident (id=%v): %w", incident.ID, err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrIncidentNotFound
	}

	return nil
}

func (r *IncidentRepo) Delete(ctx context.Context, incID int) error {
	query := `
	UPDATE incidents 
	SET 
		deleted_at = NOW(),
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`

	result, err := r.pool.Exec(ctx, query, incID)
	if err != nil {
		return fmt.Errorf("failed to soft delete incident (id=%v): %w", incID, err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrIncidentNotFound
	}

	return nil
}

func (r *IncidentRepo) ReadAllActive(ctx context.Context) ([]*entity.Incident, error) {
	query := `
	SELECT
		id, name, descr, latitude, longitude,
		radius_m, is_active, created_at, ipdated_at
	FROM incidents
	WHERE is_active=true AND deleted_at IS NULL
	ORDER BY updated_at DESC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all active incidents: %w", err)
	}
	defer rows.Close()

	incidents := make([]*entity.Incident, 0)
	for rows.Next() {
		i := &entity.Incident{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan incident from rows: %w", err)
		}
		incidents = append(incidents, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating incident rows: %w", err)
	}

	return incidents, nil
}
