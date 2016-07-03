package adapter

import (
	"errors"
	agadapter "github.com/OpsKitchen/ok_agent/adapter/augeas"
	"github.com/OpsKitchen/ok_agent/util"
	agsdk "honnef.co/go/augeas"
)

type Augeas struct {
	FilePath    string
	Lens        string
	OptionPath  string
	OptionValue string

	//internal fields, not for json
	fullOptionPath string
	incl           string
	lensFile       string
}

//***** interface method area *****//
func (item *Augeas) Process() error {
	var err error
	util.Logger.Debug("Processing Augeas: ", item.OptionPath, "@", item.FilePath)

	//check item data
	err = item.checkItem()
	if err != nil {
		return err
	}

	//parse item field
	err = item.parseItem()
	if err != nil {
		return err
	}

	//save Augeas
	err = item.saveAugeas()
	if err != nil {
		return err
	}

	return nil
}

func (item *Augeas) checkItem() error {
	var errMsg string

	//check file path
	if item.FilePath == "" {
		errMsg = "Config file path is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//check lens
	if item.Lens == "" {
		errMsg = "Augeas lens is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//check option path
	if item.OptionPath == "" {
		errMsg = "Config option path is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//check option value, empty value are supported in augeas 1.5
	if item.OptionValue == "" {
		errMsg = "Config option value is empty"
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (item *Augeas) parseItem() error {
	item.fullOptionPath = agadapter.ContextRoot + item.FilePath + "/" + item.OptionPath
	item.incl = agadapter.ContextRoot + item.FilePath
	item.lensFile = item.Lens + agadapter.LensSuffix
	return nil
}

//***** interface method area *****//

func (item *Augeas) saveAugeas() error {
	var ag agsdk.Augeas
	var err error
	var oldOptionValue string
	//new augeas
	ag, err = agsdk.New("/", "", agsdk.NoLoad)
	if err != nil {
		util.Logger.Error("Failed to initialize augeas sdk error: ", err.Error())
		return err
	}

	//set /augeas/load/lens and /augeas/load/incl
	err = ag.Set(agadapter.LoadPath+item.Lens+agadapter.LoadPathLens, item.lensFile)
	if err != nil {
		util.Logger.Error("Failed to set lens: ", err.Error())
		return err
	}
	err = ag.Set(agadapter.LoadPath+item.Lens+agadapter.LoadPathIncl, item.FilePath)
	if err != nil {
		util.Logger.Error("Failed to set incl: ", err.Error())
		return err
	}
	err = ag.Load()
	if err != nil {
		util.Logger.Error("Failed to load lens: ", err.Error())
		return err
	}

	//read option value
	oldOptionValue, err = ag.Get(item.fullOptionPath)
	if err == nil && oldOptionValue == item.OptionValue {
		util.Logger.Debug("Config option with correct value already exists, skip setting")
		return nil
	}

	//set option value
	err = ag.Set(item.fullOptionPath, item.OptionValue)
	if err != nil {
		util.Logger.Error("Failed to set option path and value: ", err.Error())
		return err
	}

	//save to disk
	err = ag.Save()
	ag.Close()
	if err != nil {
		util.Logger.Error("Failed to save augeas config option: ", err.Error())
		return err
	}
	util.Logger.Info(item.OptionPath + "@" + item.FilePath + " has been set to '" + item.OptionValue + "'")
	return nil
}
