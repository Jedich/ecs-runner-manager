package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/config"
)

func (h *handlers) CtrlRoutes(router *gin.RouterGroup, cfg config.Config) {
	router.POST("/", h.RegisterCtrl)
}
