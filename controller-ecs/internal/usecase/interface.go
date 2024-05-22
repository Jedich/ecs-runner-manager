package usecase

import (
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"
	"github.com/google/go-github/v62/github"
	"runner-controller-ecs/internal/domain/model"
)

type ICredentialUC interface {
	GetCredentials() (*model.Credentials, error)
}

type IGithubUC interface {
	GetWebhook(ip string) (*github.Hook, error)
}

type IAWSUC interface {
	GetTaskMetadata() (*metadata.TaskMetadataV4, error)
	CreateRunner() ([]*model.Runner, error)
	GetPublicIP() string
}
