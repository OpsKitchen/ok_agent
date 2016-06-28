package config

type Base struct {
	AgentVersion       string
	CredentialFile     string
	EntranceApiName    string
	EntranceApiVersion string
	GatewayHost        string
	DisableSSL         bool
	LogDir             string
	LogLevel           string
}
