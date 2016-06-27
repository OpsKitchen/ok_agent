package main

import (
	//go builtin pkg
	"encoding/json"
	"io/ioutil"
	"os"

	//local pkg
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
)

type Dispatcher struct {
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (dispatcher *Dispatcher) Dispatch(baseConfigFile string) {
	var baseConfig *config.Base = dispatcher.ParseBaseConfig(baseConfigFile)
	var credentialConfig *config.Credential = dispatcher.ParseCredentialConfig(baseConfig.CredentialFile)
	dispatcher.PrepareApiClient(baseConfig, credentialConfig)
}

func (dispatcher *Dispatcher) ParseBaseConfig(baseConfigFile string) *config.Base {
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

func (dispatcher *Dispatcher) ParseCredentialConfig(credentialConfigFile string) *config.Credential {
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

func (dispatcher *Dispatcher) PrepareApiClient(base *config.Base, credential *config.Credential) *sdk.Client {
	var client *sdk.Client = sdk.NewClient()
	//init config
	client.RequestBuilder.Config.SetAppMarketIdValue("1").SetAppVersionValue(base.AgentVersion).SetGatewayHost(
		base.GatewayHost).SetDisableSSL(base.DisableSSL)

	//init credential
	client.RequestBuilder.Credential.SetAppKey(credential.AppKey).SetSecret(credential.Secret)
	return client
}
