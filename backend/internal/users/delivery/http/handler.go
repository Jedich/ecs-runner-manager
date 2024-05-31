package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/internal/users"
	"runner-manager-backend/internal/users/dto"
	"runner-manager-backend/pkg/response"
)

type handlers struct {
	uc users.Usecase
}

func NewHandlers(uc users.Usecase) *handlers {
	return &handlers{uc}
}

func (h *handlers) CreateUser(c *gin.Context) {
	var payload *dto.CreateUserRequest
	if err := c.Bind(&payload); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	if err := payload.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	userID, err := h.uc.Create(c, payload)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(map[string]string{"id": userID}).Send(c)
}

func (h *handlers) Login(c *gin.Context) {
	var request *dto.UserLoginRequest
	if err := c.Bind(&request); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	if err := request.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	authData, err := h.uc.Login(c, request)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(authData).Send(c)
}

func (h *handlers) LoginViaApiKey(c *gin.Context) {
	var request *dto.UserLoginApiKeyRequest
	if err := c.Bind(&request); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	if err := request.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(response.ErrUserNotFound)).Send(c)
		return
	}

	authData, err := h.uc.LoginViaApiKey(c, request)
	if err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	response.SuccessBuilder(authData).Send(c)
}

func (h *handlers) GenerateApiKey(c *gin.Context) {
	userData, err := middleware.NewTokenInformation(c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	data, err := h.uc.GenerateApiKey(c, userData.Data.UserID)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(data).Send(c)
}
