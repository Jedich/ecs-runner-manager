package ctrls

import (
	"context"
	"runner-manager-backend/internal/ctrls/dto"
)

type Usecase interface {
	Register(ctx context.Context, payload *dto.CreateRunnerControllerRequest) (rsp *dto.CreateRunnerControllerResponse, err error)
}
