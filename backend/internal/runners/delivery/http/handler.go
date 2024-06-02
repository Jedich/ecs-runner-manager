package http

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"runner-manager-backend/internal/middleware"
	"runner-manager-backend/internal/runners"
	"runner-manager-backend/internal/runners/dto"
	"runner-manager-backend/pkg/response"
)

type handlers struct {
	uc runners.Usecase
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Mapping of JWT tokens to WebSocket connections
var tokenToConn = make(map[string]*websocket.Conn)

func NewHandlers(uc runners.Usecase) *handlers {
	return &handlers{uc}
}

func (h *handlers) UpdateRunners(c *gin.Context) {
	var payload *dto.UpdateRunnersRequest
	if err := c.Bind(&payload); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	userData, err := middleware.NewTokenInformation(c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}
	if userData.Data.CtrlID == "" {
		response.ErrorBuilder(response.Unauthorized(response.ErrFailedGetTokenInformation)).Send(c)
		return
	}

	if err := payload.Validate(); err != nil {
		response.ErrorBuilder(response.BadRequest(err)).Send(c)
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		fmt.Println("JWT token is missing")
		return
	}

	ctrls, err := h.uc.UpdateRunners(c, userData.Data.UserID, userData.Data.CtrlID, payload)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}
	sendToUser(userData.Data.UserID, ctrls)

	response.SuccessBuilder(nil).Send(c)
}

func (h *handlers) WsCtrl(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Validate JWT token (implement your token validation logic here)
	// For simplicity, let's assume a function validateToken(token string) bool
	userData, err := middleware.NewTokenInformation(c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}
	fmt.Println("User ID:", userData.Data.UserID)

	// Associate WebSocket connection with JWT token
	tokenToConn[userData.Data.UserID] = conn
	defer delete(tokenToConn, userData.Data.UserID) // Remove the entry from the map when the connection closes

	// Handle WebSocket connection
	for {
		// Read message from WebSocket client
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
	}
}

// Send data to the user associated with the JWT token
func sendToUser(userID string, data []*dto.RunnerControllerWSResponse) {
	// Get WebSocket connection associated with the JWT token
	conn, ok := tokenToConn[userID]
	if !ok {
		fmt.Println("No WebSocket connection found for token:", userID)
		return
	}

	jsonData, err := json.Marshal(&data)
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return
	}
	fmt.Println("Sending data to user:", string(jsonData))

	// Send the data to the user's WebSocket connection
	if err = conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		fmt.Println("Error sending data to user:", err)
	}
}
