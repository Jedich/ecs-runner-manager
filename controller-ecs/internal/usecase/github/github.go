package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
)

type GithubUC struct {
	credentialUC usecase.ICredentialUC
	ctx          context.Context
}

func NewCredentialUC(credentialUC usecase.ICredentialUC) usecase.IGithubUC {
	return &GithubUC{
		credentialUC: credentialUC,
		ctx:          context.Background(),
	}
}

func (c *GithubUC) GetWebhook() (*github.Hook, error) {
	credentials, err := c.credentialUC.GetCredentials()
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(credentials.GithubPAT)

	targetURL := "https://your-webhook-url"

	hooks, _, err := client.Repositories.ListHooks(c.ctx, credentials.Owner, credentials.Repo, nil)
	if err != nil {
		return nil, err
	}

	var existingHook *github.Hook
	for _, hook := range hooks {
		if *hook.Config.URL == targetURL {
			existingHook = hook
			break
		}
	}

	url := "https://your-webhook-url"
	secret := "your-webhook-secret"
	contentType := "json"

	if existingHook != nil {
		logs.Info("Retrieved existing webhook")
		return existingHook, nil
	} else {
		// Create a new webhook
		config := &github.HookConfig{
			ContentType: &contentType,
			InsecureSSL: nil,
			URL:         &url,
			Secret:      &secret,
		}
		hook := &github.Hook{
			Config: config,
			Events: []string{"workflow_job"},
			Active: github.Bool(true),
		}
		newHook, _, err := client.Repositories.CreateHook(c.ctx, credentials.Owner, credentials.Repo, hook)
		if err != nil {
			return nil, err
		}
		fmt.Println("Webhook created successfully")
		return newHook, nil
	}
}
