package http

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/middleware"
)

func (h *handlers) UserRoutes(router *gin.RouterGroup, cfg config.Config) {
	router.POST("/", h.CreateUser)
	router.POST("/login", h.Login)
	router.GET("/api-key", h.GenerateApiKey, middleware.JWTMiddleware(cfg))
}
