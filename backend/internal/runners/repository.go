package runners

import (
	"context"
	ctrlEntities "runner-manager-backend/internal/ctrls/entities"
	"runner-manager-backend/internal/runners/entities"
)

type Repository interface {
	UpdateRunners(ctx context.Context, userID string, ctrlID string, runners []*entities.Runner) ([]*ctrlEntities.RunnerController, map[string]*entities.Runner, error)
	SaveMetrics(ctx context.Context, metrics []*entities.Metrics) error
	GetAllMetricsByCtrlID(ctx context.Context, userID, ctrlID string) (map[string][]*entities.Metrics, error)
	GetAllCtrlsByUserID(ctx context.Context, userID string) ([]*ctrlEntities.RunnerController, error)
}
