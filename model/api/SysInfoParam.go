package api

type SysInfoParam struct {
	MachineType string `json:"machineType"`
	Hostname    string `json:"hostname"`
	Cpu         int    `json:"cpu"`
	Memory      int    `json:"memory"`
	Ip          string `json:"ip"`
}
