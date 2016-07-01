package main

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_agent/adapter"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"io/ioutil"
	"reflect"
	"os"
)

type Dispatcher struct {
	ApiClient      *sdk.Client
	ApiParam       *api.RequestParam
	BaseConfigFile string
	Config         *config.Base
	Credential     *config.Credential
	DynamicApiList []returndata.DynamicApi
}

func (dispatcher *Dispatcher) Dispatch() {
	dispatcher.parseBaseConfig()
	dispatcher.parseCredentialConfig()
	dispatcher.prepareApiClient()
	dispatcher.prepareApiParam()
	dispatcher.prepareDynamicApiList()
	dispatcher.processDynamicApi()
}

func (dispatcher *Dispatcher) parseBaseConfig() {
	var baseConfig *config.Base
	var err error
	var jsonBytes []byte

	if util.FileExist(dispatcher.BaseConfigFile) == false {
		util.Logger.Fatal("Base config file not found: ", dispatcher.BaseConfigFile)
	}

	jsonBytes, err = ioutil.ReadFile(dispatcher.BaseConfigFile)
	if err != nil {
		util.Logger.Fatal("Base config file not readable: ", dispatcher.BaseConfigFile)
	}

	err = json.Unmarshal(jsonBytes, &baseConfig)
	if err != nil {
		util.Logger.Fatal("Base config file parse failed: ", err.Error())
	}

	dispatcher.Config = baseConfig
}

func (dispatcher *Dispatcher) parseCredentialConfig() {
	var credentialConfig *config.Credential
	var err error
	var jsonBytes []byte

	if util.FileExist(dispatcher.Config.CredentialFile) == false {
		util.Logger.Fatal("Credential config file not found: ", dispatcher.Config.CredentialFile)
	}

	jsonBytes, err = ioutil.ReadFile(dispatcher.Config.CredentialFile)
	if err != nil {
		util.Logger.Fatal("Credential config file not readable: ", dispatcher.Config.CredentialFile)
	}

	err = json.Unmarshal(jsonBytes, &credentialConfig)
	if err != nil {
		util.Logger.Fatal("Credential config file parse failed: ", err.Error())
	}

	dispatcher.Credential = credentialConfig
}

func (dispatcher *Dispatcher) prepareApiClient() {
	var client *sdk.Client = sdk.NewClient()
	//inject logger
	sdk.SetDefaultLogger(util.Logger)

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

func (dispatcher *Dispatcher) prepareDynamicApiList() {
	var apiResult *model.ApiResult
	var err error

	apiResult, err = dispatcher.ApiClient.CallApi(dispatcher.Config.EntranceApiName,
		dispatcher.Config.EntranceApiVersion, dispatcher.ApiParam, &dispatcher.DynamicApiList)

	if err != nil {
		util.Logger.Fatal("Call entrance api failed", err.Error())
	}
	if apiResult.Success == false {
		util.Logger.Fatal("Entrance api return error: ", apiResult.ErrorCode, apiResult.ErrorMessage)
	}
	if len(dispatcher.DynamicApiList) == 0 {
		util.Logger.Fatal("Entrance api return empty api list")
	}
}

func (dispatcher *Dispatcher) processDynamicApi() {
	var dynamicApi returndata.DynamicApi
	var errorCount int
	for _, dynamicApi = range dispatcher.DynamicApiList {
		util.Logger.Debug("Calling dynamic api: ", dynamicApi.Name)
		var apiResult *model.ApiResult
		var apiResultDataKind reflect.Kind
		var err error

		//call dynamic api
		apiResult, err = dispatcher.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, dispatcher.ApiParam, nil)
		if err != nil {
			util.Logger.Fatal("Call api failed: ", dynamicApi.Name, dynamicApi.Version)
		}
		if apiResult.Success == false {
			util.Logger.Fatal("Api return error: ", apiResult.ErrorCode, apiResult.ErrorMessage)
		}

		//server return empty data, nothing to do, go to the next api
		if apiResult.Data == nil {
			continue
		}

		//cast item list to native go type
		apiResultDataKind = reflect.TypeOf(apiResult.Data).Kind()
		if apiResultDataKind != reflect.Slice {
			util.Logger.Fatal("Wrong return data type, expected list, got: ", reflect.TypeOf(apiResult.Data))
		}

		switch dynamicApi.ReturnDataType {
		case returndata.AugeasList:
			continue

		case returndata.CommandList:
			var item adapter.Command
			var itemList []adapter.Command = []adapter.Command{}
			err = util.JsonConvert(apiResult.Data, &itemList)
			for _, item = range itemList {
				err = item.Process()
				if err != nil {
					errorCount ++
					if DebugMode == true {
						os.Exit(1)
					}
				}
			}

		case returndata.FileList:
			var item adapter.File
			var itemList []adapter.File = []adapter.File{}
			err = util.JsonConvert(apiResult.Data, &itemList)
			for _, item = range itemList {
				err = item.Process()
				if err != nil {
					errorCount ++
					if DebugMode == true {
						os.Exit(1)
					}
				}
			}
		default:
			util.Logger.Fatal("Unsupported list: ", dynamicApi.ReturnDataType)
		}//end switch
	}//end for

	if errorCount > 0 {
		util.Logger.Fatal(errorCount, " error(s) occourred, run me with '-d' option to see more detail")
	} else {
		util.Logger.Info("Congratulations! All tasks have been done successfully!")
	}
}
