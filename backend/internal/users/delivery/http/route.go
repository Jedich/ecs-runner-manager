package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/middleware"
)

func (h *handlers) UserRoutes(router *gin.RouterGroup, cfg config.Config) {
	router.POST("/", h.CreateUser)
	router.POST("/login", h.Login)
	router.POST("/api-login", h.LoginViaApiKey)
	router.GET("/api-key", middleware.JWTMiddleware(cfg), h.GenerateApiKey)
}
