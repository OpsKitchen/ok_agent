package task

import "github.com/OpsKitchen/ok_agent/model/api/returndata"

type Updater struct {
	Api *returndata.DynamicApi
}

func (t *Updater) Run() error {
	return nil
}
