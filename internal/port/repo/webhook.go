package repo

import (
	"context"

	"github.com/4otis/geonotify-service/internal/entity"
)

type WebhookRepo interface {
	Create(ctx context.Context, w entity.Webhook) (webhookID int, err error)
	UpdateState(ctx context.Context, id int, newState string, retryCnt int) error
	Read(ctx context.Context, id int) (*entity.Webhook, error)
	ReadInProgress(ctx context.Context, limit int) ([]*entity.Webhook, error)
	MarkAsDelivered(ctx context.Context, id int) error
}
