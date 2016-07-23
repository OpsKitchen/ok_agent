package returndata

type EntranceApi struct {
	WebSocketUrl     string
	DeployApi        *DynamicApi
	ReportResultApi  *DynamicApi
	ReportSysInfoApi *DynamicApi
	UpdateAgentApi   *DynamicApi
}
