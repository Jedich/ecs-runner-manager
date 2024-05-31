package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"runner-manager-backend/internal/ctrls"
	"runner-manager-backend/internal/ctrls/entities"
	"runner-manager-backend/pkg/response"
)

type repository struct {
	coll *mongo.Collection
	//conn datasource.ConnTx
}

func NewRepository(coll *mongo.Collection) ctrls.Repository {
	return &repository{
		coll: coll,
	}
}

func (r *repository) GetCtrlByID(ctx context.Context, ctrlID string) (*entities.RunnerController, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) GetCtrlsByUserID(ctx context.Context, userID string) (*entities.RunnerController, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) SaveNewCtrl(ctx context.Context, userID string, ctrl *entities.RunnerController) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", response.ErrUserNotFound
	}

	ctrl.ID = primitive.NewObjectID()

	_, err = r.coll.UpdateOne(
		context.TODO(),
		bson.D{{"_id", objectID}},
		bson.D{{"$push", bson.D{{"ctrls", ctrl}}}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return "", err
	}
	return ctrl.ID.Hex(), nil
}
