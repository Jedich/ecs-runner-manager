package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/internal/infrastructure/logs"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/pkg/database"
	"syscall"
	"time"

	userV1 "runner-manager-backend/internal/users/delivery/http"
	userRepository "runner-manager-backend/internal/users/repository"
	userUseCase "runner-manager-backend/internal/users/usecase"

	ctrlV1 "runner-manager-backend/internal/ctrls/delivery/http"
	ctrlRepository "runner-manager-backend/internal/ctrls/repository"
	ctrlUseCase "runner-manager-backend/internal/ctrls/usecase"

	runnersV1 "runner-manager-backend/internal/runners/delivery/http"
	runnersRepository "runner-manager-backend/internal/runners/repository"
	runnersUseCase "runner-manager-backend/internal/runners/usecase"
)

type App struct {
	client *mongo.Client // Database connection.
	gin    *gin.Engine   // Gin engine for the application.
	cfg    config.Config // Configuration settings for the application.
}

func NewApp(ctx context.Context, cfg config.Config) *App {
	client, err := database.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	return &App{
		client: client,
		gin:    middleware.NewGinServer(cfg),
		cfg:    cfg,
	}
}

func (app *App) Run() error {
	logs.NewLogger()
	apiDomain := app.gin.Group("/api")
	apiDomain.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello Word 👋")
	})

	usersColl := app.client.Database(app.cfg.Database.Name).Collection("users")
	metricsColl := app.client.Database(app.cfg.Database.Name).Collection("metrics")

	userRepo := userRepository.NewRepository(usersColl)
	userUC := userUseCase.NewUseCase(userRepo, app.cfg)
	userCTRL := userV1.NewHandlers(userUC)

	ctrlRepo := ctrlRepository.NewRepository(usersColl)
	ctrlUC := ctrlUseCase.NewUseCase(userRepo, ctrlRepo, app.cfg)
	ctrlCTRL := ctrlV1.NewHandlers(ctrlUC)

	runnersRepo := runnersRepository.NewRepository(usersColl, metricsColl)
	runnersUC := runnersUseCase.NewUseCase(userRepo, runnersRepo, app.cfg)
	runnersCTRL := runnersV1.NewHandlers(runnersUC)

	userDomain := apiDomain.Group("/users")
	userCTRL.UserRoutes(userDomain, app.cfg)

	ctrlDomain := apiDomain.Group("/ctrl")
	ctrlCTRL.CtrlRoutes(ctrlDomain, app.cfg)

	runnersDomain := apiDomain.Group("/runners")
	runnersCTRL.RunnerRoutes(runnersDomain, app.cfg)

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGINT)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", app.cfg.Server.Host, app.cfg.Server.Port),
		Handler: app.gin,
	}

	go func() {
		<-quit
		logs.Info("Server is shutting down...")

		// Create a context with a timeout of 10 seconds for the server shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown gracefully.
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		err := app.client.Disconnect(ctx)
		if err != nil {
			panic(err)
		}
	}()
	logs.InfoF("Starting server on port %s", app.cfg.Server.Port)
	return server.ListenAndServe()
}
