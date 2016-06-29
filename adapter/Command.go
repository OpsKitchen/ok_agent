package adapter

import (
	"encoding/json"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"os/exec"
)

type Command struct {
	itemList []returndata.Command
}

func (commandAdapter *Command) CastItemList(list interface{}) error {
	var dataByte []byte
	dataByte, _ = json.Marshal(list)
	return json.Unmarshal(dataByte, &commandAdapter.itemList)
}

func (commandAdapter *Command) Process() error {
	var err error
	var item returndata.Command
	for _, item = range commandAdapter.itemList {
		err = commandAdapter.processCommand(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (commandAdapter *Command) processCommand(item returndata.Command) error {

	return nil
}

func (commandAdapter *Command) executeCommand(commandString string) error {
	cmd := exec.Command("/bin/sh", "-c", commandString)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
