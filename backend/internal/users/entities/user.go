package entities

import (
	"runner-manager-backend/internal/users/dto"
	"runner-manager-backend/pkg/database"
)

type User struct {
	database.Model
	Username string `bson:"username"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

func NewUser(data *dto.CreateUserRequest) *User {
	return &User{
		Username: data.Username,
		Email:    data.Email,
		Password: data.Password,
	}
}
