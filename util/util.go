package util

import (
	"github.com/Sirupsen/logrus"
	"os"
)

var Logger = logrus.New()

func FileExist(path string) bool {
	var err error
	_, err = os.Stat(path)
	return err == nil
}
