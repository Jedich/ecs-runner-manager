package model

type WorkflowJobWebhook struct {
	Action string       `json:"action"`
	Job    *workflowJob `json:"workflow_job"`
}

type workflowJob struct {
	RunnerName string   `json:"runner_name"`
	Labels     []string `json:"labels"`
}
