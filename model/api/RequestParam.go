package api

type RequestParam struct {
	InstanceId       string `json:"instanceId"`
	LastDeployTime   string `json:"lastDeployTime"`
	ServerUniqueName string `json:"serverUniqueName"`
}