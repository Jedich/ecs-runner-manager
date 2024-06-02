package response

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	INTERNAL_SERVER_ERROR = "internal_server_error"
	SUCCESS               = "success"
)

// FailedResponse represents a failed response structure for API responses.
type FailedResponse struct {
	Code    int    `json:"code" example:"500"`                      // HTTP status code.
	Message string `json:"message" example:"internal_server_error"` // Message corresponding to the status code.
	Error   string `json:"error" example:"{$err}"`                  // error message.
}

// BasicResponse represents a basic response structure for API responses.
type BasicResponse struct {
	Code    int         `json:"code" example:"500"`                      // HTTP status code.
	Message string      `json:"message" example:"internal_server_error"` // Message corresponding to the status code.
	Error   string      `json:"error" example:"{$err}"`                  // error message.
	Data    interface{} `json:"data,omitempty"`
}

// BasicBuilder constructs a BasicResponse based on the provided input.
func BasicBuilder(result BasicResponse) BasicResponse {
	return result
}

// Send sends the BasicResponse as a JSON response using the provided Gin context.
func (c BasicResponse) Send(ctx *gin.Context) {
	ctx.JSON(c.Code, c)
}

// ErrorBuilder constructs a FailedResponse based on the provided error.
func ErrorBuilder(err error) FailedResponse {
	var appErr *AppError
	if errors.As(err, &appErr) {
		var ae *AppError
		errors.As(err, &ae)

		return FailedResponse{
			Code:    ae.Code,
			Message: ae.Message,
			Error:   ae.Error(),
		}
	}

	var errString = INTERNAL_SERVER_ERROR
	if err != nil {
		errString = err.Error()
	}

	return FailedResponse{
		Code:    http.StatusInternalServerError,
		Message: INTERNAL_SERVER_ERROR,
		Error:   errString,
	}
}

// Send sends the FailedResponse as a JSON response using the provided Gin context.
func (x FailedResponse) Send(ctx *gin.Context) {
	fmt.Println(x)
	ctx.JSON(x.Code, x)
}

// SuccessResponse represents a success response structure for API responses.
type SuccessResponse struct {
	Success
	Meta
}

type ResponseFormat struct {
	Code    int    `json:"code" example:"200"` // HTTP status code.
	Message string `json:"message" example:"success"`
}

type Success struct {
	ResponseFormat
	Data interface{} `json:"data,omitempty"` // data payload.
}

type Meta struct {
	Meta interface{} `json:"meta,omitempty"` // pagination payload.
	Success
}

// SuccessBuilder constructs a SuccessResponse with a Success status and the provided response data.
func SuccessBuilder(response interface{}, meta ...interface{}) SuccessResponse {
	result := SuccessResponse{
		Success: Success{
			ResponseFormat: ResponseFormat{
				Code:    http.StatusOK,
				Message: SUCCESS,
			},
			Data: response,
		},
	}

	if len(meta) > 0 {
		result.Meta.Meta = meta[0]
	}

	return result
}

// Send sends the SuccessResponse as a JSON response using the provided Gin context.
func (c SuccessResponse) Send(ctx *gin.Context) {
	ctx.JSON(c.Code, c)
}
