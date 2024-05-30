package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/middleware"
	userV1 "runner-manager-backend/internal/users/delivery/http"
	userRepository "runner-manager-backend/internal/users/repository"
	userUseCase "runner-manager-backend/internal/users/usecase"
	"syscall"
	"time"
)

type App struct {
	//db   *sqlx.DB      // Database connection.
	gin *gin.Engine   // Gin engine for the application.
	cfg config.Config // Configuration settings for the application.
}

func NewApp(ctx context.Context, cfg config.Config) *App {
	return &App{
		gin: middleware.NewGinServer(cfg),
		cfg: cfg,
	}
}

func (app *App) Run() error {
	userRepo := userRepository.NewRepository()
	userUC := userUseCase.NewUseCase(userRepo, app.cfg)
	userCTRL := userV1.NewHandlers(userUC)

	domain := app.gin.Group("/api/v1/users")
	domain.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello Word 👋")
	})

	userCTRL.UserRoutes(domain, app.cfg)
	//if err := app.startService(); err != nil {
	//	return err
	//}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGINT)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.cfg.Server.Port),
		Handler: app.gin,
	}

	go func() {
		<-quit
		//log.Info("Server is shutting down...")

		// Create a context with a timeout of 10 seconds for the server shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown gracefully.
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		//app.db.Close()
	}()
	//log.Info("Starting server on port %s", app.cfg.Server.Port)
	return server.ListenAndServe()
}
