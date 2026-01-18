package repo

import (
	"context"
	"time"

	"github.com/4otis/geonotify-service/internal/entity"
)

type CheckRepo interface {
	Create(ctx context.Context, check entity.Check) (checkID int, err error)
	GetStats(ctx context.Context, minutes int) (userCnt, totalChecks int, periodStart time.Time, err error)
}
