package constant

type RunnerStatus string

const (
	RunnerStatusCreating   RunnerStatus = "creating"
	RunnerStatusReady      RunnerStatus = "ready"
	RunnerStatusBusy       RunnerStatus = "busy"
	RunnerStatusFailed     RunnerStatus = "failed"
	RunnerStatusFinished   RunnerStatus = "finished"
	RunnerStatusTerminated RunnerStatus = "terminated"
)
