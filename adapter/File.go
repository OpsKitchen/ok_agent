package adapter

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/util"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	FileTypeDir  = "dir"
	FileTypeFile = "file"
	FileTypeLink = "link"

	FilePathRoot = "/"

	DefaultPermDir  = 0755
	DefaultPermFile = 0644
	DefaultPermLink = 0777
)

type File struct {
	Brief       string
	FilePath    string
	User        string
	Group       string
	Permission  string
	FileType    string
	FileContent string
	NoTruncate  bool
	Target      string
	gid, uid    uint32 //internal fields, not for json
	pathExist   bool
	perm        os.FileMode
}

//***** interface method area *****//
func (item *File) GetBrief() string {
	return item.Brief
}

func (item *File) Check() error {
	//check brief
	if item.Brief == "" {
		return errors.New("adapter: file brief is empty")
	}

	//check file type
	if item.FileType == "" {
		return errors.New("adapter: file type is empty")
	}
	if item.FileType != FileTypeDir && item.FileType != FileTypeFile && item.FileType != FileTypeLink {
		return errors.New("adapter: file type is invalid: " + item.FileType)
	}

	//check file path
	if item.FilePath == "" {
		return errors.New("adapter: file path is empty")
	}
	if item.FilePath == FilePathRoot {
		return errors.New("adapter: file path is root")
	}
	if !strings.HasPrefix(item.FilePath, "/") {
		return errors.New("adapter: file path is relative")
	}

	//check symbol link target
	if item.FileType == FileTypeLink && item.Target == "" {
		return errors.New("adapter: symbol link target is empty")
	}

	return nil
}

func (item *File) Parse() error {
	//remove trailing slash
	if item.FileType == FileTypeDir && strings.HasSuffix(item.FilePath, "/") {
		item.FilePath = item.FilePath[0 : len(item.FilePath)-1]
	}
	//convert string permission to os.FileMode
	if item.Permission == "" {
		switch item.FileType {
		case FileTypeDir:
			item.perm = os.FileMode(DefaultPermDir)
		case FileTypeFile:
			item.perm = os.FileMode(DefaultPermFile)
		case FileTypeLink:
			item.perm = os.FileMode(DefaultPermLink)
		}
	} else {
		filePerm, err := strconv.ParseUint(item.Permission, 8, 32)
		if err != nil {
			return errors.New("adapter: file permission is invalid: " + item.Permission)
		}
		item.perm = os.FileMode(filePerm)
	}

	//convert string user/group to uint32
	if item.User != "" && item.Group != "" {
		groupObj, err := user.LookupGroup(item.Group)
		if err != nil {
			return errors.New("adapter: group does not exist: " + item.Group + ": " + err.Error())
		}

		userObj, err := user.Lookup(item.User)
		if err != nil {
			return errors.New("adapter: user does not exist: " + item.User + ": " + err.Error())
		}
		gid, _ := strconv.ParseUint(groupObj.Gid, 10, 32)
		uid, _ := strconv.ParseUint(userObj.Uid, 10, 32)
		item.gid = uint32(gid)
		item.uid = uint32(uid)
	}

	return nil
}

func (item *File) Process() error {
	//check path exist
	if err := item.checkFilePathExistence(); err != nil {
		return err
	}

	//create parent dir
	if err := item.createParentDir(); err != nil {
		return err
	}

	switch item.FileType {
	case FileTypeDir:
		return item.processDir()
	case FileTypeFile:
		return item.processFile()
	case FileTypeLink:
		return item.processLink()
	}
	return nil
}

func (item *File) String() string {
	str := "\n\t\tFile path: \t" + item.FilePath +
		"\n\t\tFile type: \t" + item.FileType
	if item.User != "" {
		str += "\n\t\tUser: \t\t" + item.User
	}
	if item.Group != "" {
		str += "\n\t\tGroup: \t\t" + item.Group
	}
	if item.Permission != "" {
		str += "\n\t\tPermission: \t" + item.Permission
	}
	return str
}

//***** interface method area *****//

func (item *File) processDir() error {
	//create dir
	if item.pathExist == false {
		if err := os.Mkdir(item.FilePath, item.perm); err != nil {
			return errors.New("adapter: failed to create directory: " + err.Error())
		}
		util.Logger.Info("Succeed to create directory.")
	} else {
		util.Logger.Info("Skip creating directory, because it already exists.")
	}

	//change permission
	if err := item.changeOwnership(); err != nil {
		return err
	}

	//change permission
	if err := item.changePermission(); err != nil {
		return err
	}

	return nil
}

func (item *File) processFile() error {
	skipWriteContent := false
	//create new file
	if item.pathExist == false {
		if _, err := os.Create(item.FilePath); err != nil {
			return errors.New("adapter: failed to create file: " + err.Error())
		}
		util.Logger.Info("Succeed to create file.")
		skipWriteContent = item.FileContent == ""
	} else {
		util.Logger.Info("Skip creating, because it already exists.")
		if item.FileContent == "" { //content is empty, check if NoTruncate is true
			skipWriteContent = item.NoTruncate
		} //else, content not empty, ignore NoTruncate, skipWriteContent = false
	}

	//write content
	if skipWriteContent == false {
		if err := item.writeContent(); err != nil {
			return err
		}
	}

	//change user and group
	if err := item.changeOwnership(); err != nil {
		return err
	}

	//change permission
	if err := item.changePermission(); err != nil {
		return err
	}

	return nil
}

func (item *File) processLink() error {
	//remove link if necessary
	if item.pathExist == true {
		if linkTarget, _ := os.Readlink(item.FilePath); linkTarget == item.Target {
			util.Logger.Info("Skip creating symbol link, because it already exists with correct target .")
			return nil
		}

		if err := os.Remove(item.FilePath); err != nil {
			return errors.New("adapter: failed to remove old symbol link: " + err.Error())
		}
	}

	//create link
	if err := os.Symlink(item.Target, item.FilePath); err != nil {
		return errors.New("adapter: failed to create link: " + err.Error())
	}
	util.Logger.Info("Succeed to create symbol link.")
	return nil
}

func (item *File) changeOwnership() error {
	if item.User != "" && item.Group != "" {
		stat, err := os.Stat(item.FilePath)
		if err == nil {
			stat_t, convertedOk := stat.Sys().(*syscall.Stat_t)
			if convertedOk {
				//user and group is already right, no need to change
				if item.gid == stat_t.Gid && item.uid == stat_t.Uid {
					util.Logger.Info("Skip changing ownership, because file ownership is correct.")
					return nil
				}
			}
		}

		if err := os.Lchown(item.FilePath, int(item.gid), int(item.gid)); err != nil {
			return errors.New("adapter: failed to change ownership: " + err.Error())
		}
		util.Logger.Info("Succeed to change ownership.")
	}
	return nil
}

func (item *File) changePermission() error {
	if item.Permission != "" {
		stat, _ := os.Stat(item.FilePath)
		if stat.Mode().Perm() == item.perm {
			util.Logger.Info("Skip changing permission, because file permission is correct.")
			return nil
		}

		if err := os.Chmod(item.FilePath, item.perm); err != nil {
			return errors.New("adapter: failed to change permission: " + err.Error())
		}
		util.Logger.Info("Succeed to change file permission.")
	}
	return nil
}

func (item *File) checkFilePathExistence() error {
	stat, err := os.Lstat(item.FilePath)
	if err != nil { //path not exist, do nothing
		return nil
	}

	switch item.FileType {
	case FileTypeDir:
		if stat.Mode().IsDir() == false {
			return errors.New("adapter: path name already exists, but is not a directory")
		}
	case FileTypeFile:
		if stat.Mode().IsRegular() == false {
			return errors.New("adapter: path name already exists, but is not a regular file")
		}
	case FileTypeLink:
		if stat.Mode()&os.ModeSymlink == 0 { // is not symbol link
			return errors.New("adapter: path name already exists, but is not a symbol link")
		}
	}
	item.pathExist = true
	return nil
}

func (item *File) createParentDir() error {
	parentDir := filepath.Dir(item.FilePath)
	stat, err := os.Stat(parentDir)
	if err == nil { //path exist
		if stat.Mode().IsDir() == false {
			return errors.New("adapter: parent directory name already exists, but is not a directory: " + parentDir)
		}
		util.Logger.Info("Skip creating parent directory, because it already exists.")
		return nil
	}

	if err := os.MkdirAll(parentDir, item.perm); err != nil {
		return errors.New("adapter: failed to create parent directory: " + parentDir + "\n" + err.Error())
	}
	util.Logger.Info("Succeed to create parent directory: " + parentDir)
	return nil
}

func (item *File) writeContent() error {
	if contentBytes, _ := ioutil.ReadFile(item.FilePath); item.FileContent == string(contentBytes) {
		util.Logger.Info("Skip writing content, because it is correct.")
		return nil
	}

	if err := ioutil.WriteFile(item.FilePath, []byte(item.FileContent), item.perm); err != nil {
		return errors.New("adapter: failed to write content: " + err.Error())
	}
	util.Logger.Info("Succeed to write content.")
	return nil
}
