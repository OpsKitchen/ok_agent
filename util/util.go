package util

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"os"
)

var Logger *logrus.Logger = logrus.New()
var ApiLogger *logrus.Logger = logrus.New()

func FileExist(path string) bool {
	var err error
	_, err = os.Stat(path)
	return err == nil
}

func JsonConvert(fromPointer interface{}, toPointer interface{}) error {
	var byteArray []byte
	var err error
	byteArray, err = json.Marshal(fromPointer)
	if err != nil {
		Logger.Error("Failed to encode while type converting with json: " + err.Error())
		return err
	}
	err = json.Unmarshal(byteArray, toPointer)
	if err != nil {
		Logger.Error("Failed to decode while type converting with json: " + err.Error())
		return err
	}
	return nil
}
