package api

type ReportInstanceParam struct {
	InstanceId       string `json:"instanceId"`
	ServerUniqueName string `json:"serverUniqueName"`
	MachineType      string `json:"machineType"`
	Cpu              int    `json:"cpu"`
	Memory           int    `json:"memory"`
}

func (param *ReportInstanceParam) SetServerUniqueName(ServerUniqueName string) *ReportInstanceParam {
	param.ServerUniqueName = ServerUniqueName
	return param
}

func (param *ReportInstanceParam) SetInstanceId(instanceId string) *ReportInstanceParam {
	param.InstanceId = instanceId
	return param
}

func (param *ReportInstanceParam) SetMachineType(machineType string) *ReportInstanceParam {
	param.MachineType = machineType
	return param
}

func (param *ReportInstanceParam) SetCpu(cpu int) *ReportInstanceParam {
	param.Cpu = cpu
	return param
}

func (param *ReportInstanceParam) SetMemory(memory int) *ReportInstanceParam {
	param.Memory = memory
	return param
}
