package api

type ReportResultParam struct {
	InstanceId       string `json:"instanceId"`
	ServerUniqueName string `json:"serverUniqueName"`
	Success          bool   `json:"success"`
	ErrorMessage     string `json:"errorMessage"`
}

func (param *ReportResultParam) SetServerUniqueName(ServerUniqueName string) *ReportResultParam {
	param.ServerUniqueName = ServerUniqueName
	return param
}

func (param *ReportResultParam) SetInstanceId(instanceId string) *ReportResultParam {
	param.InstanceId = instanceId
	return param
}

func (param *ReportResultParam) SetSuccess(success bool) *ReportResultParam {
	param.Success = success
	return param
}

func (param *ReportResultParam) SetErrorMessage(errorMessage string) *ReportResultParam {
	param.ErrorMessage = errorMessage
	return param
}
