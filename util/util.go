package util

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/Sirupsen/logrus"
)

var ApiClient = sdk.NewClient()
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
