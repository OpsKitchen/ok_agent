package adapter

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/adapter/file"
	"github.com/OpsKitchen/ok_agent/util"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

type File struct {
	FilePath    string
	User        string
	Group       string
	Permission  string
	FileType    string
	FileContent string
	NoTruncate  bool
	Target      string

	//internal fields, not for json
	gid, uid  uint32
	pathExist bool
	perm      os.FileMode
}

//***** interface method area *****//
func (item *File) Process() error {
	var err error
	var errMsg string
	util.Logger.Info("Processing file: ", item.FilePath)
	//check file type
	err = item.checkItem()
	if err != nil {
		return err
	}

	//parse item field
	err = item.parseItem()
	if err != nil {
		return err
	}

	//check path exist
	err = item.checkFilePathExistence()
	if err != nil {
		return err
	}

	//create parent dir
	err = item.createParentDir()
	if err != nil {
		return err
	}

	switch item.FileType {
	case file.FileTypeDir:
		return item.processDir()
	case file.FileTypeFile:
		return item.processFile()
	case file.FileTypeLink:
		return item.processLink()
	default:
		errMsg = "Unsupported file type: " + item.FileType
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func (item *File) checkItem() error {
	var errMsg string
	//check file type
	if item.FileType == "" {
		errMsg = "File type is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if item.FileType != file.FileTypeDir && item.FileType != file.FileTypeFile && item.FileType != file.FileTypeLink {
		errMsg = "Unsupported file type: " + item.FileType
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//check file path
	if item.FilePath == "" {
		errMsg = "File path is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if item.FilePath == file.FilePathRoot {
		errMsg = "File path is root"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//check symbol link target
	if item.FileType == file.FileTypeLink && item.Target == "" {
		errMsg = "Symbol link target is empty"
		util.Logger.Error(errMsg)
	}

	return nil
}

func (item *File) parseItem() error {
	var err error
	var filePerm uint64
	var gid, uid uint64
	var groupObj *user.Group
	var userObj *user.User

	//convert string permission to os.FileMode
	if item.Permission == "" {
		switch item.FileType {
		case file.FileTypeDir:
			item.perm = os.FileMode(file.DefaultPermDir)
		case file.FileTypeFile:
			item.perm = os.FileMode(file.DefaultPermFile)
		case file.FileTypeLink:
			item.perm = os.FileMode(file.DefaultPermLink)
		}
	} else {
		filePerm, err = strconv.ParseUint(item.Permission, 8, 32)
		if err != nil {
			util.Logger.Error("Invalid file mode: ", item.Permission)
			return err
		}
		item.perm = os.FileMode(filePerm)
	}

	//convert string user/group to uint32
	if item.User != "" && item.Group != "" {
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
		gid, _ = strconv.ParseUint(groupObj.Gid, 10, 32)
		uid, _ = strconv.ParseUint(userObj.Uid, 10, 32)
		item.gid = uint32(gid)
		item.uid = uint32(uid)
	}

	return nil
}

//***** interface method area *****//

func (item *File) processDir() error {
	var err error

	//create dir
	if item.pathExist == false {
		err = os.Mkdir(item.FilePath, item.perm)
		if err != nil {
			util.Logger.Error("Failed to create directory: ", item.FilePath)
			return err
		}
		util.Logger.Info("New directory created: ", item.FilePath)
	} else {
		util.Logger.Debug("Directory already exists, skip creating: ", item.FilePath)
	}

	//change permission
	err = item.changeOwnership()
	if err != nil {
		return err
	}

	//change permission
	err = item.changePermission()
	if err != nil {
		return err
	}

	return nil
}

func (item *File) processFile() error {
	var err error
	var skipWriteContent bool

	//create new file
	if item.pathExist == false {
		_, err = os.Create(item.FilePath)
		if err != nil {
			util.Logger.Error("Failed to create file: ", item.FilePath)
			return err
		}
		util.Logger.Info("New file created: ", item.FilePath)
		skipWriteContent = item.FileContent == ""
	} else {
		util.Logger.Debug("File already exists, skip creating: ", item.FilePath)
		if item.FileContent == "" { //content is empty, check if NoTruncate is true
			skipWriteContent = item.NoTruncate
		} //else, content not empty, ignore NoTruncate, skipWriteContent = false
	}

	//write content
	if skipWriteContent == false {
		err = item.writeContent()
		if err != nil {
			return err
		}
	}

	//change user and group
	err = item.changeOwnership()
	if err != nil {
		return err
	}

	//change permission
	err = item.changePermission()
	if err != nil {
		return err
	}

	return nil
}

func (item *File) processLink() error {
	var err error
	var linkTarget string

	//remove link if necessary
	if item.pathExist == true {
		linkTarget, _ = os.Readlink(item.FilePath)
		if linkTarget == item.Target {
			util.Logger.Debug("Symbol link with correct target already exists, skip creating: ", item.FilePath)
			return nil
		}
		err = os.Remove(item.FilePath)
		if err != nil {
			util.Logger.Error("Failed to remove symbol old symbol link: ", item.FilePath)
			return err
		}
	}

	//create link
	err = os.Symlink(item.Target, item.FilePath)
	if err != nil {
		util.Logger.Error("Failed to create link: ", item.FilePath)
		return err
	}
	util.Logger.Info("Symbol link created: ", item.FilePath)
	return nil
}

func (item *File) changeOwnership() error {
	var err error
	var convertedOk bool
	var stat os.FileInfo
	var stat_t *syscall.Stat_t

	if item.User != "" && item.Group != "" {
		stat, err = os.Stat(item.FilePath)
		if err == nil {
			stat_t, convertedOk = stat.Sys().(*syscall.Stat_t)
			if convertedOk {
				//user and group is already right, no need to change
				if item.gid == stat_t.Gid && item.uid == stat_t.Uid {
					util.Logger.Debug("File ownership was right, skip changing ownership: ", item.FilePath)
					return nil
				}
			}
		}

		err = os.Lchown(item.FilePath, int(item.gid), int(item.gid))
		if err != nil {
			util.Logger.Error("Failed to change owner/group to: ", item.User, "/", item.Group)
			return err
		}
		util.Logger.Info("Ownership changed to: ", item.User, "/", item.Group)
	}
	return nil
}

func (item *File) changePermission() error {
	var err error
	var stat os.FileInfo

	if item.Permission != "" {
		stat, err = os.Stat(item.FilePath)
		if stat.Mode().Perm() == item.perm {
			util.Logger.Debug("File permission was right, skip changing permission: ", item.FilePath)
			return nil
		}
		err = os.Chmod(item.FilePath, item.perm)
		if err != nil {
			util.Logger.Error("Failed to change permission: ", item.FilePath)
			return err
		}
		util.Logger.Info("File permission changed to : ", item.Permission, " ", item.FilePath)
	}
	return nil
}

func (item *File) checkFilePathExistence() error {
	var err error
	var errMsg string
	var stat os.FileInfo

	stat, err = os.Lstat(item.FilePath)
	if err != nil { //path not exist, do nothing
		return nil
	}

	switch item.FileType {
	case file.FileTypeDir:
		if stat.Mode().IsDir() == false {
			errMsg = "Path name already exists, but is not a directory: " + item.FilePath
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
	case file.FileTypeFile:
		if stat.Mode().IsRegular() == false {
			errMsg = "Path name already exists, but is not a regular file: " + item.FilePath
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
	case file.FileTypeLink:
		if stat.Mode()&os.ModeSymlink == 0 { // is not symbol link
			errMsg = "Path name already exists, but is not a symbol link: " + item.FilePath
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
	}
	item.pathExist = true
	return nil
}

func (item *File) createParentDir() error {
	var err error
	var errMsg string
	var parentDir string
	var stat os.FileInfo

	parentDir = filepath.Dir(item.FilePath)
	stat, err = os.Stat(parentDir)
	if err == nil { //path exist
		if stat.Mode().IsDir() == false {
			errMsg = "Parent directory name already exists, but is not a directory: " + parentDir
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		util.Logger.Debug("Parent directory already exists, skip creating")
		return nil
	}

	err = os.MkdirAll(parentDir, item.perm)
	if err != nil {
		util.Logger.Error("Failed to create parent directory: ", parentDir)
		return err
	}
	util.Logger.Info("Parent directory created: ", parentDir)
	return nil
}

func (item *File) writeContent() error {
	var contentBytes []byte
	var err error
	contentBytes, _ = ioutil.ReadFile(item.FilePath)
	if item.FileContent == string(contentBytes) {
		util.Logger.Debug("File already has the correct content, skip writing content: ", item.FilePath)
		return nil
	}
	err = ioutil.WriteFile(item.FilePath, []byte(item.FileContent), item.perm)
	if err != nil {
		util.Logger.Error("Failed to write content to: ", item.FilePath)
		return err
	}
	util.Logger.Info("Content written to: ", item.FilePath)
	return nil
}
