package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"runner-manager-backend/internal/users/dto"
	"time"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	ApiKey    string             `bson:"api_key"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func NewUser(data *dto.CreateUserRequest) *User {
	return &User{
		Username:  data.Username,
		Email:     data.Email,
		Password:  data.Password,
		CreatedAt: time.Now(),
	}
}
