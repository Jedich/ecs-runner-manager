package users

import (
	"context"
	"runner-manager-backend/internal/users/entities"
)

type Repository interface {
	GetUserByID(context.Context, string) (*entities.User, error)
	GetUserByEmail(context.Context, string) (*entities.User, error)
	SaveNewUser(context.Context, *entities.User) (string, error)
	UpdateUserByID(context.Context, *entities.User) error
	IsUserExist(ctx context.Context, email string) bool
}
