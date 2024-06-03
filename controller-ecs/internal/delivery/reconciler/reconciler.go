package reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runner-controller-ecs/internal/delivery"
	"runner-controller-ecs/internal/domain/model"
	"runner-controller-ecs/internal/infrastructure/logs"
	"runner-controller-ecs/internal/tools"
	"runner-controller-ecs/internal/usecase"
	"runner-controller-ecs/internal/usecase/broker"
	"runner-controller-ecs/internal/usecase/credentials"
	gh "runner-controller-ecs/internal/usecase/github"
	"runner-controller-ecs/internal/usecase/prometheus"
	"syscall"
	"time"
)

type Reconciler struct {
	awsUC         usecase.IAWSUC
	credentialsUC usecase.ICredentialUC
	promUC        usecase.IPrometheusUC
	name          string

	broker *broker.Broker[model.WorkflowJobWebhook]

	runners map[string]*model.Runner
	jwt     string
}

func NewReconciler(awsUC usecase.IAWSUC, broker *broker.Broker[model.WorkflowJobWebhook]) delivery.Reconciler {
	return &Reconciler{
		broker:  broker,
		awsUC:   awsUC,
		runners: make(map[string]*model.Runner),
	}
}

const (
	TerminatedDeregTimeout = 2 * time.Minute
	CompletedDeregTimeout  = 1 * time.Minute
)

func (c *Reconciler) SubscribeBroker() chan model.WorkflowJobWebhook {
	return c.broker.Subscribe()
}

func (c *Reconciler) Init() error {
	c.name = "controller-" + tools.RandString(6)
	c.runners = make(map[string]*model.Runner)
	c.credentialsUC = credentials.NewCredentialUC()
	c.promUC = prometheus.NewPrometheusUC()

	logs.InfoF("Controller name: %s", c.name)

	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return err
	}

	// Define the POST data
	var jsonStr = map[string]interface{}{
		"name":    c.name,
		"api_key": creds.ApiKey,
	}

	jsonData, err := json.Marshal(jsonStr)
	if err != nil {
		return err
	}
	url := creds.BackendURL

	// Make the POST request
	response, err := http.Post(url+"/api/ctrl/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Error(err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logs.Error(errors.New(fmt.Sprintf("error: %s", response.StatusCode)))
		return nil
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Error(err)
		return nil
	}

	var rsp model.AuthResponse
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		logs.Error(err)
		return nil
	}
	c.jwt = rsp.Data.AccessToken

	// Print the response body
	logs.Info(string(body))

	if c.jwt == "" {
		logs.Error(errors.New("no jwt token found"))
		return nil
	}

	_, err = c.awsUC.GetTaskMetadata()
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
					newRunner := &model.Runner{
						Name:        "linux-" + tools.RandString(6),
						Status:      model.RunnerStatusCreating,
						PrivateIPv4: "0.0.0.0",
						Metrics:     map[string]float64{},
					}
					c.runners[newRunner.Name] = newRunner
					err := c.SendRunners()
					if err != nil {
						logs.ErrorF("Error sending initial runner: %w", err)
					}
					go func() {
						runner, err := c.awsUC.CreateRunner(newRunner)
						if err != nil {
							logs.Error(err)
						}
						err = c.SendRunners()
						if err != nil {
							logs.ErrorF("Error sending idle runner: %w", err)
						}
						logs.InfoF("%v", runner)

					}()
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
			if _, ok := c.runners[data.Job.RunnerName]; !ok {
				return nil
			}
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusBusy
			c.runners[data.Job.RunnerName].UpdatedAt = time.Now()
		case "completed":
			if _, ok := c.runners[data.Job.RunnerName]; !ok {
				return nil
			}
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusFinished
			c.runners[data.Job.RunnerName].Metrics = map[string]float64{}
			c.runners[data.Job.RunnerName].UpdatedAt = time.Now()
		case "failed":
			if _, ok := c.runners[data.Job.RunnerName]; !ok {
				return nil
			}
			c.runners[data.Job.RunnerName].Status = model.RunnerStatusFailed
			c.runners[data.Job.RunnerName].UpdatedAt = time.Now()
		}

		err := c.FetchMetrics()
		if err != nil {
			return err
		}

		err = c.SendRunners()
		if err != nil {
			return err
		}
	default:
		err := c.reconcileDefault()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Reconciler) reconcileDefault() error {
	err := c.FetchMetrics()
	if err != nil {
		return err
	}

	err = c.SendRunners()
	if err != nil {
		return err
	}

	return nil
}

func (c *Reconciler) FetchMetrics() error {
	readers := make(map[string]io.Reader)
	for name, runner := range c.runners {
		if runner == nil {
			logs.Info("Runner is nil. Skipping...")
			continue
		}
		if runner.Status == model.RunnerStatusFinished || runner.Status == model.RunnerStatusFailed {
			//logs.InfoF("Runner %s is in status %s. Skipping...", runner.Name, runner.Status)
			if runner.UpdatedAt != (time.Time{}) && runner.UpdatedAt.Add(CompletedDeregTimeout).Before(time.Now()) {
				runner.Metrics = map[string]float64{}
				runner.Status = model.RunnerStatusTerminated
				logs.InfoF("Runner %s terminated", runner.Name)
			}

			continue
		}
		if runner.Status == model.RunnerStatusTerminated {
			if runner.UpdatedAt != (time.Time{}) && runner.UpdatedAt.Add(TerminatedDeregTimeout).Before(time.Now()) {
				delete(c.runners, name)
				logs.InfoF("Runner %s deleted", runner.Name)
			}
			continue
		}
		cli := http.DefaultClient
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		logs.InfoF("Fetching metrics from runner %s with status %s", runner.PrivateIPv4, runner.Status)
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:9779/metrics", runner.PrivateIPv4), nil)
		if err != nil {
			return err
		}

		res, err := cli.Do(req.WithContext(ctx))
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				if _, ok := c.runners[name]; !ok {
					return nil
				}
				c.runners[name].Status = model.RunnerStatusReady
				logs.InfoF("Metrics request timeout %s", runner.Name)
				continue
			case errors.Is(err, syscall.ECONNREFUSED):
				//logs.InfoF("Runner %s is not ready yet. Skipping...", runner.Name)
				continue
			case errors.Is(err, syscall.EHOSTUNREACH):
				if _, ok := c.runners[name]; !ok {
					return nil
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
		if c.runners[k].Status == model.RunnerStatusFinished || c.runners[k].Status == model.RunnerStatusTerminated {
			c.runners[k].Metrics = map[string]float64{}
			continue
		}
		c.runners[k].Metrics = v
	}

	return nil
}

func (c *Reconciler) SendRunners() error {
	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return err
	}
	url := creds.BackendURL

	rq := &model.ControllerRequest{
		Name:    c.name,
		Runners: make([]*model.RequestRunner, 0, len(c.runners)),
	}
	for _, runner := range c.runners {
		m := make([]model.Metrics, 0, 1)
		if len(runner.Metrics) > 0 {
			m = append(m, runner.Metrics)
		}
		if runner.Status == model.RunnerStatusTerminated || runner.Status == model.RunnerStatusFinished || runner.Status == model.RunnerStatusFailed {
			m = []model.Metrics{}
		}
		rq.Runners = append(rq.Runners, &model.RequestRunner{
			Name:        runner.Name,
			PrivateIPv4: runner.PrivateIPv4,
			Status:      runner.Status,
			Metrics:     m,
		})
	}

	m, err := json.MarshalIndent(rq, "", "  ")
	if err != nil {
		logs.Error(err)
		return nil
	}

	logs.InfoF("Runners to be sent: %s", string(m))

	jsonData, err := json.Marshal(rq)
	if err != nil {
		logs.Error(err)
		return nil
	}

	req, err := http.NewRequest("POST", url+"/api/runners/", bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Error(err)
		return nil
	}

	if c.jwt == "" {
		logs.Error(errors.New("no jwt token found"))
		return nil
	}

	// Add the Content-Type and Authorization headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.jwt))

	// Perform the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logs.Error(err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logs.Error(errors.New(fmt.Sprintf("error: %s", response.StatusCode)))
		return nil
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logs.Error(err)
		return nil
	}

	logs.Info(string(body))

	return nil
}
