package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"runner-manager-backend/internal/ctrls/entities"
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

	RunnerController []*entities.RunnerController `bson:"ctrls"`
}

func NewUser(data *dto.CreateUserRequest) *User {
	return &User{
		Username:         data.Username,
		Email:            data.Email,
		Password:         data.Password,
		RunnerController: make([]*entities.RunnerController, 0),
		CreatedAt:        time.Now(),
	}
}
