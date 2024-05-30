package usecase

import (
	"context"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/dto"
	"runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/app_crypto"
	"runner-manager-backend/pkg/response"
)

type usecase struct {
	repo users.Repository
	cfg  config.Config
}

func NewUseCase(repo users.Repository, cfg config.Config) users.Usecase {
	return &usecase{repo, cfg}
}

func (uc *usecase) Login(ctx context.Context, request *dto.UserLoginRequest) (response *dto.UserLoginResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (uc *usecase) Create(ctx context.Context, payload *dto.CreateUserRequest) (userID int64, err error) {
	if exist := uc.repo.IsUserExist(ctx, payload.Email); exist {
		return userID, response.Conflict(response.ErrEmailAlreadyExist)
	}

	hashedPassword, err := app_crypto.Hash(payload.Password)
	if err != nil {
		return userID, response.InternalServerError(err)
	}
	payload.Password = hashedPassword

	userID, err = uc.repo.SaveNewUser(ctx, entities.NewUser(payload))
	if err != nil {
		return userID, response.InternalServerError(err)
	}

	return userID, nil
}

func (uc *usecase) GenerateApiKey(ctx context.Context, userID int64) (apiKey string, err error) {
	//TODO implement me
	panic("implement me")
}
