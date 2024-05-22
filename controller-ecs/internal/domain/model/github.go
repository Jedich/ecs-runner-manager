package model

type Webhook struct {
	Action     string `json:"action"`
	Workflow   string `json:"workflow"`
	Job        string `json:"job"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}
