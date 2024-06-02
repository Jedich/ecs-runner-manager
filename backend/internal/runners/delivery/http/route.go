package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/middleware"
)

func (h *handlers) RunnerRoutes(router *gin.RouterGroup, cfg config.Config) {
	router.POST("/", middleware.JWTMiddleware(cfg), h.UpdateRunners)
	router.GET("/ws", middleware.JWTMiddleware(cfg), h.WsCtrl)
}
