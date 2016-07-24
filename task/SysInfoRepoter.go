package task

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/util"
)

type SysInfoReporter struct {
	Api    *returndata.DynamicApi
	Params *api.SysInfoParam
}

func (t *SysInfoReporter) Run() error {
	util.Logger.Debug("Calling report instance api")
	if err := t.setCpu(); err != nil {
		return err
	}
	if err := t.setHostname(); err != nil {
		return err
	}
	if err := t.setIp(); err != nil {
		return err
	}
	if err := t.setMachineType(); err != nil {
		return err
	}
	if err := t.setMemory(); err != nil {
		return err
	}

	reportResult, err := util.ApiClient.CallApi(t.Api.Name, t.Api.Version, t.Params, nil)
	if err != nil {
		util.Logger.Error("Failed to call sys info report api: " + t.Api.Name + "\t" + t.Api.Version)
		return err
	}
	if reportResult.Success == false {
		errMsg := "api return error: " + reportResult.ErrorCode + "\t" + reportResult.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	util.Logger.Debug("Succeed to call sys info report api.")
	return nil
}

func (t *SysInfoReporter) setCpu() error {
	t.Params.Cpu = 1
	return nil
}

func (t *SysInfoReporter) setHostname() error {
	t.Params.Hostname = "localhost"
	return nil
}

func (t *SysInfoReporter) setIp() error {
	t.Params.Hostname = "172.16.0.1,192.168.0.1"
	return nil
}

func (t *SysInfoReporter) setMachineType() error {
	t.Params.MachineType = "virtual"
	return nil
}

func (t *SysInfoReporter) setMemory() error {
	t.Params.Memory = 1024
	return nil
}
