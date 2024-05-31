package runners

import (
	"context"
	"runner-manager-backend/internal/runners/entities"
)

type Repository interface {
	UpdateRunners(ctx context.Context, userID string, ctrlID string, runners []*entities.Runner) (map[string]*entities.Runner, error)
	SaveMetrics(ctx context.Context, metrics []*entities.Metrics) error
}
