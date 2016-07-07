package util

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"os"
)

var Logger = logrus.New()
var ApiLogger = logrus.New()

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
		return err
	}
	return json.Unmarshal(byteArray, toPointer)
}
