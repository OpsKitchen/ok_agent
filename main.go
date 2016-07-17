package main

import (
	"flag"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/Sirupsen/logrus"
)

var DebugAgent bool

func main() {
	//parse config file from cli argument
	baseConfigFile := flag.String("c", "/etc/ok_agent.json", "base config file path")
	debugAgent := flag.Bool("d", false, "enable debug mode")
	debugApi := flag.Bool("debug-api", false, "enable open api debug log")
	flag.Parse()

	//enable agent debug mode
	if *debugAgent {
		DebugAgent = true
		util.Logger.Level = logrus.DebugLevel
	}

	//enable api log
	if *debugApi {
		util.ApiLogger.Level = logrus.DebugLevel
	}

	//dispatch
	dispatcher := &Dispatcher{
		BaseConfigFile: *baseConfigFile,
	}
	dispatcher.Dispatch()
}
