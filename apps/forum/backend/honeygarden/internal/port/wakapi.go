package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type WakapiRepository interface {
	Save(ctx context.Context, acc domain.WakapiAccount) error
	Get(ctx context.Context, userID uuid.UUID) (domain.WakapiAccount, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

type WakapiClient interface {
	FetchStats(ctx context.Context, instanceURL, apiKey, username string) (domain.WakapiStats, error)
}
