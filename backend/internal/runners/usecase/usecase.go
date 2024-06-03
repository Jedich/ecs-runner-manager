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
				Id:          runner.ID.Hex(),
				Name:        runner.Name,
				PrivateIPv4: runner.PrivateIPv4,
				Status:      runner.Status,
			})
		}
		rsp = append(rsp, &dto.RunnerControllerWSResponse{
			Id:                ctrl.ID.Hex(),
			Name:              ctrl.Name,
			RunnersWSResponse: runnersWSResponse,
		})
	}

	return rsp, nil
}

func (uc *usecase) GetAllCtrlsByUserID(ctx context.Context, userID string) ([]*dto.RunnerControllerWSResponse, error) {
	ctrls, err := uc.repo.GetAllCtrlsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	rsp := make([]*dto.RunnerControllerWSResponse, 0, len(ctrls))
	for _, ctrl := range ctrls {
		runnersWSResponse := make([]*dto.RunnerWSResponse, 0, len(ctrl.Runners))
		for _, runner := range ctrl.Runners {
			runnersWSResponse = append(runnersWSResponse, &dto.RunnerWSResponse{
				Id:          runner.ID.Hex(),
				Name:        runner.Name,
				PrivateIPv4: runner.PrivateIPv4,
				Status:      runner.Status,
			})
		}
		rsp = append(rsp, &dto.RunnerControllerWSResponse{
			Id:                ctrl.ID.Hex(),
			Name:              ctrl.Name,
			RunnersWSResponse: runnersWSResponse,
		})
	}
	return rsp, nil
}

func (uc *usecase) GetAllMetricsByCtrlID(ctx context.Context, userID, ctrlID string) (*dto.MetricsCtrlWSResponse, error) {
	metrics, err := uc.repo.GetAllMetricsByCtrlID(ctx, userID, ctrlID)
	if err != nil {
		return nil, err
	}

	r := make([]*dto.MetricsRunnerWSResponse, 0, 1)
	for runnerName, m := range metrics {
		runnerMetrics := make([]*dto.MetricsWSResponse, 0, len(m))
		for _, metric := range m {
			runnerMetrics = append(runnerMetrics, &dto.MetricsWSResponse{
				Timestamp: metric.Timestamp.Format(time.RFC3339),
				Metadata:  metric.Metadata,
			})
		}
		r = append(r, &dto.MetricsRunnerWSResponse{
			Name:          runnerName,
			RunnerMetrics: runnerMetrics,
		})
	}
	res := &dto.MetricsCtrlWSResponse{
		RunnerMetrics: r,
	}

	return res, nil
}
