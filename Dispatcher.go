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

const (
	// Time allowed to write a message to the peer.
	writeWait = 3 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 100 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Dispatcher struct {
	EntranceApiResult returndata.EntranceApi
	wsConn            *websocket.Conn
	wsBroken          chan bool
}

func (d *Dispatcher) Dispatch() {
	util.Logger.Debug("Calling entrance api")
	param := &api.EntranceApiParam{ServerUniqueName: config.C.ServerUniqueName}
	result, err := util.ApiClient.CallApi(config.B.EntranceApiName, config.B.EntranceApiVersion, param)
	if err != nil {
		util.Logger.Error("Failed to call entrance api: " + err.Error())
		return
	}
	if result.Success == false {
		util.Logger.Error("Entrance api return error: " + result.ErrorCode + ": " + result.ErrorMessage)
		return
	}
	if result.Data == nil {
		util.Logger.Error("Entrance api return empty data.")
		return
	}
	result.ConvertDataTo(&d.EntranceApiResult)
	util.Logger.Info("Succeed to call entrance api.")
	go d.reportSysInfo()
	go d.updateAgent()
	d.listenWebSocket()
}

func (d *Dispatcher) listenWebSocket() {
	conn, resp, err := websocket.DefaultDialer.Dial(d.EntranceApiResult.WebSocketUrl, nil)
	if err != nil {
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			bytes, _ := ioutil.ReadAll(resp.Body)
			util.Logger.Error("Failed to handshake with web socket server: " + resp.Status + ": " + string(bytes))
		} else {
			util.Logger.Error("Failed to dial to web socket server: " + err.Error())
		}
		return
	}
	defer conn.Close()
	util.Logger.Info("Web socket server connected, waiting for task...")

	d.wsConn = conn
	d.wsBroken = make(chan bool, 1)
	go d.PingWsServer()
	d.ReadWsMessage()
}

func (d *Dispatcher) ReadWsMessage() {
	defer func() {
		d.wsBroken <- true
	}()
	for {
		mt, message, err := d.wsConn.ReadMessage()
		if err != nil {
			util.Logger.Error("Can not read message: "+err.Error()+"\t message type: ", mt)
			return
		}
		d.execTask(string(message))
	}
}

func (d *Dispatcher) PingWsServer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		d.wsBroken <- true
	}()
	for {
		select {
		case <-d.wsBroken:
			return
		case <-ticker.C:
			if err := d.wsConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				util.Logger.Error("Ping failed of ws connection: " + err.Error())
				return
			}
		}
	}
}

func (d *Dispatcher) execTask(msg string) {
	switch msg {
	case task.FlagDeploy:
		util.Logger.Debug("Received deploy task.")
		deployer := &task.Deployer{
			DeployApi:       d.EntranceApiResult.DeployApi,
			ReportResultApi: d.EntranceApiResult.ReportResultApi,
		}
		deployer.Run()

	case task.FlagReportSysInfo:
		util.Logger.Debug("Received report sys info task.")
		d.reportSysInfo()

	case task.FlagUpdateAgent:
		util.Logger.Debug("Received update agent task.")
		d.updateAgent()

	default:
		errMsg := "Unsupported task: " + msg
		util.Logger.Error(errMsg)
	}
}

func (d *Dispatcher) reportSysInfo() {
	reporter := &task.SysInfoReporter{Api: d.EntranceApiResult.ReportSysInfoApi}
	reporter.Run()
}

func (d *Dispatcher) updateAgent() {
	updater := &task.Updater{Api: d.EntranceApiResult.UpdateAgentApi}
	updater.Run()
}
