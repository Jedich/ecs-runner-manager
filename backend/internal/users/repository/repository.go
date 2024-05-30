package repository

import (
	"context"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/entities"
)

type repository struct {
	//db   *sqlx.DB
	//conn datasource.ConnTx
}

func NewRepository() users.Repository {
	return &repository{}
}

func (r *repository) GetUserByID(ctx context.Context, i int64) (entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) GetUserByEmail(ctx context.Context, s string) (entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) SaveNewUser(ctx context.Context, user entities.User) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *repository) UpdateUserByID(ctx context.Context, user entities.User) error {
	//TODO implement me
	panic("implement me")
}

func (r *repository) IsUserExist(ctx context.Context, email string) bool {
	//TODO implement me
	panic("implement me")
}
