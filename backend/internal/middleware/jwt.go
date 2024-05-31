package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"runner-manager-backend/internal/config"
	"runner-manager-backend/pkg/response"
	"runner-manager-backend/pkg/utils"
)

type PayloadToken struct {
	Data *Data `json:"data"`
	jwt.StandardClaims
}

type Data struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// JWTMiddleware Middleware function to validate JWT token
func JWTMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			response.ErrorBuilder(response.Unauthorized(errors.New("Missing JWT token"))).Send(c)
			c.Abort()
			return
		}

		// Remove "Bearer " prefix from token string
		tokenString = tokenString[len("Bearer "):]

		token, err := jwt.ParseWithClaims(tokenString, &PayloadToken{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Key), nil
		})

		if err != nil {
			response.ErrorBuilder(response.Unauthorized(errors.New("invalid JWT token"))).Send(c)
			c.Abort()
			return
		}
		if !token.Valid {
			response.ErrorBuilder(response.Unauthorized(errors.New("JWT token is not valid"))).Send(c)
			c.Abort()
			return
		}

		// Store the token claims in the request context for later use
		claims := token.Claims.(*PayloadToken)
		c.Set(utils.AuthCtxKey, claims)

		c.Next()
	}
}

func NewTokenInformation(c *gin.Context) (*PayloadToken, error) {
	value := c.MustGet(utils.AuthCtxKey)
	if value == nil {
		return nil, response.Unauthorized(response.ErrFailedGetTokenInformation)
	}

	tokenInformation, ok := value.(*PayloadToken)
	if !ok {
		return nil, response.Unauthorized(errors.New("invalid token information type"))
	}

	return tokenInformation, nil
}
