package usecase

import (
	metadata "github.com/brunoscheufler/aws-ecs-metadata-go"
	"github.com/google/go-github/v62/github"
	"io"
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
	CreateRunner(runner *model.Runner) (*model.Runner, error)
	GetPublicIP() string
}

type IPrometheusUC interface {
	Combine(readers map[string]io.Reader) (string, error)
	ConvertToMap(readers map[string]io.Reader) (map[string]model.Metrics, error)
}
