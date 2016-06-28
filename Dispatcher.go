package main

import (
	//go builtin pkg
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	//local pkg
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
)

type Dispatcher struct {
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (dispatcher *Dispatcher) Dispatch(baseConfigFile string) {
	var baseConfig *config.Base = dispatcher.parseBaseConfig(baseConfigFile)
	var credentialConfig *config.Credential = dispatcher.parseCredentialConfig(baseConfig.CredentialFile)
	var client *sdk.Client = dispatcher.prepareApiClient(baseConfig, credentialConfig)
	var resp *model.ApiResult
	var err error

	sdk.SetDefaultLogger(debugLogger)

	var params = &api.RequestParam{}
	params.ServerUniqueName = credentialConfig.ServerUniqueName
	resp, err = client.CallApi(baseConfig.EntryApiName, baseConfig.EntryApiVersion, params)
	if err != nil {
	} else {
		fmt.Println(resp)
	}

}

func (dispatcher *Dispatcher) parseBaseConfig(baseConfigFile string) *config.Base {
	var baseConfig *config.Base
	var err error
	var jsonBytes []byte

	debugLogger.Info("base config file: ", baseConfigFile)
	if _, err := os.Stat(baseConfigFile); os.IsNotExist(err) {
		debugLogger.Fatal("base config file not found")
	}

	jsonBytes, err = ioutil.ReadFile(baseConfigFile)
	if err != nil {
		debugLogger.Fatal("base config file not readable")
	}

	err = json.Unmarshal(jsonBytes, &baseConfig)
	if err != nil {
		debugLogger.Fatal("json decode failed: ", err.Error())
	}

	return baseConfig
}

func (dispatcher *Dispatcher) parseCredentialConfig(credentialConfigFile string) *config.Credential {
	var credentialConfig *config.Credential
	var err error
	var jsonBytes []byte

	debugLogger.Info("credential config file: ", credentialConfigFile)
	if _, err := os.Stat(credentialConfigFile); os.IsNotExist(err) {
		debugLogger.Fatal("credential config file not found")
	}

	jsonBytes, err = ioutil.ReadFile(credentialConfigFile)
	if err != nil {
		debugLogger.Fatal("credential config file not readable")
	}

	err = json.Unmarshal(jsonBytes, &credentialConfig)
	if err != nil {
		debugLogger.Fatal("json decode failed: ", err.Error())
	}

	return credentialConfig
}

func (dispatcher *Dispatcher) prepareApiClient(base *config.Base, credential *config.Credential) *sdk.Client {
	var client *sdk.Client = sdk.NewClient()
	//init config
	client.RequestBuilder.Config.SetAppMarketIdValue("1").SetAppVersionValue(base.AgentVersion).SetGatewayHost(
		base.GatewayHost).SetDisableSSL(base.DisableSSL)

	//init credential
	client.RequestBuilder.Credential.SetAppKey(credential.AppKey).SetSecret(credential.Secret)
	return client
}