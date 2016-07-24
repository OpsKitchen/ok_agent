package main

import (
	"github.com/OpsKitchen/ok_agent/model/api"
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
	Credential        *config.Credential
	EntranceApiResult returndata.EntranceApi
}

func (d *Dispatcher) Dispatch() {
	util.Logger.Info("Calling entrance api")
	param := &api.EntranceApiParam{ServerUniqueName: d.Credential.ServerUniqueName}
	result, err := util.ApiClient.CallApi(d.Config.EntranceApiName,
		d.Config.EntranceApiVersion, param)
	if err != nil {
		util.Logger.Error("Failed to call entrance api.")
		return
	}
	if result.Success == false {
		errMsg := "Entrance api return error: " + result.ErrorCode + "\t" + result.ErrorMessage
		util.Logger.Error(errMsg)
		return
	}
	if result.Data == nil {
		util.Logger.Error("Entrance api return empty data.")
		return
	}
	result.ConvertDataTo(&d.EntranceApiResult)
	util.Logger.Info("Succeed to call entrance api.")
	d.listenWebSocket()
}

func (d *Dispatcher) listenWebSocket() {
	conn, _, err := websocket.DefaultDialer.Dial(d.EntranceApiResult.WebSocketUrl, nil)
	if err != nil {
		util.Logger.Error("Failed to connect to web socket server: " + err.Error())
		time.Sleep(5 * time.Second)
		return
	}
	util.Logger.Info("Web socket server connected, waiting for task...")
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				util.Logger.Debug("Server exit abnormally: " + err.Error())
				time.Sleep(time.Second)
			} else {
				//server sent 1000 error code (websocket.CloseNormalClosure)
				util.Logger.Debug("Server sends me close frame: " + err.Error())
				time.Sleep(10 * time.Second)
			}
			return
		}

		msg := string(message)
		var taskErr error
		switch msg {
		case task.FlagDeploy:
			util.Logger.Info("Received deploy task.")
			deployer := &task.Deployer{Api: d.EntranceApiResult.DeployApi}
			taskErr = deployer.Run()

		case task.FlagReportSysInfo:
			util.Logger.Info("Received sys info report task.")
			reporter := &task.SysInfoReporter{Api: d.EntranceApiResult.ReportSysInfoApi}
			taskErr = reporter.Run()

		case task.FlagUpdateAgent:
			util.Logger.Info("Received agent update task.")
			updater := &task.Updater{Api: d.EntranceApiResult.UpdateAgentApi}
			taskErr = updater.Run()

		default:
			util.Logger.Error("Unsupported task: " + msg)
		}

		d.reportResult(taskErr)
	}
}

func (d *Dispatcher) reportResult(err error) {
	param := &model.ApiResult{}
	if err != nil {
		param.ErrorMessage = "" //read error log from file
	} else {
		param.Success = true
	}
	result, err := util.ApiClient.CallApi(d.EntranceApiResult.ReportResultApi.Name,
		d.EntranceApiResult.ReportResultApi.Version, param)
	if err != nil {
		util.Logger.Error("Failed to call result report api: ", d.EntranceApiResult.ReportResultApi.Name,
			d.EntranceApiResult.ReportResultApi.Version)
		return
	}
	if result.Success == false {
		util.Logger.Error("Result report api return error: " + result.ErrorCode + "\t" + result.ErrorMessage)
	}
}
