package credentials

import (
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/usecase"
)

type Credentializer struct {
}

func NewCredentializer() usecase.Credentializer {
	return &Credentializer{}
}

func (c *Credentializer) GetCredentials() (*model.Credentials, error) {
	//TODO implement me
	panic("implement me")
}
