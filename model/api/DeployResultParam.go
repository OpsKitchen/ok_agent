package api
type DeployResultParam struct {
	ServerUniqueName string      `json:"serverUniqueName"`
	Success          bool        `json:"success"`
	ErrorCode        string      `json:"errorCode,omitempty"`
	ErrorMessage     string      `json:"errorMessage,omitempty"`
	Data             interface{} `json:"data,omitempty"`
}
