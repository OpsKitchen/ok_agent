package adapter

import (
	"os/exec"
	"github.com/OpsKitchen/ok_agent/util"
	"errors"
	"os"
)

type Command struct {
	Command string
	Cwd     string
	Path    string
	User    string
	OnlyIf  string
	Unless  string
}

func (item *Command) Process() error {
	var err error
	var stat os.FileInfo
	if item.Command == "" {
		util.Logger.Error("Command is empty")
		return errors.New("Command is empty")
	}

	util.Logger.Debug("Processing command: ", item.Command)

	//check cwd
	if item.Cwd != "" {
		stat, err = os.Stat(item.Cwd)
		if err != nil {
			util.Logger.Error("Cwd does not exist: ", item.Cwd)
			return errors.New("Cwd does not exist: " + item.Cwd)
		} else if stat.IsDir() == false {
			util.Logger.Error("Cwd is not a directory: ", item.Cwd)
			return errors.New("Cwd is not a directory: " + item.Cwd)
		}//@todo else if item.User has no permission on this dir?
	}

	//check if necessary to run command
	if item.OnlyIf != "" && item.check(item.OnlyIf) == false {
		return nil
	}
	if item.Unless != "" && item.check(item.Unless) == true {
		return nil
	}

	//run command
	util.Logger.Info("Runing command: ", item.Command)
	return nil
}

func (item *Command) check(command string) bool {
	var cmd *exec.Cmd
	var err error
	cmd = exec.Command("/bin/sh", "-c", command)
	item.setCwd(cmd)
	item.setPath(cmd)
	err = cmd.Run()
	return err == nil
}

func (item *Command) setCwd(cmd *exec.Cmd) {
	if item.Cwd != "" {
		cmd.Dir = item.Cwd
	}
}

func (item *Command) setPath(cmd *exec.Cmd) {
	if item.Path != "" {
		cmd.Env = append(cmd.Env, "PATH="+item.Path)
	}
}