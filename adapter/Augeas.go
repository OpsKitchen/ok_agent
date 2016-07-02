package adapter

import (
	"github.com/OpsKitchen/ok_agent/util"
	"errors"
	agadapter "github.com/OpsKitchen/ok_agent/adapter/augeas"
	agsdk "honnef.co/go/augeas"
)

type Augeas struct {
	FilePath    string
	Lens        string
	OptionPath  string
	OptionValue string

	//internal fields, not for json
	incl        string
	lensFile    string
}

//***** interface method area *****//
func (item *Augeas) Process() error {
	var err error
	util.Logger.Debug("Processing Augeas: ", item.OptionPath)

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
	item.incl = agadapter.ContextRoot + item.FilePath + "/"
	item.lensFile = item.Lens + agadapter.LensSuffix
	return nil
}

//***** interface method area *****//

func (item *Augeas) saveAugeas() error {
	ag, err := agsdk.New("/", "", agsdk.NoLoad)
	if err != nil {
		util.Logger.Error("Failed to initialize augeas sdk error: ", err.Error())
		return err
	}
	
	err = ag.Set(agadapter.LoadPath+item.Lens+agadapter.Lens, item.lensFile)
	if err != nil {
		util.Logger.Error("Failed to set lens: ", err.Error())
		return err
	}
	err = ag.Set(agadapter.LoadPath+item.Lens+agadapter.Incl, item.FilePath)
	if err != nil {
		util.Logger.Error("Failed to set incl: ", err.Error())
		return err
	}

	err = ag.Set(item.incl+item.OptionPath, item.OptionValue)
	if err != nil {
		util.Logger.Error("Failed to set option path and value: ", err.Error())
		return err
	}

	err = ag.Save()
	if err != nil {
		util.Logger.Error("Failed to save augeas config option: ", err.Error())
		return err
	}
	ag.Close()
	return nil
}