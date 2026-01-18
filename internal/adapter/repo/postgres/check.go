package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.CheckRepo = (*CheckRepo)(nil)

type CheckRepo struct {
	pool *pgxpool.Pool
}

func NewCheckRepo(pool *pgxpool.Pool) *CheckRepo {
	return &CheckRepo{pool: pool}
}

func (r *CheckRepo) Create(ctx context.Context, check entity.Check) (checkID int, err error) {
	query := `
	INSERT INTO checks (user_id, latitude, longitude, has_alert, created_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	err = r.pool.QueryRow(ctx, query,
		check.UserID,
		check.Latitude,
		check.Longitude,
		check.HasAlert,
		time.Now(),
	).Scan(&checkID)

	if err != nil {
		return 0, fmt.Errorf("failed to create check: %w", err)
	}

	return checkID, nil
}

func (r *CheckRepo) GetStats(ctx context.Context, windowMinutes int) (userCount, totalChecks int, periodStart time.Time, err error) {
	query := `
	SELECT 
		COUNT(DISTINCT user_id) as user_count,
		COUNT(*) as total_checks,
		NOW() - INTERVAL '%d minutes' as period_start
	FROM checks
	WHERE created_at >= NOW() - INTERVAL '%d minutes';
	`

	query = fmt.Sprintf(query, windowMinutes, windowMinutes)

	err = r.pool.QueryRow(ctx, query).Scan(&userCount, &totalChecks, &periodStart)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to get stats: %w", err)
	}

	return userCount, totalChecks, periodStart, nil
}
