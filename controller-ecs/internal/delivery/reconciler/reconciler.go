package reconciler

import (
	"encoding/json"
	"errors"
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
	"runner-controller-ecs/internal/usecase/prometheus"
	"syscall"
)

type Reconciler struct {
	awsUC         usecase.IAWSUC
	credentialsUC usecase.ICredentialUC
	promUC        usecase.IPrometheusUC

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
	c.runners = make(map[string]*model.Runner)
	c.credentialsUC = credentials.NewCredentialUC()
	c.promUC = prometheus.NewPrometheusUC()

	_, err := c.awsUC.GetTaskMetadata()
	if err != nil {
		return err
	}

	githubUC := gh.NewGithubUC(c.credentialsUC)

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

		switch data.Action {
		case "queued":
			for _, label := range data.Job.Labels {
				if label == "self-hosted" {
					runner, err := c.awsUC.CreateRunner()
					if err != nil {
						return err
					}

					c.runners[runner.Name] = runner
					logs.InfoF("%v", runner)
				}
			}
		default:
			logs.InfoF("Runner assigned to job: '%s'", data.Job.RunnerName)
			if _, ok := c.runners[data.Job.RunnerName]; !ok {
				logs.InfoF("Runner %s not found. Skipping...", data.Job.RunnerName)
				return nil
			}
			fallthrough
		case "in_progress":
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusBusy
		case "completed":
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusFinished
		case "failed":
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusFailed
		}

		return nil
	default:
		err := c.reconcileDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Reconciler) reconcileDefault() error {
	readers := make(map[string]io.Reader)
	for name, runner := range c.runners {
		if runner == nil {
			logs.Info("Runner is nil. Skipping...")
			continue
		}
		if runner.Status == "" || runner.Status == model.RunnerStatusFinished || runner.Status == model.RunnerStatusFailed {
			logs.InfoF("Runner %s is in status %s. Skipping...", runner.Name, runner.Status)
			continue
		}

		requestURL := fmt.Sprintf("http://%s:9779/metrics", runner.PrivateIPv4)
		res, err := http.Get(requestURL)
		if err != nil {
			switch {
			case errors.Is(err, syscall.ECONNREFUSED):
				//logs.InfoF("Runner %s is not ready yet. Skipping...", runner.Name)
				continue
			case errors.Is(err, syscall.EHOSTUNREACH):
				if _, ok := c.runners[name]; !ok {
					//logs.InfoF("Runner %s not found. Skipping...", runner.Name)
					test, err := json.MarshalIndent(c.runners, "", "  ")
					if err != nil {
						return err
					}
					logs.InfoF(string(test))
				}
				c.runners[name].Status = model.RunnerStatusFinished
				logs.InfoF("Runner %s terminated", runner.Name)
				continue
			default:
				logs.ErrorF("error making http request: %s", err)
				continue
			}
		}

		readers[name] = res.Body
	}

	toMap, err := c.promUC.ConvertToMap(readers)
	if err != nil {
		return err
	}

	if len(toMap) == 0 {
		return nil
	}

	//logs.InfoF("Received metrics")

	for k, v := range toMap {
		if _, ok := c.runners[k]; !ok {
			continue
		}
		c.runners[k].Metrics = v
	}

	m, err := json.MarshalIndent(c.runners, "", "  ")
	if err != nil {
		return err
	}

	logs.InfoF(string(m))
	return nil
}
