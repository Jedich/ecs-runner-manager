package model

type ControllerRequest struct {
	Runners []*RequestRunner `json:"runners"`
}

type RequestRunner struct {
	Name        string       `json:"name"`
	PrivateIPv4 string       `json:"private_ip"`
	Status      RunnerStatus `json:"status"`
	Metrics     []Metrics    `json:"metrics"`
}

type AuthResponse struct {
	Data AuthResponseData `json:"data"`
}

type AuthResponseData struct {
	AccessToken string `json:"access_token"`
}
