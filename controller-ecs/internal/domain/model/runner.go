package model

type RunnerStatus string

const (
	RunnerStatusReady    RunnerStatus = "ready"
	RunnerStatusBusy     RunnerStatus = "busy"
	RunnerStatusFailed   RunnerStatus = "failed"
	RunnerStatusFinished RunnerStatus = "finished"
)

type Runner struct {
	Name        string       `json:"name"`
	ARN         string       `json:"-"`
	PrivateIPv4 string       `json:"private_ip"`
	Status      RunnerStatus `json:"status"`
	Metrics     Metrics      `json:"metrics"`
}

type Metrics map[string]float64
