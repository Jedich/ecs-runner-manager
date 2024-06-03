package http

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"runner-manager-backend/internal/infrastructure/logs"
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
var tokenToConn = make(map[string]*UserWS)

type UserWS struct {
	Data       *RequestWSData
	Connection *websocket.Conn
}

type RequestWSData struct {
	CtrlID string `json:"ctrl_id"`
	Event  string `json:"event"`
}

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

	_, err = h.uc.UpdateRunners(c, userData.Data.UserID, userData.Data.CtrlID, payload)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	userWS, ok := tokenToConn[userData.Data.UserID]
	if !ok {
		response.SuccessBuilder(nil).Send(c)
		return
	}

	err = h.updateControllersWS(c, userData)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	err = h.updateMetricsWS(userData, userWS.Data, c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}

	response.SuccessBuilder(nil).Send(c)
}

func (h *handlers) updateControllersWS(c *gin.Context, userData *middleware.PayloadToken) error {
	ctrls, err := h.uc.GetAllCtrlsByUserID(c, userData.Data.UserID)
	if err != nil {
		return err
	}

	res := map[string]interface{}{
		"event": "ctrls",
		"data":  ctrls,
	}

	if len(ctrls) == 0 {
		res["empty"] = true
	}

	jsonData, err := json.Marshal(&res)
	if err != nil {
		return err
	}

	logs.Info("Sending data to user")

	sendToUser(userData.Data.UserID, jsonData)

	return nil
}

func (h *handlers) updateMetricsWS(userData *middleware.PayloadToken, data *RequestWSData, c *gin.Context) error {
	if data.CtrlID == "" {
		logs.InfoF("CtrlID is missing, skipping")
		return nil
	}

	metrics, err := h.uc.GetAllMetricsByCtrlID(c, userData.Data.UserID, data.CtrlID)
	if err != nil {
		return err
	}

	res := map[string]interface{}{
		"event": "metrics",
		"data":  metrics,
	}

	jsonData, err := json.Marshal(&res)
	if err != nil {
		return err
	}
	logs.Info("Sending data to user")

	sendToUser(userData.Data.UserID, jsonData)

	return nil
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
	tokenPayload, err := middleware.NewTokenInformation(c)
	if err != nil {
		response.ErrorBuilder(err).Send(c)
		return
	}
	fmt.Println("User ID:", tokenPayload.Data.UserID)

	user := &UserWS{
		Data:       &RequestWSData{},
		Connection: conn,
	}
	// Associate WebSocket connection with JWT token
	tokenToConn[tokenPayload.Data.UserID] = user
	defer delete(tokenToConn, tokenPayload.Data.UserID) // Remove the entry from the map when the connection closes

	// Handle WebSocket connection
	for {
		// Read message from WebSocket client
		_, message, err := conn.ReadMessage()
		if err != nil {
			logs.InfoF("Read error: ", err)
			break
		}

		data := &RequestWSData{}

		if err = json.Unmarshal(message, &data); err != nil {
			logs.InfoF("Error unmarshalling message:", err)
			return
		}

		if _, ok := tokenToConn[tokenPayload.Data.UserID]; ok {
			tokenToConn[tokenPayload.Data.UserID].Data = data
		}

		if data.Event != "" {
			switch data.Event {
			case "metrics":
				err = h.updateMetricsWS(tokenPayload, data, c)
				if err != nil {
					logs.InfoF("Error updating metrics:", err)
					return
				}
			case "ctrls":
				err = h.updateControllersWS(c, tokenPayload)
				if err != nil {
					logs.InfoF("Error updating controllers:", err)
					return
				}
			}
		}

		logs.InfoF("Received message: %s", message)
	}
}

// Send data to the user associated with the JWT token
func sendToUser(userID string, data []byte) {
	// Get WebSocket connection associated with the JWT token
	user, ok := tokenToConn[userID]
	if !ok {
		fmt.Println("No WebSocket connection found for token:", userID)
		return
	}

	// Send the data to the user's WebSocket connection
	if err := user.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("Error sending data to user:", err)
	}
}
