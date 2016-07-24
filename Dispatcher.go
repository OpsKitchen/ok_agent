package main

import (
	"errors"
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

func (d *Dispatcher) Dispatch() error {
	if err := d.callEntranceApi(); err != nil {
		time.Sleep(5 * time.Second)
		return err
	}

	for {
		d.listenWebSocket()
	}
}

func (d *Dispatcher) callEntranceApi() error {
	util.Logger.Debug("Calling entrance api")
	param := &api.EntranceApiParam{ServerUniqueName: d.Credential.ServerUniqueName}
	apiResult, err := util.ApiClient.CallApi(d.Config.EntranceApiName,
		d.Config.EntranceApiVersion, param, &d.EntranceApiResult)
	if err != nil {
		util.Logger.Error("Failed to call entrance api.")
		return err
	}
	if apiResult.Success == false {
		errMsg := "Entrance api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	util.Logger.Info("Succeed to call entrance api.")
	return nil
}

func (d *Dispatcher) listenWebSocket() {
	conn, _, err := websocket.DefaultDialer.Dial(d.EntranceApiResult.WebSocketUrl, nil)
	if err != nil {
		util.Logger.Error("Failed to connect to web socket server: " + err.Error())
		time.Sleep(5 * time.Second)
		return
	}
	util.Logger.Info("Web socket connected, waiting for task...")
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				util.Logger.Error("Server exit abnormally: " + err.Error())
				time.Sleep(5 * time.Second)
			} else {
				//server sent 1000 error code (websocket.CloseNormalClosure)
				util.Logger.Debug("Server sends me close frame: " + err.Error())
				time.Sleep(10 * time.Second)
			}
			continue
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
