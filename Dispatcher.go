package main

import (
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/task"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/gorilla/websocket"
	"io/ioutil"
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
		errMsg := "Entrance api return error: " + result.ErrorCode + ": " + result.ErrorMessage
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
	conn, resp, err := websocket.DefaultDialer.Dial(d.EntranceApiResult.WebSocketUrl, nil)
	defer resp.Body.Close()
	if err != nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		util.Logger.Error("Failed to connect to web socket server: " + resp.Status + ": " + string(bytes))
		return
	}
	util.Logger.Info("Web socket server connected, waiting for task...")
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				util.Logger.Debug("Connection breaks abnormally: " + err.Error())
			} else {
				//server sent 1000 error code (websocket.CloseNormalClosure)
				util.Logger.Debug("Server sends me close frame: " + err.Error())
				time.Sleep(10 * time.Second)
			}
			return
		}

		d.execTask(string(message))
	}
}

func (d *Dispatcher) execTask(msg string) {
	switch msg {
	case task.FlagDeploy:
		util.Logger.Info("Received deploy task.")
		deployer := &task.Deployer{Api: d.EntranceApiResult.DeployApi}
		deployer.Run()

	case task.FlagReportSysInfo:
		util.Logger.Info("Received report sys info task.")
		reporter := &task.SysInfoReporter{Api: d.EntranceApiResult.ReportSysInfoApi}
		reporter.Run()

	case task.FlagUpdateAgent:
		util.Logger.Info("Received update agent task.")
		updater := &task.Updater{Api: d.EntranceApiResult.UpdateAgentApi}
		updater.Run()

	default:
		errMsg := "Unsupported task: " + msg
		util.Logger.Error(errMsg)
	}
}
