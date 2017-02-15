package adapter

import (
	"bufio"
	"errors"
	"github.com/OpsKitchen/ok_agent/util"
	"io"
	"os"
	"os/exec"
	"os/user"
)

const (
	EnvKeyPath          = "PATH"
	DefaultShell        = "/sbin/runuser"
	DefaultUser         = "root"
	ReadStringDelimiter = '\n'
)

type Command struct {
	Brief    string
	Command  string
	Cwd      string
	Path     string
	User     string
	RunIf    string
	NotRunIf string
}

//***** interface method area *****//
func (item *Command) GetBrief() string {
	return item.Brief
}

func (item *Command) Check() error {
	//check brief
	if item.Brief == "" {
		return errors.New("adapter: command brief is empty")
	}

	//check command
	if item.Command == "" {
		return errors.New("adapter: command is empty")
	}

	//check cwd
	if item.Cwd != "" {
		stat, err := os.Stat(item.Cwd)
		if err != nil {
			return errors.New("adapter: cwd does not exist")
		} else if stat.IsDir() == false {
			return errors.New("adapter: cwd is not a directory")
		}
	}

	//check user
	if item.User != "" {
		if _, err := user.Lookup(item.User); err != nil {
			return errors.New("adapter: user does not exist: " + item.User + ": " + err.Error())
		}
	}
	return nil
}

func (item *Command) Parse() error {
	if item.User == "" {
		item.User = DefaultUser
	}
	return nil
}

func (item *Command) Process() error {
	//check if necessary to run command
	if item.RunIf != "" && item.fastRun(item.RunIf) == false {
		util.Logger.Info("Skip running, because 'RunIf' retunrs false")
		return nil
	}
	if item.NotRunIf != "" && item.fastRun(item.NotRunIf) == true {
		util.Logger.Info("Skip running, because 'NotRunIf' returns true")
		return nil
	}

	//run command
	return item.runWithMessage()
}

func (item *Command) String() string {
	str := "\n\t\tCommand: \t" + item.Command
	if item.Cwd != "" {
		str += "\n\t\tCwd: \t\t" + item.Cwd
	}
	if item.User != "" {
		str += "\n\t\tUser: \t\t" + item.User
	}
	if item.Path != "" {
		str += "\n\t\tPath: \t\t" + item.Path
	}
	if item.RunIf != "" {
		str += "\n\t\tRun if: \t" + item.RunIf
	}
	if item.NotRunIf != "" {
		str += "\n\t\tNot run if: \t" + item.NotRunIf
	}
	return str
}

//***** interface method area *****//

func (item *Command) fastRun(line string) bool {
	cmd := exec.Command(DefaultShell, item.User, "-c", line)
	item.setCwd(cmd)
	item.setPath(cmd)
	err := cmd.Run()
	return err == nil
}

func (item *Command) runWithMessage() error {
	//prepare cmd object
	cmd := exec.Command(DefaultShell, item.User, "-c", item.Command)
	item.setCwd(cmd)
	item.setPath(cmd)

	outPipe, _ := cmd.StdoutPipe()
	errPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return errors.New("adapter: failed to start default shell: " + DefaultShell + "\n" + err.Error())
	}

	//real-time output of std out
	outReader := bufio.NewReader(outPipe)
	for {
		line, err := outReader.ReadString(ReadStringDelimiter)
		if err != nil || io.EOF == err {
			break
		}
		util.Logger.Debug(line)
	}

	//real-time output of std err
	errReader := bufio.NewReader(errPipe)
	errorLineAll := ""
	for {
		line, err := errReader.ReadString(ReadStringDelimiter)
		if err != nil || io.EOF == err {
			break
		}
		errorLineAll += line
		util.Logger.Warn(line)
	}

	if err := cmd.Wait(); err != nil {
		if errorLineAll != "" {
			return errors.New("adapter: failed to run command: " + errorLineAll)
		}
		return errors.New("adapter: failed to run command: " + err.Error())
	} else {
		util.Logger.Info("Succeed to run command.")
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
		cmd.Env = append(cmd.Env, EnvKeyPath+"="+item.Path)
	} else {
		cmd.Env = append(cmd.Env, EnvKeyPath+"="+os.Getenv(EnvKeyPath))
	}
}
