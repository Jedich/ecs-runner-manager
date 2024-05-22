package http

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase/broker"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

const PORT = 80

// LoggerMiddleware is a middleware function to log incoming requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Call the next handler
		c.Next()

		logs.InfoF("%s %s %d {%s}",
			c.Request.Method,
			c.Request.RequestURI,
			c.Writer.Status(),
			blw.body.String(),
		)
	}
}

func StartWebhookServer(broker *broker.Broker[model.WorkflowJobWebhook]) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Apply the logger middleware
	router.Use(LoggerMiddleware())

	// Define a route to receive webhook events
	router.POST("/ecs_runner_hook", func(c *gin.Context) {
		// Parse the webhook payload
		var payload model.WorkflowJobWebhook
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook payload"})
			return
		}

		broker.Publish(payload)

		c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
	})

	// Run the HTTP server in a Goroutine
	go func() {
		if err := router.Run(fmt.Sprintf(":%d", PORT)); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()
	logs.InfoF("Launched GIN to listen to Github Webhook requests at port :%d", PORT)
}
