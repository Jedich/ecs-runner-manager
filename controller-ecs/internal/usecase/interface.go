package usecase

import (
	"github.com/google/go-github/v62/github"
	"runner-controller-ecs/internal/domain/model"
)

type ICredentialUC interface {
	GetCredentials() (*model.Credentials, error)
}

type IGithubUC interface {
	GetWebhook() (*github.Hook, error)
}

type IAWSUC interface {
	GetTaskEnvironment() (map[string]string, error)
	CreateRunner() error
}
