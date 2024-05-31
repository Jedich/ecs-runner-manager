package users

import (
	"context"
	"runner-manager-backend/internal/users/entities"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByApiKey(ctx context.Context, apiKey string) (*entities.User, error)
	SaveNewUser(ctx context.Context, user *entities.User) (string, error)
	UpdateUserByID(ctx context.Context, userID string, user *entities.User) error
	IsUserExist(ctx context.Context, email string) bool
}
