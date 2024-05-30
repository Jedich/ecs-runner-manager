package middleware

import (
	"github.com/gin-gonic/gin"
	"runner-manager-backend/internal/config"
)

func NewGinServer(cfg config.Config) *gin.Engine {
	router := gin.Default()
	// Add your middlewares here, e.g., CORS, Logging, etc.
	return router
}
