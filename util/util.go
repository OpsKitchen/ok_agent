package util

import (
	"github.com/Sirupsen/logrus"
	"os"
	"encoding/json"
)

var Logger = logrus.New()

func FileExist(path string) bool {
	var err error
	_, err = os.Stat(path)
	return err == nil
}

func JsonConvert(fromPointer interface{}, toPointer interface{}) error {
	var byteArray []byte
	byteArray, _ = json.Marshal(fromPointer)
	return json.Unmarshal(byteArray, toPointer)
}