package model

import "time"

type RunnerStatus string

const (
	RunnerStatusCreating RunnerStatus = "creating"
	RunnerStatusReady    RunnerStatus = "ready"
	RunnerStatusBusy     RunnerStatus = "busy"
	RunnerStatusFailed   RunnerStatus = "failed"
	RunnerStatusFinished RunnerStatus = "finished"
)

type Runner struct {
	Name        string       `json:"name"`
	ARN         string       `json:"-"`
	PrivateIPv4 string       `json:"private_ipv4"`
	Status      RunnerStatus `json:"status"`
	Metrics     Metrics      `json:"metrics"`
	UpdatedAt   time.Time    `json:"-"`
}

type Metrics map[string]float64
