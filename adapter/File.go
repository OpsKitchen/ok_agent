package adapter

import (
	//go builtin pkg
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	//local pkg
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/util"
)

type File struct {
	itemList []returndata.File
}

func (fileAdapter *File) CastItemList(list interface{}) error {
	var dataByte []byte
	dataByte, _ = json.Marshal(list)
	return json.Unmarshal(dataByte, &fileAdapter.itemList)
}

func (fileAdapter *File) Process() error {
	var err error
	var item returndata.File
	for _, item = range fileAdapter.itemList {
		err = fileAdapter.processItem(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fileAdapter *File) processItem(item returndata.File) error {
	var err error
	var parentDir string

	if item.FilePath == "" {
		util.Logger.Error("File path is empty")
		return errors.New("File path is empty")
	}
	if item.FilePath == "/" {
		util.Logger.Error("File path is root")
		return errors.New("File path is root")
	}
	if item.FileType == "" {
		util.Logger.Error("File type is empty")
		return errors.New("File type is empty")
	}

	//create parent dir
	parentDir = filepath.Dir(item.FilePath)
	if util.FileExist(parentDir) == false {
		err = os.MkdirAll(parentDir, 0755)
		if err != nil {
			util.Logger.Error("Failed to create parent directory: ", parentDir)
			return err
		} else {
			util.Logger.Info("Parent directory created: ", parentDir)
		}
	}

	switch item.FileType {
	case "file":
		return fileAdapter.processFile(item)
	case "dir":
		return fileAdapter.processDir(item)
	case "link":
		return fileAdapter.processLink(item)
	}
	return nil
}

func (fileAdapter *File) processDir(item returndata.File) error {
	var err error
	//create dir
	if util.FileExist(item.FilePath) == false {
		err = os.Mkdir(item.FilePath, 0755)
		if err != nil {
			util.Logger.Error("Failed to create directory: ", item.FilePath)
			return err
		} else {
			util.Logger.Info("New directory created: ", item.FilePath)
		}
	}

	err = fileAdapter.changeMode(item)
	if err != nil {
		return err
	}
	err = fileAdapter.changeGroup(item)
	if err != nil {
		return err
	}
	err = fileAdapter.changeOwner(item)
	if err != nil {
		return err
	}

	return nil
}

func (fileAdapter *File) processFile(item returndata.File) error {
	var err error
	var fileExist bool
	var skipWriteContent bool

	fileExist = util.FileExist(item.FilePath)

	//create new file
	if fileExist == false {
		_, err = os.Create(item.FilePath)
		if err != nil {
			util.Logger.Error("Failed to create file: ", item.FilePath)
			return err
		} else {
			util.Logger.Info("New file created: ", item.FilePath)
		}
	}

	//write content
	if fileExist == true {
		if item.FileContent == "" { //content is empty, check if NoTruncate is true
			skipWriteContent = item.NoTruncate
		} // else, content not empty, ignore NoTruncate, skipWriteContent = false
	} else {
		skipWriteContent = item.FileContent == ""
	}
	if skipWriteContent == false {
		err = fileAdapter.writeContent(item)
		if err != nil {
			return err
		}
	}

	//change permission
	if item.Mode != "" {
		err = fileAdapter.changeMode(item)
		if err != nil {
			return err
		}
	}

	//change user and group
	if item.User != "" {
		err = fileAdapter.changeGroup(item)
		if err != nil {
			return err
		}
	}
	if item.UserGroup != "" {
		err = fileAdapter.changeOwner(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fileAdapter *File) processLink(item returndata.File) error {
	var err error
	if item.Target == "" {
		util.Logger.Error("Link target is empty")
		return errors.New("Link target is empty")
	}

	//remove link if exists
	if util.FileExist(item.FilePath) == true {
		err = os.Remove(item.FilePath)
		if err != nil {
			util.Logger.Error("Failed to remove old link: ", item.FilePath)
			return err
		} else {
			util.Logger.Info("Old link removed: ", item.FilePath)
		}
	}

	//create link
	err = os.Symlink(item.Target, item.FilePath)
	if err != nil {
		util.Logger.Error("Failed to create link: ", item.FilePath)
		return err
	} else {
		util.Logger.Info("New symbol link created: ", item.FilePath)
	}
	return nil
}

func (fileAdapter *File) changeMode(item returndata.File) error {
	var err error
	var modeInt int64
	var mode os.FileMode
	var stat os.FileInfo
	modeInt, err = strconv.ParseInt(item.Mode, 8, 32)
	if err != nil {
		util.Logger.Error("Invalid file mode: ", item.Mode, item.FilePath)
		return err
	}

	mode = os.FileMode(modeInt)
	stat, _ = os.Lstat(item.FilePath)
	if stat.Mode().Perm() != mode {
		err = os.Chmod(item.FilePath, mode)
		if err != nil {
			util.Logger.Error("Failed to change mode: ", item.FilePath)
			return err
		} else {
			util.Logger.Info("File mode changed to : ", item.Mode, item.FilePath)
		}
	}
	return nil
}

func (fileAdapter *File) changeGroup(item returndata.File) error {

	return nil
}

func (fileAdapter *File) changeOwner(item returndata.File) error {

	return nil
}

func (fileAdapter *File) writeContent(item returndata.File) error {
	var contentBytes []byte
	var err error
	contentBytes, _ = ioutil.ReadFile(item.FilePath)
	if item.FileContent != string(contentBytes) {
		err = ioutil.WriteFile(item.FilePath, []byte(item.FileContent), 0644)
		if err != nil {
			util.Logger.Error("Failed to write content to: ", item.FilePath)
			return err
		} else {
			util.Logger.Info("Content written to: ", item.FilePath)
		}
	}
	return nil
}
