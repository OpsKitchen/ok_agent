package api

type EntranceApiParam struct {
	ServerUniqueName string `json:"serverUniqueName"`
}

func (param *EntranceApiParam) SetServerUniqueName(ServerUniqueName string) *EntranceApiParam {
	param.ServerUniqueName = ServerUniqueName
	return param
}
