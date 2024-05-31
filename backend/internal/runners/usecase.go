package runners

import (
	"context"
	"runner-manager-backend/internal/runners/dto"
)

type Usecase interface {
	UpdateRunners(ctx context.Context, userID, ctrlID string, payload *dto.UpdateRunnersRequest) (err error)
}
