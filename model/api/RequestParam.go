package api

type RequestParam struct {
	instanceId string
	lastDeployTime string
	serverUniqueName string
}

func (requestParam *RequestParam) Map() map[string]interface{} {
	var paramMap = make(map[string]interface{})
	paramMap["instanceId"] = requestParam.instanceId
	paramMap["lastDeployTime"] = requestParam.lastDeployTime
	paramMap["serverUniqueName"] = requestParam.serverUniqueName

	return paramMap
}

func (requestParam *RequestParam) SetInstanceId(instanceId string) *RequestParam {
	requestParam.instanceId = instanceId
	return requestParam
}

func (requestParam *RequestParam) SetLastDeployTime(lastDeployTime string) *RequestParam {
	requestParam.lastDeployTime = lastDeployTime
	return requestParam
}

func (requestParam *RequestParam) SetServerUniqueName(serverUniqueName string) *RequestParam {
	requestParam.serverUniqueName = serverUniqueName
	return requestParam
}