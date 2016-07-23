package main

import (
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/task"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"github.com/gorilla/websocket"
	"time"
)

type Dispatcher struct {
	Config            *config.Base
	EntranceApiResult returndata.EntranceApi
}

func (d *Dispatcher) Dispatch() error {
	if err := d.callEntranceApi(); err != nil {
		return err
	}

	for {
		d.listenWebSocket()
	}
}

func (d *Dispatcher) callEntranceApi() error {
	util.Logger.Debug("Calling entrance api")

	apiResult, err := util.ApiClient.CallApi(d.Config.EntranceApiName,
		d.Config.EntranceApiVersion, util.ApiParam, &d.EntranceApiResult)

	if err != nil {
		util.Logger.Error("Failed to call entrance api.")
		return err
	}
	if apiResult.Success == false {
		util.Logger.Error("Entrance api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage)
		return err
	}
	util.Logger.Info("Succeed to call entrance api.")
	return nil
}

func (d *Dispatcher) listenWebSocket() {
	conn, _, err := websocket.DefaultDialer.Dial(d.EntranceApiResult.WebSocketUrl, nil)
	defer conn.Close()
	if err != nil {
		util.Logger.Error("dial:", err)
		time.Sleep(5 * time.Second)
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			util.Logger.Error("read:", err)
			continue
		}

		msg := string(message)
		util.Logger.Debug("recv:", msg)

		var taskErr error
		switch msg {
		case task.FlagDeploy:
			deployer := &task.Deployer{Api: d.EntranceApiResult.DeployApi}
			taskErr = deployer.Run()

		case task.FlagReportSysInfo:
			reporter := &task.SysInfoReporter{Api: d.EntranceApiResult.ReportSysInfoApi}
			taskErr = reporter.Run()

		case task.FlagUpdateAgent:
			updater := &task.Updater{Api: d.EntranceApiResult.UpdateAgentApi}
			taskErr = updater.Run()

		default:
			util.Logger.Error("Unsupported task: ", msg)
		}

		d.reportResult(taskErr)
	}
}

func (d *Dispatcher) reportResult(err error) {
	reportResultParam := &model.ApiResult{}
	if err != nil {
		reportResultParam.ErrorMessage = "" //read error log from file
	} else {
		reportResultParam.Success = true
	}
	reportResult, err := util.ApiClient.CallApi(d.EntranceApiResult.ReportResultApi.Name,
		d.EntranceApiResult.ReportResultApi.Version, reportResultParam, nil)
	if err != nil {
		util.Logger.Error("Failed to call result report api: ", d.EntranceApiResult.ReportResultApi.Name,
			d.EntranceApiResult.ReportResultApi.Version)
		return
	}
	if reportResult.Success == false {
		util.Logger.Error("Result report api return error: ", reportResult.ErrorCode, reportResult.ErrorMessage)
	}
}
