package config

type Base struct {
	AgentVersion   string
	CredentialFile string
	EntryApiName   string
	GatewayHost    string
	DisableSSL     bool
	LogDir         string
	LogLevel       string
}
