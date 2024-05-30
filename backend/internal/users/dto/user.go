package dto

import "github.com/invopop/validation"

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	UserID    int64  `json:"user_id"`
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expired_at"`
}

func (cup CreateUserRequest) Validate() error {
	return validation.ValidateStruct(&cup,
		validation.Field(&cup.Username, validation.Required, validation.Length(0, 50)),
		validation.Field(&cup.Email, validation.Required),
		validation.Field(&cup.Password, validation.Required),
	)
}
