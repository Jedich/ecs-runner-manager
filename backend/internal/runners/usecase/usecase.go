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

func (uc *usecase) UpdateRunners(ctx context.Context, userID, ctrlID string, payload *dto.UpdateRunnersRequest) (err error) {
	r := make([]*entities.Runner, 0, len(payload.Runners))
	for _, runner := range payload.Runners {
		r = append(r, entities.NewRunner(&runner))
	}

	if len(payload.Runners) == 0 {
		return nil
	}

	runnerMap, err := uc.repo.UpdateRunners(ctx, userID, ctrlID, r)
	if err != nil {
		return err
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
			return err
		}
	}

	return nil
}
