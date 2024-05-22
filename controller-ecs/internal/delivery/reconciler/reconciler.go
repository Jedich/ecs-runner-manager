package reconciler

import (
	"fmt"
	"io"
	"net/http"
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/usecase"
	"runner-controller-ecs/internal/usecase/broker"
	"runner-controller-ecs/internal/usecase/credentials"
	gh "runner-controller-ecs/internal/usecase/github"
	"strings"
)

type Reconciler struct {
	awsUC usecase.IAWSUC

	broker *broker.Broker[model.WorkflowJobWebhook]

	runners map[string]*model.Runner
}

func NewReconciler(awsUC usecase.IAWSUC, broker *broker.Broker[model.WorkflowJobWebhook]) delivery.Reconciler {
	return &Reconciler{
		broker:  broker,
		awsUC:   awsUC,
		runners: make(map[string]*model.Runner),
	}
}

func (c *Reconciler) SubscribeBroker() chan model.WorkflowJobWebhook {
	return c.broker.Subscribe()
}

func (c *Reconciler) Init() error {
	creds := credentials.NewCredentialUC()

	_, err := c.awsUC.GetTaskMetadata()
	if err != nil {
		return err
	}

	githubUC := gh.NewGithubUC(creds)

	_, err = githubUC.GetWebhook(c.awsUC.GetPublicIP())
	if err != nil {
		return err
	}

	return nil
}

func (c *Reconciler) Reconcile(brokerChannel chan model.WorkflowJobWebhook) error {
	select {
	case data := <-brokerChannel:
		if data.Action == "" || data.Job == nil {
			logs.Info("Webhook received, but no action or job data found. Skipping...")
			return nil
		}

		logs.InfoF("Received webhook data: %v", data)

		if data.Action != "queued" {
			logs.Info("Job is not queued. Skipping...")
			return nil
		}
		for _, label := range data.Job.Labels {
			if label == "self-hosted" {
				runner, err := c.awsUC.CreateRunner()
				if err != nil {
					return err
				}

				for _, r := range runner {
					c.runners[r.MetricsPrivateIP] = r
					logs.InfoF("%v", r)
				}
			}
		}
		return nil
	default:
		for _, runner := range c.runners {
			requestURL := fmt.Sprintf("http://%s:9779/metrics", runner.MetricsPrivateIP)
			res, err := http.Get(requestURL)
			if err != nil {
				return fmt.Errorf("error making http request: %s", err)
			}

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("client: could not read response body: %s", err)
			}
			logs.InfoF("Runner: %s", runner.MetricsPrivateIP)
			logs.InfoF("Metrics: %s", strings.ReplaceAll(string(resBody), "\n", " "))
		}
		logs.Info("Reconcile AFK...")
	}
	return nil
}
