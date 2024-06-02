package runners

import (
	"context"
	"runner-manager-backend/internal/runners/dto"
)

type Usecase interface {
	UpdateRunners(ctx context.Context, userID, ctrlID string, payload *dto.UpdateRunnersRequest) ([]*dto.RunnerControllerWSResponse, error)
	GetAllCtrlsByUserID(ctx context.Context, userID string) ([]*dto.RunnerControllerWSResponse, error)
	GetAllMetricsByCtrlID(ctx context.Context, userID, ctrlID string) (*dto.MetricsCtrlWSResponse, error)
}
