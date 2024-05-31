package usecase

import (
	"context"
	"github.com/golang-jwt/jwt"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/dto"
	"runner-manager-backend/internal/users/entities"
	"runner-manager-backend/pkg/app_crypto"
	"runner-manager-backend/pkg/response"
	"time"
)

type usecase struct {
	repo users.Repository
	cfg  config.Config
}

func NewUseCase(repo users.Repository, cfg config.Config) users.Usecase {
	return &usecase{repo, cfg}
}

func (uc *usecase) Login(ctx context.Context, request *dto.UserLoginRequest) (rsp *dto.UserLoginResponse, err error) {
	dataLogin, err := uc.repo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return rsp, response.InternalServerError(err)
	}

	if !app_crypto.Verify(dataLogin.Password, request.Password) {
		return rsp, response.Unauthorized(response.ErrInvalidPassword)
	}

	jwtData := &middleware.Data{
		UserID: dataLogin.ID.Hex(),
		Email:  dataLogin.Email,
	}

	return uc.generateJWT(jwtData)
}

func (uc *usecase) Create(ctx context.Context, payload *dto.CreateUserRequest) (userID string, err error) {
	if exist := uc.repo.IsUserExist(ctx, payload.Email); exist {
		return userID, response.Conflict(response.ErrEmailAlreadyExist)
	}

	hashedPassword, err := app_crypto.Hash(payload.Password)
	if err != nil {
		return "", response.InternalServerError(err)
	}
	payload.Password = hashedPassword

	userID, err = uc.repo.SaveNewUser(ctx, entities.NewUser(payload))
	if err != nil {
		return userID, response.InternalServerError(err)
	}

	return userID, nil
}

func (uc *usecase) GenerateApiKey(ctx context.Context, userID string) (apiKey string, err error) {
	key, err := app_crypto.GenerateAPIKey(userID, uc.cfg.Authentication.SignatureKey, 32)
	if err != nil {
		return apiKey, response.InternalServerError(err)
	}

	user := &entities.User{
		ApiKey: key,
	}

	err = uc.repo.UpdateUserByID(ctx, userID, user)
	if err != nil {
		return "", response.InternalServerError(err)
	}

	return key, nil
}

func (uc *usecase) LoginViaApiKey(ctx context.Context, request *dto.UserLoginApiKeyRequest) (rsp *dto.UserLoginResponse, err error) {
	apiKey := request.ApiKey

	dataLogin, err := uc.repo.GetUserByApiKey(ctx, apiKey)
	if err != nil {
		return rsp, response.ErrUserNotFound
	}

	jwtData := &middleware.Data{
		UserID: dataLogin.ID.Hex(),
		Email:  dataLogin.Email,
	}

	return uc.generateJWT(jwtData)
}

func (uc *usecase) generateJWT(data *middleware.Data) (*dto.UserLoginResponse, error) {
	claims := middleware.PayloadToken{
		Data: data,
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

	return &dto.UserLoginResponse{AccessToken: tokenString, ExpiredAt: expiresIn}, nil
}
