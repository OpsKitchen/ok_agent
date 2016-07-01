package api

type RequestParam struct {
	InstanceId       string `json:"instanceId"`
	ServerUniqueName string `json:"serverUniqueName"`
}

func (param *RequestParam) SetServerUniqueName(ServerUniqueName string) *RequestParam {
	param.ServerUniqueName = ServerUniqueName
	return param
}

func (param *RequestParam) SetInstanceId(instanceId string) *RequestParam {
	param.InstanceId = instanceId
	return param
}
