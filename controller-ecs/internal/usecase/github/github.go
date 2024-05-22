package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"math/rand"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
	"strings"
)

type GithubUC struct {
	credentialUC  usecase.ICredentialUC
	webhookSecret string
	ctx           context.Context
}

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewGithubUC(credentialUC usecase.ICredentialUC) usecase.IGithubUC {
	return &GithubUC{
		credentialUC:  credentialUC,
		webhookSecret: randString(16),
		ctx:           context.Background(),
	}
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func (c *GithubUC) GetWebhook(ip string) (*github.Hook, error) {
	credentials, err := c.credentialUC.GetCredentials()
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(credentials.GithubPAT)

	targetEndpoint := "ecs_runner_hook"

	hooks, _, err := client.Repositories.ListHooks(c.ctx, credentials.Owner, credentials.Repo, nil)
	if err != nil {
		return nil, err
	}

	var existingHook *github.Hook
	for _, hook := range hooks {
		if strings.Contains(*hook.Config.URL, targetEndpoint) {
			existingHook = hook
			break
		}
	}

	url := fmt.Sprintf("http://%s/%s", ip, targetEndpoint)
	contentType := "json"

	if existingHook != nil {
		_, err = client.Repositories.DeleteHook(c.ctx, credentials.Owner, credentials.Repo, *existingHook.ID)
		if err != nil {
			return nil, err
		}

		logs.Info("Deleted existing webhook")
		return existingHook, nil
	}
	// Create a new webhook
	config := &github.HookConfig{
		ContentType: &contentType,
		InsecureSSL: nil,
		URL:         &url,
		Secret:      &c.webhookSecret,
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

	logs.Info("Webhook created successfully")
	return newHook, nil
}
