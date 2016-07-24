package main

import (
	"flag"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/Sirupsen/logrus"
	"time"
)

func main() {
	//parse config file from cli argument
	baseConfigFile := flag.String("c", "/etc/ok_agent.json", "base config file path")
	debugAgent := flag.Bool("d", false, "enable debug mode")
	debugApi := flag.Bool("debug-api", false, "enable open api debug log")
	flag.Parse()

	//enable agent debug mode
	if *debugAgent {
		util.Logger.Level = logrus.DebugLevel
	}

	//enable api log
	if *debugApi {
		util.ApiLogger.Level = logrus.DebugLevel
	}

	//prepare config
	baseConf := &config.Base{}
	if err := util.ParseJsonFile(*baseConfigFile, baseConf); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}
	util.Logger.Info("Version: " + baseConf.AgentVersion)

	//prepare credential
	credential := &config.Credential{}
	if err := util.ParseJsonFile(baseConf.CredentialFile, credential); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}

	//prepare api client
	sdk.SetDefaultLogger(util.ApiLogger)
	util.ApiClient.RequestBuilder.Config.SetDisableSSL(baseConf.DisableSSL).SetGatewayHost(baseConf.GatewayHost).
		SetAppMarketIdValue(baseConf.AppMarketId).SetAppVersionValue(baseConf.AgentVersion)
	util.ApiClient.RequestBuilder.Credential.SetAppKey(credential.AppKey).SetSecret(credential.Secret)

	//dispatch
	d := &Dispatcher{
		Config:     baseConf,
		Credential: credential,
	}
	for {
		d.Dispatch()
		time.Sleep(2 * time.Second)
	}
}
