package dto

import (
	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
	"runner-manager-backend/pkg/constant"
)

type CreateRunnerRequest struct {
	Name        string                `json:"name"`
	PrivateIPv4 string                `json:"private_ipv4"`
	Status      constant.RunnerStatus `json:"status"`
}

type UpdateRunnersRequest struct {
	Runners []UpdateRunnerRequest `json:"runners"`
}

type UpdateRunnerRequest struct {
	Name        string                   `json:"name"`
	PrivateIPv4 string                   `json:"private_ipv4"`
	Status      constant.RunnerStatus    `json:"status"`
	Metrics     []map[string]interface{} `json:"metrics"`
}

type RunnerControllerWSResponse struct {
	Id                string              `json:"id"`
	Name              string              `json:"name"`
	RunnersWSResponse []*RunnerWSResponse `json:"runners"`
}

type RunnerWSResponse struct {
	Id          string                   `json:"id"`
	Name        string                   `json:"name"`
	PrivateIPv4 string                   `json:"private_ipv4"`
	Status      constant.RunnerStatus    `json:"status"`
	Metrics     []map[string]interface{} `json:"metrics"`
}

type MetricsCtrlWSResponse struct {
	RunnerMetrics []*MetricsRunnerWSResponse `json:"runners"`
}

type MetricsRunnerWSResponse struct {
	Name          string               `json:"name"`
	RunnerMetrics []*MetricsWSResponse `json:"metrics"`
}

type MetricsWSResponse struct {
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (cup *UpdateRunnersRequest) Validate() error {
	if len(cup.Runners) == 0 {
		return nil
	}
	for _, runner := range cup.Runners {
		err := validation.ValidateStruct(&runner,
			validation.Field(&runner.PrivateIPv4, validation.Required, is.IPv4),
			validation.Field(&runner.Name, validation.Required),
			validation.Field(&runner.Status, validation.Required, validation.In(
				constant.RunnerStatusCreating,
				constant.RunnerStatusReady,
				constant.RunnerStatusBusy,
				constant.RunnerStatusFailed,
				constant.RunnerStatusFinished),
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
