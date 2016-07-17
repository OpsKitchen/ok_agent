package main

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_agent/adapter"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"io/ioutil"
	"os"
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
	dispatcher.prepareEntranceApi()
	dispatcher.prepareDynamicApiList()
	dispatcher.processDynamicApi()
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

func (dispatcher *Dispatcher) prepareEntranceApi() {
	var err error
	util.Logger.Debug("Calling entrance api")

	apiResult, err := dispatcher.ApiClient.CallApi(dispatcher.Config.EntranceApiName,
		dispatcher.Config.EntranceApiVersion, dispatcher.ApiParam, &dispatcher.EntranceApi)

	if err != nil {
		util.Logger.Fatal("Failed to call entrance api.")
	}
	if apiResult.Success == false {
		util.Logger.Fatal("Entrance api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage)
	}
	util.Logger.Info("Succeed to call entrance api.")
}

func (dispatcher *Dispatcher) prepareDynamicApiList() {
	var err error
	util.Logger.Debug("Calling deploy api")

	apiResult, err := dispatcher.ApiClient.CallApi(dispatcher.EntranceApi.DeployApiParams.Name,
		dispatcher.EntranceApi.DeployApiParams.Version, dispatcher.ApiParam, &dispatcher.DeployApi)

	if err != nil {
		util.Logger.Fatal("Failed to call deploy api.")
	}
	if apiResult.Success == false {
		util.Logger.Fatal("Deploy api return error: " + apiResult.ErrorCode + "\t" + apiResult.ErrorMessage)
	}
	if len(dispatcher.DeployApi.ApiList) == 0 {
		util.Logger.Fatal("Deploy api return empty api list")
	}
	util.Logger.Info("Succeed to call deploy api.")
	util.Logger.Info("Product version: " + dispatcher.DeployApi.ProductVersion)
	util.Logger.Info("Server name: " + dispatcher.DeployApi.ServerName)

}

func (dispatcher *Dispatcher) processDynamicApi() {
	errorCount := 0
	for _, dynamicApi := range dispatcher.DeployApi.ApiList {
		util.Logger.Debug("Calling dynamic api: ", dynamicApi.Name)
		var mapItemList []map[string]interface{}

		//call dynamic api
		apiResult, err := dispatcher.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, dispatcher.ApiParam, &mapItemList)
		if err != nil {
			util.Logger.Fatal("Failed to call api: ", dynamicApi.Name, dynamicApi.Version)
		}
		if apiResult.Success == false {
			util.Logger.Fatal("Api return error: ", apiResult.ErrorCode, apiResult.ErrorMessage)
		}
		if apiResult.Data == nil {
			util.Logger.Debug("Api returns empty data, nothing to do, go to next api")
			continue
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
				util.Logger.Fatal("Unsupported list: ", dynamicApi.ReturnDataType)
			}

			//data type casting with json
			err = util.JsonConvert(mapItem, &item)
			if err != nil {
				util.Logger.Fatal("Failed to convert item data type")
			}
			util.Logger.Info("Processing..." + item.Brief())
			if item.Check() == nil && item.Parse() == nil && item.Process() == nil {
				continue
			}

			errorCount++
			if DebugAgent == true {
				os.Exit(1)
			}
		} //end for "range apiResultData"
	} //end for "range dispatcher.DynamicApiList"

	if errorCount > 0 {
		util.Logger.Fatal(errorCount, " error(s) occourred, run me with '-d' option to see more detail")
	} else {
		util.Logger.Info("Congratulations! All tasks have been done successfully!")
	}
}
