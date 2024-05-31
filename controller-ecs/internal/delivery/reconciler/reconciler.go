package reconciler

import (
	"bytes"
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
	jwt     string
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
	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return err
	}

	// Define the POST data
	var jsonStr = map[string]interface{}{
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

	creds, err := c.credentialsUC.GetCredentials()
	if err != nil {
		return err
	}
	url := creds.BackendURL

	rq := &model.ControllerRequest{
		Runners: make([]*model.RequestRunner, 0, len(c.runners)),
	}
	for _, runner := range c.runners {
		rq.Runners = append(rq.Runners, &model.RequestRunner{
			Name:        runner.Name,
			PrivateIPv4: runner.PrivateIPv4,
			Status:      runner.Status,
			Metrics:     []model.Metrics{runner.Metrics},
		})
	}

	m, err := json.MarshalIndent(rq, "", "  ")
	if err != nil {
		logs.Error(err)
		return nil
	}

	logs.InfoF(string(m))

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
