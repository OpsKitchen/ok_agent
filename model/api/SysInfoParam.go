package api

type SysInfoParam struct {
	Cpu         int      `json:"cpu"`
	Hostname    string   `json:"hostname"`
	Ip          []string `json:"ip"`
	MachineType string   `json:"machineType"`
	Memory      int      `json:"memory"`
}

func (p *SysInfoParam) Equals(p1 *SysInfoParam) bool {
	if p.Cpu != p1.Cpu {
		return false
	}
	if p.Hostname != p1.Hostname {
		return false
	}
	if p.MachineType != p1.MachineType {
		return false
	}
	if p.Memory != p1.Memory {
		return false
	}
	//check ip list
	ipAmount := len(p.Ip)
	if ipAmount != len(p1.Ip) {
		return false
	}
	for i := 0; i < ipAmount; i++ {
		if p.Ip[i] != p1.Ip[i] {
			return false
		}
	}
	return true
}
