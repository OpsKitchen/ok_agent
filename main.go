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
	if err := util.ParseJsonFile(*baseConfigFile, config.B); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}

	//prepare credential
	if err := util.ParseJsonFile(config.B.CredentialFile, config.C); err != nil {
		util.Logger.Fatal("Failed to parse base config file: " + err.Error())
	}

	//check log dir
	if err := util.PrepareLogFile(); err != nil {
		util.Logger.Fatal("Failed to prepare log file: " + err.Error())
	}

	//prepare api client
	util.PrepareApiClient()

	//dispatch
	for d := new(Dispatcher); ; {
		d.Dispatch()
		time.Sleep(2 * time.Second)
	}
}
