package api

type SysInfoParam struct {
	ServerUniqueName string   `json:"serverUniqueName"`
	Cpu              int      `json:"cpu"`
	Hostname         string   `json:"hostname"`
	Ip               []string `json:"ip"`
	MachineType      string   `json:"machineType"`
	Memory           int      `json:"memory"`
}
