package constant

type RunnerStatus string

const (
	RunnerStatusReady    RunnerStatus = "ready"
	RunnerStatusBusy     RunnerStatus = "busy"
	RunnerStatusFailed   RunnerStatus = "failed"
	RunnerStatusFinished RunnerStatus = "finished"
)
