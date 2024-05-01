package usecase

import "runner-controller-ecs/internal/domain/model"

type Credentializer interface {
	GetCredentials() (*model.Credentials, error)
}
