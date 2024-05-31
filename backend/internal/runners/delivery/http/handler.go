package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/internal/runners"
	"runner-manager-backend/internal/runners/dto"
	"runner-manager-backend/pkg/response"
)

type handlers struct {
	uc runners.Usecase
}

func NewHandlers(uc runners.Usecase) *handlers {
	return &handlers{uc}
}

func (h *handlers) UpdateRunners(c *gin.Context) {
	var payload *dto.UpdateRunnersRequest
	if err := c.Bind(&payload); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	userData, err := middleware.NewTokenInformation(c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}
	if userData.Data.CtrlID == "" {
		response.ErrorBuilder(response.Unauthorized(response.ErrFailedGetTokenInformation)).Send(c)
		return
	}

	if err := payload.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	err = h.uc.UpdateRunners(c, userData.Data.UserID, userData.Data.CtrlID, payload)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(nil).Send(c)
}
