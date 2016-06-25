package agentFile

import (
	"errors"
	"fmt"
	"io"
	"logger"
	"okAgent/command"
	"os"
	"strings"
)

type httpFile struct {
	FilePath    string
	FileContent string
	FileType    string
	User        string
	UserGroup   string
	Mode        string
	Target      string
}

func loadHttpFile(cmdMap map[string]interface{}) *httpFile {
	var HttpFile = &httpFile{}

	for key, val := range cmdMap {
		mapStr := val.(string)
		switch key {
		case "filePath":
			if mapStr != "" {
				HttpFile.FilePath = mapStr
			}
		case "fileContent":
			if mapStr != "" {
				HttpFile.FileContent = mapStr
			}
		case "fileType":
			if mapStr != "" {
				HttpFile.FileType = mapStr
			}
		case "owner":
			if mapStr != "" {
				HttpFile.User = mapStr
			}
		case "userGroup":
			if mapStr != "" {
				HttpFile.UserGroup = mapStr
			}
		case "mode":
			if mapStr != "" {
				HttpFile.Mode = mapStr

			}
		case "target":
			if mapStr != "" {
				HttpFile.Target = mapStr
			}
		default:
		}
	}
	return HttpFile
}

func DoFile(cmdMap map[string]interface{}) error {
	var httpFile = loadHttpFile(cmdMap)

	if httpFile.FilePath == "" {
		logger.Info("Failed to create a empty filePath ")
		return nil
	}

	if httpFile.FileType == "" {
		logger.Info("The file type can't be empty, please check your configuration")
		return errors.New("empty of file type")
	}

	switch httpFile.FileType {
	case "file":
		var f *os.File
		var err1 error
		// 1、mkdir -p cwd && touch fileName
		fileName := httpFile.FilePath
		filePathArray := strings.Split(httpFile.FilePath, "/") //split dir
		cwd := ""
		if len(filePathArray) > 1 {
			fileName = filePathArray[len(filePathArray)-1]
			fileNameArray := strings.Split(httpFile.FilePath, fileName)
			if len(fileNameArray) > 1 {
				cwd = fileNameArray[0]
			}
		}
		if cwd != "" {
			err := os.MkdirAll(cwd, os.ModePerm)
			if err != nil {
				return err
			}
		}

		f, err1 = os.Create(httpFile.FilePath)  //creat a file and cover the file
		fmt.Println("touch", httpFile.FilePath) //only for a note
		if err1 != nil {
			logger.Info("Failed to create " + httpFile.FilePath)
			return err1
		}

		io.WriteString(f, httpFile.FileContent) //writing content
		if err1 != nil {
			logger.Info("Failed to Write " + httpFile.FilePath + " with content \"" + httpFile.FileContent + "\"")
			return err1
		}

		// 修改文件模式
		if httpFile.Mode != "" {
			fileModeMap := map[string]interface{}{"command": "chmod " + httpFile.Mode + " " + httpFile.FilePath}
			err := agentCommand.DoCommand(fileModeMap)
			if err != nil {
				return err
			}
		}

		//chown (os.Chown need gid and uid)
		if httpFile.User != "" || httpFile.UserGroup != "" {
			fileUserGroupMap := make(map[string]interface{})
			if httpFile.User != "" && httpFile.UserGroup == "" {
				fileUserGroupMap = map[string]interface{}{"command": "chown " + httpFile.User + " " + httpFile.FilePath}
			}
			if httpFile.User == "" && httpFile.UserGroup != "" {
				fileUserGroupMap = map[string]interface{}{"command": "chgrp " + httpFile.UserGroup + " " + httpFile.FilePath}
			}
			if httpFile.User != "" && httpFile.UserGroup != "" {
				fileUserGroupMap = map[string]interface{}{"command": "chown " + httpFile.User + ":" + httpFile.UserGroup + " " + httpFile.FilePath}
			}
			err := agentCommand.DoCommand(fileUserGroupMap)
			if err != nil {
				return err
			}
		}

	case "dir":
		if httpFile.FilePath != "" {
			// mkdir
			err := os.MkdirAll(httpFile.FilePath, os.ModePerm)
			if err != nil {
				return err
			}

			// chmod
			if httpFile.Mode != "" {
				fileModeMap := map[string]interface{}{"command": "chmod " + httpFile.Mode + " " + httpFile.FilePath}
				err := agentCommand.DoCommand(fileModeMap)
				if err != nil {
					return err
				}
			}

			//chown (os.Chown need gid and uid)
			if httpFile.User != "" || httpFile.UserGroup != "" {
				fileUserGroupMap := make(map[string]interface{})
				if httpFile.User != "" && httpFile.UserGroup == "" {
					fileUserGroupMap = map[string]interface{}{"command": "chown " + httpFile.User + " " + httpFile.FilePath}
				}
				if httpFile.User == "" && httpFile.UserGroup != "" {
					fileUserGroupMap = map[string]interface{}{"command": "chgrp " + httpFile.UserGroup + " " + httpFile.FilePath}
				}
				if httpFile.User != "" && httpFile.UserGroup != "" {
					fileUserGroupMap = map[string]interface{}{"command": "chown " + httpFile.User + ":" + httpFile.UserGroup + " " + httpFile.FilePath}
				}
				err := agentCommand.DoCommand(fileUserGroupMap)
				if err != nil {
					return err
				}
			}

		}
	case "link":
		// symlink os.Symlink no support windows,but only linux and unix
		if httpFile.FilePath != "" && httpFile.Target != "" {
			if _, err := os.Stat(httpFile.Target); err == nil {
				os.Remove(httpFile.Target)
			}
			fmt.Println("ln -sf", httpFile.FilePath, httpFile.Target)
			err := os.Symlink(httpFile.FilePath, httpFile.Target)
			if err != nil {
				return err
			}

		}

	default:
	}
	return nil

}
