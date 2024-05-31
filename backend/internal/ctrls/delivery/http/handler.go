package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/ctrls"
	"runner-manager-backend/internal/ctrls/dto"
	"runner-manager-backend/pkg/response"
)

type handlers struct {
	uc ctrls.Usecase
}

func NewHandlers(uc ctrls.Usecase) *handlers {
	return &handlers{uc}
}

func (h *handlers) RegisterCtrl(c *gin.Context) {
	var payload *dto.CreateRunnerControllerRequest
	if err := c.Bind(&payload); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	if err := payload.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	rsp, err := h.uc.Register(c, payload)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(rsp).Send(c)
}
