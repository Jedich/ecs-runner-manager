package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"runner-manager-backend/internal/runners"
	"runner-manager-backend/internal/runners/entities"
	userEntities "runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/response"
)

type repository struct {
	usersColl   *mongo.Collection
	metricsColl *mongo.Collection
	//conn datasource.ConnTx
}

func NewRepository(usersColl *mongo.Collection, metricsColl *mongo.Collection) runners.Repository {
	return &repository{
		usersColl:   usersColl,
		metricsColl: metricsColl,
	}
}

func (r *repository) UpdateRunners(ctx context.Context, userID string, ctrlID string, runners []*entities.Runner) (map[string]*entities.Runner, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, response.ErrUserNotFound
	}

	ctrlObjectID, err := primitive.ObjectIDFromHex(ctrlID)
	if err != nil {
		return nil, response.ErrUserNotFound
	}

	var user userEntities.User
	err = r.usersColl.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	var ctrlFound bool
	var runnerMap map[string]*entities.Runner
	for i, ctrl := range user.RunnerController {
		if ctrl.ID == ctrlObjectID {
			ctrlFound = true
			runnerMap = make(map[string]*entities.Runner)
			for _, existingRunner := range ctrl.Runners {
				runnerMap[existingRunner.Name] = existingRunner
			}

			for _, newRunner := range runners {
				if oldRunner, ok := runnerMap[newRunner.Name]; ok {
					newRunner.ID = oldRunner.ID
				} else {
					newRunner.ID = primitive.NewObjectID()
				}
				runnerMap[newRunner.Name] = newRunner
			}

			var updatedRunners []*entities.Runner
			for _, runner := range runnerMap {
				updatedRunners = append(updatedRunners, runner)
			}

			user.RunnerController[i].Runners = updatedRunners
			runners = updatedRunners
			break
		}
	}

	if !ctrlFound {
		return nil, errors.New("controller not found")
	}

	// Update the user document with the modified controllers
	_, err = r.usersColl.ReplaceOne(ctx, bson.M{"_id": userObjectID}, user)
	if err != nil {
		return nil, err
	}

	return runnerMap, nil
}

func (r *repository) SaveMetrics(ctx context.Context, metrics []*entities.Metrics) error {
	s := make([]interface{}, len(metrics))
	for i, v := range metrics {
		s[i] = v
	}

	_, err := r.metricsColl.InsertMany(ctx, s)
	if err != nil {
		return err
	}

	return nil
}
