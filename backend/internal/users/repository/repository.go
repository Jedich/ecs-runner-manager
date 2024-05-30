package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"runner-manager-backend/internal/infrastructure/logs"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/response"
)

type repository struct {
	coll *mongo.Collection
	//conn datasource.ConnTx
}

func NewRepository(coll *mongo.Collection) users.Repository {
	return &repository{
		coll: coll,
	}
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	filter := bson.D{{"email", email}}

	var user entities.User
	err := r.coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, response.ErrUserNotFound
		default:
			logs.Error(err)
			return nil, err
		}
	}

	return &user, nil
}

func (r *repository) SaveNewUser(ctx context.Context, user *entities.User) (string, error) {
	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *repository) UpdateUserByID(ctx context.Context, user *entities.User) error {
	//TODO implement me
	panic("implement me")
}

func (r *repository) IsUserExist(ctx context.Context, email string) bool {
	filter := bson.D{{"email", email}}

	var user entities.User
	err := r.coll.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return false
		default:
			logs.Error(err)
			return false
		}
	}

	return true
}
