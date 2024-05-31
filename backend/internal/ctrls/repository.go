package ctrls

import (
	"context"
	"runner-manager-backend/internal/ctrls/entities"
)

type Repository interface {
	GetCtrlByID(ctx context.Context, ctrlID string) (*entities.RunnerController, error)
	GetCtrlsByUserID(ctx context.Context, userID string) (*entities.RunnerController, error)
	SaveNewCtrl(ctx context.Context, userID string, ctrl *entities.RunnerController) (string, error)
}
