package credentials

import (
	"os"
	"runner-controller-ecs/internal/domain"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/usecase"
	"strings"
)

type CredentialUC struct {
	credentials *model.Credentials
}

func NewCredentialUC() usecase.ICredentialUC {
	return &CredentialUC{}
}

func (c *CredentialUC) GetCredentials() (*model.Credentials, error) {
	if c.credentials == nil {
		if !strings.Contains(os.Getenv("REPO"), "/") {
			return nil, domain.ErrInvalidRepoFormat
		}

		repoOwner := strings.Split(os.Getenv("REPO"), "/")
		c.credentials = &model.Credentials{
			Owner:              repoOwner[0],
			Repo:               repoOwner[1],
			GithubPAT:          os.Getenv("GITHUB_PAT"),
			AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			AWSRegion:          os.Getenv("AWS_REGION"),
		}
	}
	return c.credentials, nil
}
