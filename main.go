package main

import (
	"flag"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
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
	if util.FileExist(*baseConfigFile) {
		if err := util.ParseJsonFile(*baseConfigFile, config.B); err != nil {
			util.Logger.Fatal("can not parse base config file: " + err.Error())
		}
	}

	//prepare credential
	if !util.FileExist(config.B.CredentialFile) {
		util.Logger.Fatal("credential file not found.")
	}
	if err := util.ParseJsonFile(config.B.CredentialFile, config.C); err != nil {
		util.Logger.Fatal("credential file pasring error: " + err.Error())
	}

	//check log dir
	if err := util.PrepareLogFile(); err != nil {
		util.Logger.Fatal("can not open or create log file: " + err.Error())
	}

	//prepare api client
	util.PrepareApiClient()

	//dispatch
	for d := new(Dispatcher); ; time.Sleep(2 * time.Second) {
		d.Dispatch()
	}
}
