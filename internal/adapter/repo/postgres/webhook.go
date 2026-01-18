package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.WebhookRepo = (*WebhookRepo)(nil)

type WebhookRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookRepo(pool *pgxpool.Pool) *WebhookRepo {
	return &WebhookRepo{pool: pool}
}

func (r *WebhookRepo) Create(ctx context.Context, webhook entity.Webhook) (int, error) {
	query := `
	INSERT INTO webhooks (
		check_id, state, retry_cnt, payload, created_at, updated_at, scheduled_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id;
	`

	var webhookID int
	err := r.pool.QueryRow(ctx, query,
		webhook.CheckID,
		webhook.State,
		webhook.RetryCnt,
		webhook.Payload,
		time.Now(),
		time.Now(),
		time.Now(),
	).Scan(&webhookID)

	if err != nil {
		return 0, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhookID, nil
}

func (r *WebhookRepo) UpdateState(ctx context.Context, id int, state string, retryCnt int) error {
	query := `
	UPDATE webhooks 
	SET 
		state = $1, 
		retry_cnt = $2, 
		updated_at = NOW(),
		scheduled_at = CASE 
			WHEN $1 = 'in progress' THEN NOW() + INTERVAL '1 minute' * $2
			ELSE scheduled_at
		END
	WHERE id = $3;
	`

	result, err := r.pool.Exec(ctx, query, state, retryCnt, id)
	if err != nil {
		return fmt.Errorf("failed to update webhook status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

func (r *WebhookRepo) MarkAsDelivered(ctx context.Context, id int) error {
	query := `
	UPDATE webhooks 
	SET 
		state = 'delivered', 
		updated_at = NOW()
	WHERE id = $1;
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark webhook as delivered: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

func (r *WebhookRepo) ReadInProgress(ctx context.Context, limit int) ([]*entity.Webhook, error) {
	query := `
	SELECT 
		id, check_id, state, retry_cnt, payload, 
		created_at, updated_at, scheduled_at
	FROM webhooks
	WHERE state='in progress'
		AND scheduled_at <= NOW()
	ORDER BY scheduled_at ASC
	LIMIT $1;
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query 'in progress' webhooks: %w", err)
	}
	defer rows.Close()

	webhooks := make([]*entity.Webhook, 0, limit)
	for rows.Next() {
		wh := &entity.Webhook{}

		err := rows.Scan(
			&wh.ID,
			&wh.CheckID,
			&wh.State,
			&wh.RetryCnt,
			&wh.Payload,
			&wh.CreatedAt,
			&wh.UpdatedAt,
			&wh.ScheduledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		webhooks = append(webhooks, wh)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating webhook rows: %w", err)
	}

	return webhooks, nil
}
