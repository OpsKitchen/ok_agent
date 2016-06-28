package config

type Base struct {
	AgentVersion     string
	CredentialFile   string
	EntryApiName     string
	EntryApiVersion  string
	GatewayHost      string
	DisableSSL       bool
	LogDir           string
	LogLevel         string
}
