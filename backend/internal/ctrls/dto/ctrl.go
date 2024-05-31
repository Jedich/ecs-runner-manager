package dto

import (
	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
)

type CreateRunnerControllerRequest struct {
	ApiKey string `json:"api_key"`
}

type CreateRunnerControllerResponse struct {
	CtrlID      string `json:"ctrl_id"`
	AccessToken string `json:"access_token"`
	ExpiredAt   int64  `json:"expired_at"`
}

func (cup *CreateRunnerControllerRequest) Validate() error {
	return validation.ValidateStruct(cup,
		validation.Field(&cup.ApiKey, validation.Required, is.ASCII, validation.Length(64, 64)),
	)
}
