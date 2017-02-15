package adapter

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/util"
	"honnef.co/go/augeas"
)

const (
	ActionRemove = "rm"
	ActionSet    = "set"
	ContextRoot  = "/files"
	LoadPath     = "/augeas/load/"
	LoadPathIncl = "/incl"
	LoadPathLens = "/lens"
	LensSuffix   = ".lns"
)

type Augeas struct {
	Action         string
	Brief          string
	FilePath       string
	Lens           string
	OptionPath     string
	OptionValue    string
	fullOptionPath string //internal fields, not for json
	lensFile       string
}

//***** interface method area *****//
func (item *Augeas) GetBrief() string {
	return item.Brief
}

func (item *Augeas) Check() error {
	//check brief
	if item.Brief == "" {
		return errors.New("adapter: augeas brief is empty")
	}

	//check action
	if item.Action != "" && item.Action != ActionRemove && item.Action != ActionSet {
		return errors.New("adapter: augeas action is invalid")
	}

	//check file path
	if item.FilePath == "" {
		return errors.New("adapter: augeas config file path is empty")
	}

	//check lens
	if item.Lens == "" {
		return errors.New("adapter: augeas lens is empty")
	}

	//check option path
	if item.OptionPath == "" {
		return errors.New("adapter: augeas config option path is empty")
	}

	//check option value, empty value are supported in augeas 1.5
	if item.OptionValue == "" {
		return errors.New("adapter: augeas config option value is empty")
	}

	return nil
}

func (item *Augeas) Parse() error {
	if item.Action == "" {
		item.Action = ActionSet
	}
	item.fullOptionPath = ContextRoot + item.FilePath + "/" + item.OptionPath
	item.lensFile = item.Lens + LensSuffix
	return nil
}

func (item *Augeas) Process() error {
	if err := item.saveAugeas(); err != nil {
		return err
	}
	return nil
}

func (item *Augeas) String() string {
	str := "\n\t\tFile path: \t" + item.FilePath +
		"\n\t\tLens: \t\t" + item.Lens +
		"\n\t\tOption path: \t" + item.OptionPath
	if item.OptionValue != "" {
		str += "\n\t\tOption value: \t" + item.OptionValue
	}
	if item.Action != "" {
		str += "\n\t\tAction: \t" + item.Action
	}
	return str
}

//***** interface method area *****//

func (item *Augeas) saveAugeas() error {
	//new augeas
	ag, err := augeas.New("/", "", augeas.NoLoad)
	if err != nil {
		return errors.New("adapter: failed to initialize augeas sdk: " + err.Error())
	}

	//set /augeas/load/lens and /augeas/load/incl
	err = ag.Set(LoadPath+item.Lens+LoadPathLens, item.lensFile)
	if err != nil {
		return errors.New("adapter: failed to set lens: " + err.Error())
	}
	err = ag.Set(LoadPath+item.Lens+LoadPathIncl, item.FilePath)
	if err != nil {
		return errors.New("adapter: failed to set incl: " + err.Error())
	}
	err = ag.Load()
	if err != nil {
		return errors.New("adapter: failed to load lens: " + err.Error())
	}

	//remove action
	if item.Action == ActionSet { //action = "set"
		oldOptionValue, err := ag.Get(item.fullOptionPath)
		if err == nil && oldOptionValue == item.OptionValue {
			util.Logger.Info("Skip to set config option value, because it is correct.")
			return nil
		}

		//set option value
		err = ag.Set(item.fullOptionPath, item.OptionValue)
		if err != nil {
			return errors.New("adapter: failed to set option path and value: " + err.Error())
		}
		util.Logger.Debug("Successfully set " + item.fullOptionPath)
	} else if item.Action == ActionRemove { //action = "rm"
		_, err = ag.Get(item.fullOptionPath)
		if err != nil {
			util.Logger.Info("Skip to remove config option, because it does not exists.")
			return nil
		}
		num := ag.Remove(item.fullOptionPath)
		if num == 0 {
			return errors.New("adapter: failed to remove option: " + err.Error())
		}
		util.Logger.Debug("Successfully removed " + item.fullOptionPath)
	}

	//save config file change to disk
	err = ag.Save()
	ag.Close()
	if err != nil {
		return errors.New("adapter: failed to save change of config file: " + err.Error())
	}
	util.Logger.Info("Successfully saved change of config file.")
	return nil
}
