package main

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_agent/adapter"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_agent/wsclient"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"os"
	"os/signal"
	"time"
	"unsafe"
)

type Dispatcher struct {
	ApiClient      *sdk.Client
	ApiParam       *api.RequestParam
	BaseConfigFile string
	Config         *config.Base
	Credential     *config.Credential
	EntranceApi    returndata.EntranceApi
	DeployApi      returndata.DeployApi
}

func (dispatcher *Dispatcher) Dispatch() {
	dispatcher.parseBaseConfig()
	dispatcher.parseCredentialConfig()
	dispatcher.prepareApiClient()
	dispatcher.prepareApiParam()
	dispatcher.processEntranceApi()
	dispatcher.processWebSocket()
}

func (dispatcher *Dispatcher) parseBaseConfig() {
	if util.FileExist(dispatcher.BaseConfigFile) == false {
		util.Logger.Fatal("Base config file not found: ", dispatcher.BaseConfigFile)
	}
	jsonBytes, err := ioutil.ReadFile(dispatcher.BaseConfigFile)
	if err != nil {
		util.Logger.Fatal("Base config file not readable: ", dispatcher.BaseConfigFile)
	}
	if err := json.Unmarshal(jsonBytes, &dispatcher.Config); err != nil {
		util.Logger.Fatal("Base config file is not a valid json file: " + dispatcher.BaseConfigFile)
	}
	util.Logger.Info("Runing opskitchen agent " + dispatcher.Config.AgentVersion)
}

func (dispatcher *Dispatcher) parseCredentialConfig() {
	if util.FileExist(dispatcher.Config.CredentialFile) == false {
		util.Logger.Fatal("Credential config file not found: ", dispatcher.Config.CredentialFile)
	}
	jsonBytes, err := ioutil.ReadFile(dispatcher.Config.CredentialFile)
	if err != nil {
		util.Logger.Fatal("Credential config file not readable: ", dispatcher.Config.CredentialFile)
	}
	if err := json.Unmarshal(jsonBytes, &dispatcher.Credential); err != nil {
		util.Logger.Fatal("Credential config file is not a valid json file: " + dispatcher.Config.CredentialFile)
	}
}

func (dispatcher *Dispatcher) prepareApiClient() {
	client := sdk.NewClient()
	//inject logger
	sdk.SetDefaultLogger(util.ApiLogger)

	//init config
	client.RequestBuilder.Config.SetAppMarketIdValue(dispatcher.Config.AppMarketId).SetAppVersionValue(
		dispatcher.Config.AgentVersion).SetGatewayHost(
		dispatcher.Config.GatewayHost).SetDisableSSL(dispatcher.Config.DisableSSL)

	//init credential
	client.RequestBuilder.Credential.SetAppKey(dispatcher.Credential.AppKey).SetSecret(
		dispatcher.Credential.Secret)
	dispatcher.ApiClient = client
}

func (dispatcher *Dispatcher) prepareApiParam() {
	dispatcher.ApiParam = &api.RequestParam{}
	dispatcher.ApiParam.SetServerUniqueName(dispatcher.Credential.ServerUniqueName).SetInstanceId(
		dispatcher.Credential.InstanceId)
}

func (dispatcher *Dispatcher) processEntranceApi() {
	util.Logger.Debug("Calling entrance api")

	apiResult, err := dispatcher.ApiClient.CallApi(dispatcher.Config.EntranceApiName,
		dispatcher.Config.EntranceApiVersion, dispatcher.ApiParam, &dispatcher.EntranceApi)

	if err != nil {
		util.Logger.Debug("Failed to call entrance api.")
		return
	}
	if apiResult.Success == false {
		util.Logger.Debug("Entrance api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage)
		return
	}
	util.Logger.Info("Succeed to call entrance api.")

	if dispatcher.EntranceApi.ReportInstance {
		util.Logger.Debug("Calling report instance api")
		var reportInstanceParam = api.ReportInstanceParam{}
		var returnDataPointer interface{}
		reportInstanceParam.SetServerUniqueName(dispatcher.Credential.ServerUniqueName).SetInstanceId(
			dispatcher.Credential.InstanceId).SetMachineType(string("virtual")).SetCpu(
			int(1)).SetMemory(int(1024))

		reportResult, err := dispatcher.ApiClient.CallApi(dispatcher.EntranceApi.ReportInstanceApiParams.Name,
			dispatcher.EntranceApi.ReportInstanceApiParams.Version, reportInstanceParam, returnDataPointer)
		if err != nil {
			util.Logger.Debug("Failed to call report instance api: ", dispatcher.EntranceApi.ReportInstanceApiParams.Name,
				dispatcher.EntranceApi.ReportInstanceApiParams.Version)
			return
		}
		if reportResult.Success == false {
			util.Logger.Debug("Api return error: ", reportResult.ErrorCode, reportResult.ErrorMessage)
			return
		}
		util.Logger.Info("Succeed to call report instance api.")
	}
}

func (dispatcher *Dispatcher) processWebSocket() {
	var conn *websocket.Conn
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn = dispatcher.connWebSocket()
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer conn.Close()
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				util.Logger.Error("read:", err)
				conn = dispatcher.connWebSocket()
				continue
			}

			msg := *(*string)(unsafe.Pointer(&message))
			util.Logger.Debug("recv:", msg)

			switch msg {
			case wsclient.TaskFlagDeploy:
				dispatcher.prepareDeploy()

			case wsclient.TaskFlagUpgrade:
				//update agent version

			default:
				util.Logger.Error("Unsupported msg: ", msg)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-interrupt:
			util.Logger.Debug("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				util.Logger.Error("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			conn.Close()
			return
		}
	}
}

func (dispatcher *Dispatcher) connWebSocket() *websocket.Conn {
	for {
		c, _, err := websocket.DefaultDialer.Dial(dispatcher.EntranceApi.WebSocketUrl, nil)
		if err != nil {
			util.Logger.Error("dial:", err)
			time.Sleep(5 * time.Second)
			dispatcher.processEntranceApi()
			continue
		}
		return c
	}
}

func (dispatcher *Dispatcher) prepareDeploy() {
	var reportResultParam = api.ReportResultParam{}
	var returnDataPointer interface{}
	reportResultParam.SetServerUniqueName(dispatcher.Credential.ServerUniqueName).SetInstanceId(
		dispatcher.Credential.InstanceId)
	util.Logger.Debug("Calling deploy server")
	apiListResult := dispatcher.prepareDynamicApiList()
	if apiListResult.Success {
		for _, dynamicApi := range dispatcher.DeployApi.ApiList {
			apiResult := dispatcher.processDynamicApi(dynamicApi)
			if !apiResult.Success {
				util.Logger.Debug(apiResult.ErrorMessage)
				util.Logger.Debug("Calling report result api")
				reportResultParam.SetSuccess(apiResult.Success).SetErrorMessage(apiResult.ErrorMessage)

				reportResult, err := dispatcher.ApiClient.CallApi(dispatcher.EntranceApi.ReportResultApiParams.Name,
					dispatcher.EntranceApi.ReportResultApiParams.Version, reportResultParam, returnDataPointer)
				if err != nil {
					util.Logger.Debug("Failed to call report result api: ", dispatcher.EntranceApi.ReportResultApiParams.Name,
						dispatcher.EntranceApi.ReportResultApiParams.Version)
				}
				if reportResult.Success == false {
					util.Logger.Debug("Api return error: ", reportResult.ErrorCode, reportResult.ErrorMessage)
				}
				return
			}
		}
		util.Logger.Info("Congratulations! All tasks have been done successfully!")
	} else {
		util.Logger.Debug(apiListResult.ErrorMessage)
	}
	util.Logger.Debug("Calling deploy server")
	reportResultParam.SetSuccess(apiListResult.Success).SetErrorMessage(apiListResult.ErrorMessage)

	reportResult, err := dispatcher.ApiClient.CallApi(dispatcher.EntranceApi.ReportResultApiParams.Name,
		dispatcher.EntranceApi.ReportResultApiParams.Version, reportResultParam, returnDataPointer)
	if err != nil {
		util.Logger.Debug("Failed to call report result api: ", dispatcher.EntranceApi.ReportResultApiParams.Name,
			dispatcher.EntranceApi.ReportResultApiParams.Version)
	}
	if reportResult.Success == false {
		util.Logger.Debug("Api return error: ", reportResult.ErrorCode, reportResult.ErrorMessage)
	}
}

func (dispatcher *Dispatcher) prepareDynamicApiList() *model.ApiResult {
	util.Logger.Debug("Calling deploy api")
	var apiResult *model.ApiResult

	apiResult, err := dispatcher.ApiClient.CallApi(dispatcher.EntranceApi.DeployApiParams.Name,
		dispatcher.EntranceApi.DeployApiParams.Version, dispatcher.ApiParam, &dispatcher.DeployApi)

	if err != nil {
		errorMessage := "Failed to call deploy api."
		util.Logger.Debug(errorMessage)
		apiResult.Success = false
		apiResult.ErrorMessage = errorMessage
		return apiResult
	}
	if apiResult.Success == false {
		util.Logger.Debug("Deploy api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage)
		return apiResult
	}
	if len(dispatcher.DeployApi.ApiList) == 0 {
		errorMessage := "Deploy api return empty api list."
		util.Logger.Debug(errorMessage)
		apiResult.Success = false
		apiResult.ErrorMessage = errorMessage
		return apiResult
	}
	util.Logger.Info("Succeed to call deploy api.")
	util.Logger.Info("Product version: " + dispatcher.DeployApi.ProductVersion)
	util.Logger.Info("Server name: " + dispatcher.DeployApi.ServerName)
	apiResult.Data = nil
	return apiResult
}

func (dispatcher *Dispatcher) processDynamicApi(dynamicApi returndata.DynamicApi) *model.ApiResult {
	util.Logger.Debug("Calling dynamic api: ", dynamicApi.Name)
	var mapItemList []map[string]interface{}

	//call dynamic api
	apiResult, err := dispatcher.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, dispatcher.ApiParam, &mapItemList)
	if err != nil {
		util.Logger.Debug("Failed to call api: ", dynamicApi.Name, dynamicApi.Version)
		apiResult.Success = false
		apiResult.ErrorMessage = "Failed to call api: " + dynamicApi.Name + dynamicApi.Version
		return apiResult
	}
	if apiResult.Success == false {
		util.Logger.Debug("Api return error: ", apiResult.ErrorCode, apiResult.ErrorMessage)
		return apiResult
	}
	if apiResult.Data == nil {
		util.Logger.Debug("Api returns empty data, nothing to do, go to next api")
		return apiResult
	}

	//cast item list to native go type
	for _, mapItem := range mapItemList {
		var item adapter.AdapterInterface
		switch dynamicApi.ReturnDataType {
		case returndata.AugeasList:
			item = &adapter.Augeas{}

		case returndata.CommandList:
			item = &adapter.Command{}

		case returndata.FileList:
			item = &adapter.File{}

		default:
			util.Logger.Debug("Unsupported list: ", dynamicApi.ReturnDataType)
			apiResult.Success = false
			apiResult.ErrorMessage = "Unsupported list: " + dynamicApi.ReturnDataType
			return apiResult
		}

		//data type casting with json
		err = util.JsonConvert(mapItem, &item)
		if err != nil {
			errorMessage := "Failed to convert item data type"
			util.Logger.Debug(errorMessage)
			apiResult.Success = false
			apiResult.ErrorMessage = "Unsupported list: " + dynamicApi.ReturnDataType
			return apiResult
		}
		util.Logger.Info("Processing..." + item.Brief())
		if item.Check() == nil && item.Parse() == nil && item.Process() == nil {
			continue
		} else {
			apiResult.Success = false
			apiResult.ErrorMessage = "Failed to adapter exec."
			return apiResult
		}
	} //end for "range apiResultData"
	util.Logger.Info("Succeed to call dynamic api: ", dynamicApi.Name)

	apiResult.Data = nil
	return apiResult
}
