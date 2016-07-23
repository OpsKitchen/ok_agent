package util

import (
	"encoding/json"
	"errors"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

var ApiClient = sdk.NewClient()
var ApiParam = &api.RequestParam{}
var ApiLogger = logrus.New()
var Logger = logrus.New()

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
	if _, err := os.Stat(file); err != nil {
		return errors.New("util: file not found: " + err.Error())
	}

	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New("util: file not readable: " + err.Error())
	}

	if err = json.Unmarshal(jsonBytes, out); err != nil {
		return errors.New("util: json decode failed: " + err.Error())
	}
	return nil
}
