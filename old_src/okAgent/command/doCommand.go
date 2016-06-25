package agentCommand

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"logger"
	"os"
	"os/exec"
)

type httpCommand struct {
	User    string
	Name    string
	Command string
	Cwd     string
	Unless  string
	OnlyIf  string
	Path    string
}

/*Execute the command */
func DoCommand(cmdMap map[string]interface{}) error {
	var HttpCommand = loadHttpCommand(cmdMap)

	if HttpCommand.Unless != "" {
		if err := exCommand(HttpCommand.Unless); err == nil {
			return nil
		}
	}
	if HttpCommand.OnlyIf != "" {
		if err := exCommand(HttpCommand.OnlyIf); err != nil {
			return nil
		}
	}
	fmt.Println(HttpCommand.Command)

	oldPath := os.Getenv("PATH")
	if HttpCommand.Path != "" {
		os.Setenv("PATH", oldPath+":"+HttpCommand.Path)
	}

	in := bytes.NewBuffer(nil)
	// /bin/bash
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = in
	if HttpCommand.Command != "" {
		if HttpCommand.Cwd != "" {
			// 1.judge dir
			// 2.if on exist,mkdir -p
			// 3.cd dir
			if err := exCommand(" ls " + HttpCommand.Cwd); err != nil {
				logger.Info("cwd " + HttpCommand.Cwd + " no exist. Execute a command of \"mkdir -p " + HttpCommand.Cwd + "\"")

				if err := os.MkdirAll(HttpCommand.Cwd, os.ModePerm); err != nil {
					logger.Info("Execute the command of mkdir -p \"" + HttpCommand.Cwd + "\" failed")
					return err
				}
			}
			in.WriteString("cd " + HttpCommand.Cwd + "\n")
		}

		in.WriteString(HttpCommand.Command + "\n")
		in.WriteString("exit\n")

		stdout, _ := cmd.StdoutPipe() //standard output
		stderr, _ := cmd.StderrPipe() //err output
		cmd.Start()
		reader := bufio.NewReader(stdout)
		//实时循环读取标准输出流中的一行内容
		for {
			line, err := reader.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			fmt.Println(line)
		}
		readerErr := bufio.NewReader(stderr)
		//实时循环读取错误输出流中的一行内容
		for {
			line, err := readerErr.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			fmt.Println(line)
		}

		if err := cmd.Wait(); err != nil {
			err_string := "Execute the command \"" + HttpCommand.Command + "\" failed"
			logger.Info(err_string)
			return err
		}

		os.Setenv("PATH", oldPath)
		// if err != nil {
		// 	return errors.New("Reset path failed")
		// }
	}
	return nil
}

func loadHttpCommand(cmdMap map[string]interface{}) *httpCommand {
	var HttpCommand = &httpCommand{}

	for key, val := range cmdMap {
		mapStr := val.(string)
		switch key {
		case "user":
			if cmdMap["user"] != nil {
				HttpCommand.User = mapStr
			}
		case "name":
			if cmdMap["name"] != nil {
				HttpCommand.Name = mapStr
			}
		case "command":
			if cmdMap["command"] != nil {
				HttpCommand.Command = mapStr
			}
		case "cwd":
			if cmdMap["cwd"] != nil {
				HttpCommand.Cwd = mapStr
			}
		case "unless":
			if cmdMap["unless"] != nil {
				HttpCommand.Unless = mapStr
			}
		case "onlyIf":
			if cmdMap["onlyIf"] != nil {
				HttpCommand.OnlyIf = mapStr
			}
		case "path":
			if cmdMap["path"] != nil {
				HttpCommand.Path = mapStr
			}
		default:
		}
	}
	return HttpCommand

}

func exCommand(command string) error {
	in := bytes.NewBuffer(nil)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdin = in
	if err := cmd.Run(); err != nil {
		// logger.Info("Execute the command \"" + command + "\" failed")
		return err
	}
	return nil
}
