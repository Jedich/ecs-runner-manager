package usecase

import (
	"context"
	"github.com/golang-jwt/jwt"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/ctrls"
	"runner-manager-backend/internal/ctrls/dto"
	"runner-manager-backend/internal/ctrls/entities"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/pkg/response"
	"time"
)

type usecase struct {
	usersRepo users.Repository

	repo ctrls.Repository
	cfg  config.Config
}

func NewUseCase(usersRepo users.Repository, repo ctrls.Repository, cfg config.Config) ctrls.Usecase {
	return &usecase{usersRepo, repo, cfg}
}

func (uc *usecase) Register(ctx context.Context, request *dto.CreateRunnerControllerRequest) (rsp *dto.CreateRunnerControllerResponse, err error) {
	apiKey := request.ApiKey

	dataLogin, err := uc.usersRepo.GetUserByApiKey(ctx, apiKey)
	if err != nil {
		return rsp, response.ErrUserNotFound
	}

	ctrlID, err := uc.repo.SaveNewCtrl(ctx, dataLogin.ID.Hex(), entities.NewRunnerController(request))
	if err != nil {
		return nil, err
	}

	jwtData := &middleware.Data{
		UserID: dataLogin.ID.Hex(),
		CtrlID: ctrlID,
		Email:  dataLogin.Email,
	}

	claims := middleware.PayloadToken{
		Data: jwtData,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		},
	}

	// Calculate the expiration time in seconds
	expiresIn := claims.ExpiresAt - time.Now().Unix()

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	tokenString, err := token.SignedString([]byte(uc.cfg.JWT.Key))
	if err != nil {
		return nil, err
	}

	return &dto.CreateRunnerControllerResponse{CtrlID: ctrlID, AccessToken: tokenString, ExpiredAt: expiresIn}, nil
}
