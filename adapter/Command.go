package adapter

import (
	"bufio"
	"errors"
	"github.com/OpsKitchen/ok_agent/util"
	"io"
	"os"
	"os/exec"
)

type Command struct {
	Command  string
	Cwd      string
	Path     string
	User     string
	RunIf    string
	NotRunIf string
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
		}
	}

	//check if necessary to run command
	if item.RunIf != "" && item.fastRun(item.RunIf) == false {
		return nil
	}
	if item.NotRunIf != "" && item.fastRun(item.NotRunIf) == true {
		return nil
	}

	//run command
	return item.runWithOutput()
}

func (item *Command) fastRun(command string) bool {
	var cmd *exec.Cmd
	var err error
	cmd = exec.Command("/bin/sh", "-c", command)
	item.setPath(cmd)
	err = cmd.Run()
	return err == nil
}

func (item *Command) runWithOutput() error {
	var cmd *exec.Cmd
	var err error
	var errPipe, outPipe io.ReadCloser
	var errReader, outReader *bufio.Reader
	var line string

	//prepare cmd object
	cmd = exec.Command("/bin/sh", "-c", item.Command)
	item.setCwd(cmd)
	item.setPath(cmd)

	outPipe, _ = cmd.StdoutPipe()
	errPipe, _ = cmd.StderrPipe()
	err = cmd.Start()
	if err != nil {
		util.Logger.Error("Can not start command: ", item.Command)
		return err
	} else {
		util.Logger.Info("Running command: ", item.Command)
	}

	//real-time output of std out
	outReader = bufio.NewReader(outPipe)
	for {
		line, err = outReader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		util.Logger.Debug(line)
	}

	//real-time output of std err
	errReader = bufio.NewReader(errPipe)
	for {
		line, err = errReader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		util.Logger.Error(line)
	}
	err = cmd.Wait()
	if err != nil {
		util.Logger.Error("Command finished, but error occourred: ", item.Command)
		return err
	} else {
		return nil
	}
}

func (item *Command) setCwd(cmd *exec.Cmd) {
	if item.Cwd != "" {
		cmd.Dir = item.Cwd
	}
}

func (item *Command) setPath(cmd *exec.Cmd) {
	if item.Path != "" {
		cmd.Env = append(cmd.Env, "PATH="+item.Path)
	} else {
		cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	}
}
