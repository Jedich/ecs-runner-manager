package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/users"
)

type handlers struct {
	uc users.Usecase
}

func NewHandlers(uc users.Usecase) *handlers {
	return &handlers{uc}
}

func (h *handlers) CreateUser(c *gin.Context) {
}

func (h *handlers) Login(c *gin.Context) {
}

func (h *handlers) GenerateApiKey(c *gin.Context) {

}
