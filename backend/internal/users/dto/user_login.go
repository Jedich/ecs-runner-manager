package dto

import (
	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
)

type (
	UserLoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	UserLoginApiKeyRequest struct {
		ApiKey string `json:"api_key"`
	}

	UserLoginResponse struct {
		AccessToken string `json:"access_token"`
		ExpiredAt   int64  `json:"expired_at"`
	}
)

func (ulr UserLoginRequest) Validate() error {
	return validation.ValidateStruct(&ulr,
		validation.Field(&ulr.Email, validation.Required),
		validation.Field(&ulr.Password, validation.Required),
	)
}

func (ulr UserLoginApiKeyRequest) Validate() error {
	return validation.ValidateStruct(&ulr,
		validation.Field(&ulr.ApiKey, validation.Required, is.ASCII, validation.Length(64, 64)),
	)
}
