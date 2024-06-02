package repository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	ctrlEntities "runner-manager-backend/internal/ctrls/entities"
	"runner-manager-backend/internal/runners"
	"runner-manager-backend/internal/runners/entities"
	userEntities "runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/response"
	"time"
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

func (r *repository) UpdateRunners(ctx context.Context, userID string, ctrlID string, runners []*entities.Runner) ([]*ctrlEntities.RunnerController, map[string]*entities.Runner, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil, response.ErrUserNotFound
	}

	ctrlObjectID, err := primitive.ObjectIDFromHex(ctrlID)
	if err != nil {
		return nil, nil, response.ErrUserNotFound
	}

	var user *userEntities.User
	err = r.usersColl.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, nil, err
	}

	var ctrlFound bool
	var runnerMap map[string]*entities.Runner
	var ctrls []*ctrlEntities.RunnerController
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
					newRunner.CreatedAt = oldRunner.CreatedAt
					if newRunner.Status != oldRunner.Status {
						newRunner.UpdatedAt = time.Now()
					} else {
						newRunner.UpdatedAt = oldRunner.UpdatedAt
					}
				} else {
					newRunner.ID = primitive.NewObjectID()
					newRunner.Color = randomColor()
					user.RunnerController[i].UpdatedAt = time.Now()
				}
				runnerMap[newRunner.Name] = newRunner
			}

			var updatedRunners []*entities.Runner
			for _, runner := range runnerMap {
				updatedRunners = append(updatedRunners, runner)
			}

			user.RunnerController[i].Runners = updatedRunners
			ctrls = user.RunnerController
			runners = updatedRunners
			break
		}
	}

	if !ctrlFound {
		return ctrls, nil, errors.New("controller not found")
	}

	newCtrls := make([]*ctrlEntities.RunnerController, 0, len(user.RunnerController))
	for _, ctrl := range user.RunnerController {
		if len(ctrl.Runners) == 0 && ctrl.UpdatedAt.Add(time.Minute*60).Before(time.Now()) {
			continue
		}
		newCtrls = append(newCtrls, ctrl)
	}
	user.RunnerController = newCtrls

	// Update the user document with the modified controllers
	_, err = r.usersColl.ReplaceOne(ctx, bson.M{"_id": userObjectID}, user)
	if err != nil {
		return nil, nil, err
	}

	return ctrls, runnerMap, nil
}

func (r *repository) SaveMetrics(ctx context.Context, metrics []*entities.Metrics) error {
	s := make([]interface{}, len(metrics))
	for i, v := range metrics {
		s[i] = v
	}

	if len(s) == 0 {
		return nil
	}
	_, err := r.metricsColl.InsertMany(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetAllMetricsByCtrlID(ctx context.Context, userID, ctrlID string) (map[string][]*entities.Metrics, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, response.ErrUserNotFound
	}

	ctrlObjectID, err := primitive.ObjectIDFromHex(ctrlID)
	if err != nil {
		return nil, response.ErrUserNotFound
	}

	var user *userEntities.User
	err = r.usersColl.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	runnerIDs := make([]primitive.ObjectID, 0, 4)
	for _, ctrl := range user.RunnerController {
		if ctrl.ID != ctrlObjectID {
			continue
		}
		for _, runner := range ctrl.Runners {
			runnerIDs = append(runnerIDs, runner.ID)
		}
		break
	}

	res := make(map[string][]*entities.Metrics)
	for _, runnerID := range runnerIDs {
		var metrics []*entities.Metrics
		cursor, err := r.metricsColl.Find(
			ctx,
			bson.M{"metadata.runner_id": runnerID},
			options.Find().SetProjection(bson.M{"metadata.runner_id": 0}),
		)
		if err != nil {
			return nil, err
		}
		if err = cursor.All(ctx, &metrics); err != nil {
			return nil, err
		}
		res[runnerID.Hex()] = metrics
	}

	return res, nil
}

func (r *repository) GetAllCtrlsByUserID(ctx context.Context, userID string) ([]*ctrlEntities.RunnerController, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, response.ErrUserNotFound
	}

	var user *userEntities.User
	err = r.usersColl.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user.RunnerController, nil
}

func randomColor() string {
	m := 100
	r := m + rand.Intn(156) // 100-255
	g := m + rand.Intn(156) // 100-255
	b := m + rand.Intn(156) // 100-255
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
