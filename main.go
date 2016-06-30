package main

import (
	//go builtin pkg
	"flag"

	//local pkg
	"github.com/Sirupsen/logrus"
	"github.com/OpsKitchen/ok_agent/util"
)

func main() {
	var baseConfigFile *string
	var debugMode *bool
	var dispatcher *Dispatcher

	//parse config file from cli argument
	baseConfigFile = flag.String("c", "/etc/ok_agent.json", "base config file path")
	debugMode = flag.Bool("d", false, "enable debug log")
	flag.Parse()

	//create debug logger
	if *debugMode {
		util.Logger.Level = logrus.DebugLevel
	}

	dispatcher = &Dispatcher{
		BaseConfigFile: *baseConfigFile,
	}
	dispatcher.Dispatch()
}
