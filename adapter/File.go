package adapter

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/util"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

type File struct {
	FilePath    string
	User        string
	Group       string
	Mode        string
	FileType    string
	FileContent string
	NoTruncate  bool
	Target      string
}

func (item *File) Process() error {
	var err error
	var errMsg string
	var parentDir string

	if item.FilePath == "" {
		errMsg = "File path is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if item.FilePath == "/" {
		errMsg = "File path is root"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if item.FileType == "" {
		errMsg = "File type is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	util.Logger.Debug("Processing file: ", item.FilePath)

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
		return item.processFile()
	case "dir":
		return item.processDir()
	case "link":
		return item.processLink()
	default:
		errMsg = "Unsupported file type: " + item.FileType
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func (item *File) processDir() error {
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

	err = item.changeMode()
	if err != nil {
		return err
	}
	err = item.changeOwnerAndGroup()
	if err != nil {
		return err
	}

	return nil
}

func (item *File) processFile() error {
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
		err = item.writeContent()
		if err != nil {
			return err
		}
	}

	//change permission
	if item.Mode != "" {
		err = item.changeMode()
		if err != nil {
			return err
		}
	}

	//change user and group
	if item.User != "" && item.Group != "" {
		err = item.changeOwnerAndGroup()
		if err != nil {
			return err
		}
	}

	return nil
}

func (item *File) processLink() error {
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

func (item *File) changeMode() error {
	var err error
	var modeInt int64
	var mode os.FileMode
	var stat os.FileInfo
	modeInt, err = strconv.ParseInt(item.Mode, 8, 32)
	if err != nil {
		util.Logger.Error("Invalid file mode: ", item.Mode)
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
			util.Logger.Info("File mode changed to : ", item.Mode, " ", item.FilePath)
		}
	}
	return nil
}

func (item *File) changeOwnerAndGroup() error {
	var err error
	var gid, uid int64
	var groupObj *user.Group
	var userObj *user.User

	groupObj, err = user.LookupGroup(item.Group)
	if err != nil {
		util.Logger.Error("Group does not exist: ", item.Group)
		return err
	}

	userObj, err = user.Lookup(item.User)
	if err != nil {
		util.Logger.Error("User does not exist: ", item.User)
		return err
	}
	gid, _ = strconv.ParseInt(groupObj.Gid, 10, 32)
	uid, _ = strconv.ParseInt(userObj.Uid, 10, 32)

	err = os.Chown(item.FilePath, int(uid), int(gid))
	if err != nil {
		util.Logger.Error("Failed to change owner/group to: ", item.User, "/", item.Group)
		return err
	} else {
		util.Logger.Info("Owner/group changed to: ", item.User, "/", item.Group)
		return nil
	}
}

func (item *File) writeContent() error {
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
