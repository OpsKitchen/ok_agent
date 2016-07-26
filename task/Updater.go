package task

import (
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
)

type Updater struct {
	Api *returndata.DynamicApi
}

func (t *Updater) Run() error {
	deployer := &Deployer{}
	return deployer.processDynamicApi(*t.Api)
}
