package users

import (
	"context"
	"runner-manager-backend/internal/users/dto"
)

type Usecase interface {
	Login(ctx context.Context, request *dto.UserLoginRequest) (response *dto.UserLoginResponse, err error)
	Create(ctx context.Context, payload *dto.CreateUserRequest) (userID int64, err error)
	GenerateApiKey(ctx context.Context, userID int64) (apiKey string, err error)
}
