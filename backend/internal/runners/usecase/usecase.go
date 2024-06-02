package usecase

import (
	"context"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/runners"
	"runner-manager-backend/internal/runners/dto"
	"runner-manager-backend/internal/runners/entities"
	"runner-manager-backend/internal/users"
	"time"
)

type usecase struct {
	usersRepo users.Repository

	repo runners.Repository
	cfg  config.Config
}

func NewUseCase(usersRepo users.Repository, repo runners.Repository, cfg config.Config) runners.Usecase {
	return &usecase{usersRepo, repo, cfg}
}

func (uc *usecase) UpdateRunners(ctx context.Context, userID, ctrlID string, payload *dto.UpdateRunnersRequest) ([]*dto.RunnerControllerWSResponse, error) {
	r := make([]*entities.Runner, 0, len(payload.Runners))
	for _, runner := range payload.Runners {
		r = append(r, entities.NewRunner(&runner))
	}

	if len(payload.Runners) == 0 {
		return []*dto.RunnerControllerWSResponse{}, nil
	}

	ctrls, runnerMap, err := uc.repo.UpdateRunners(ctx, userID, ctrlID, r)
	if err != nil {
		return nil, err
	}

	for _, runner := range payload.Runners {
		metrics := make([]*entities.Metrics, 0, len(runner.Metrics))

		runnerID := runnerMap[runner.Name].ID
		for _, metric := range runner.Metrics {
			if timestamp, ok := metric["timestamp"]; ok {
				delete(metric, "timestamp")
				metric["runner_id"] = runnerID
				metrics = append(metrics, &entities.Metrics{
					Timestamp: time.Unix(int64(timestamp.(float64)), 0),
					Metadata:  metric,
				})
			}
		}

		err = uc.repo.SaveMetrics(ctx, metrics)
		if err != nil {
			return nil, err
		}
	}

	rsp := make([]*dto.RunnerControllerWSResponse, 0, len(ctrls))
	for _, ctrl := range ctrls {
		runnersWSResponse := make([]*dto.RunnerWSResponse, 0, len(ctrl.Runners))
		for _, runner := range ctrl.Runners {
			runnersWSResponse = append(runnersWSResponse, &dto.RunnerWSResponse{
				Name:        runner.Name,
				PrivateIPv4: runner.PrivateIPv4,
				Status:      runner.Status,
			})
		}
		rsp = append(rsp, &dto.RunnerControllerWSResponse{
			Name:              "controller",
			RunnersWSResponse: runnersWSResponse,
		})
	}

	return rsp, nil
}
