package main

import (
	"flag"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/Sirupsen/logrus"
	"os"
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
	if err := config.ParseBaseConfig(*baseConfigFile); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}

	//prepare credential
	if err := config.ParseCredential(); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}

	//check log dir
	if _, err := os.Stat(config.B.LogDir); err != nil { //log dir not exists
		if os.MkdirAll(config.B.LogDir, 0755) != nil {
			util.Logger.Fatal("Failed to create log dir [" + config.B.LogDir + "]: " + err.Error())
		}
	}
	logFileName := config.B.LogDir + "/info.log"
	var fileHandle *os.File
	if _, err := os.Stat(logFileName); err != nil { //log file not exists
		if fileHandle, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			util.Logger.Fatal("Failed to create log file [" + logFileName + "]: " + err.Error())
		}
	} else if fileHandle, err = os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		util.Logger.Fatal("Failed to open log file [" + logFileName + "] for writing: " + err.Error())
	}
	defer fileHandle.Close()
	util.Logger.Out = fileHandle
	util.Logger.Info("Version: " + config.B.AgentVersion)

	//prepare api client
	sdk.SetDefaultLogger(util.ApiLogger)
	util.ApiClient.RequestBuilder.Config.SetDisableSSL(config.B.DisableSSL).SetGatewayHost(config.B.GatewayHost).
		SetAppMarketIdValue(config.B.AppMarketId).SetAppVersionValue(config.B.AgentVersion)
	util.ApiClient.RequestBuilder.Credential.SetAppKey(config.C.AppKey).SetSecret(config.C.Secret)

	//dispatch
	d := &Dispatcher{}
	for {
		d.Dispatch()
		time.Sleep(2 * time.Second)
	}
}
