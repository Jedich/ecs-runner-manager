package usecase

import (
	"context"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/dto"
)

type usecase struct {
	repo users.Repository
	cfg  config.Config
}

func NewUseCase(repo users.Repository, cfg config.Config) users.Usecase {
	return &usecase{repo, cfg}
}

func (u *usecase) Login(ctx context.Context, request dto.UserLoginRequest) (response dto.UserLoginResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (u *usecase) Create(ctx context.Context, payload dto.CreateUserRequest) (userID int64, err error) {
	//TODO implement me
	panic("implement me")
}
