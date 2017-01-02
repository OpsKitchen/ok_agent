package util

import (
	"encoding/json"
	"errors"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

var ApiClient = sdk.NewClient()
var ApiLogger = logrus.New()
var Logger = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: &logrus.JSONFormatter{},
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

func FileExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func JsonConvert(fromPointer interface{}, toPointer interface{}) error {
	byteArray, err := json.Marshal(fromPointer)
	if err != nil {
		Logger.Error("Failed to encode while type converting with json: " + err.Error())
		return err
	}

	if err := json.Unmarshal(byteArray, toPointer); err != nil {
		Logger.Error("Failed to decode while type converting with json: " + err.Error())
		return err
	}
	return nil
}

func ParseJsonFile(file string, out interface{}) error {
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New("util: can not read file: " + err.Error())
	}

	if err = json.Unmarshal(jsonBytes, out); err != nil {
		return errors.New("util: json decode failed: " + err.Error())
	}
	return nil
}

func PrepareApiClient() {
	sdk.SetDefaultLogger(ApiLogger)
	ApiClient.RequestBuilder.Config.SetDisableSSL(config.B.DisableSSL).SetGatewayHost(config.B.GatewayHost).
		SetGatewayPort(config.B.GatewayPort).SetGatewayPath(config.B.GatewayPath).
		SetAppMarketIdValue(config.B.AppMarketId).SetAppVersionValue(config.B.AgentVersion)

	ApiClient.RequestBuilder.Credential.SetAppKey(config.C.AppKey).SetSecret(config.C.Secret)
}

func PrepareLogFile() error {
	if _, err := os.Stat(config.B.LogDir); err != nil { //log dir not exists
		if os.MkdirAll(config.B.LogDir, 0755) != nil {
			return errors.New("Failed to create log dir [" + config.B.LogDir + "]: " + err.Error())
		}
	}
	filename := config.B.LogDir + "/info.log"
	var fileHandle *os.File
	if _, err := os.Stat(filename); err != nil { //log file not exists
		if fileHandle, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return errors.New("Failed to create log file [" + filename + "]: " + err.Error())
		}
	} else if fileHandle, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		return errors.New("Failed to open log file [" + filename + "] for writing: " + err.Error())
	}
	Logger.Out = fileHandle
	return nil
}
