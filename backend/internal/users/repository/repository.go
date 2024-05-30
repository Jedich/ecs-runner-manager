package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/entities"
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

func (r *repository) GetUserByID(ctx context.Context, i int64) (*entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) GetUserByEmail(ctx context.Context, s string) (*entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) SaveNewUser(ctx context.Context, user *entities.User) (int64, error) {
	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return 0, err
	}
	return result.InsertedID.(int64), nil
}

func (r *repository) UpdateUserByID(ctx context.Context, user *entities.User) error {
	//TODO implement me
	panic("implement me")
}

func (r *repository) IsUserExist(ctx context.Context, email string) bool {
	//TODO implement me
	panic("implement me")
}
