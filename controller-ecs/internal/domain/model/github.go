package model

type WorkflowJobWebhook struct {
	Action string       `json:"action"`
	Job    *workflowJob `json:"workflow_job"`
}

type workflowJob struct {
	Labels []string `json:"labels"`
}
