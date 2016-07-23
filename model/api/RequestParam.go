package api

type RequestParam struct {
	ServerUniqueName string `json:"serverUniqueName"`
}

func (param *RequestParam) SetServerUniqueName(ServerUniqueName string) *RequestParam {
	param.ServerUniqueName = ServerUniqueName
	return param
}
