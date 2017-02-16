package task

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/adapter"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Deployer struct {
	DeployApi         *returndata.DynamicApi
	ReportResultApi   *returndata.DynamicApi
	mainLogFileHandle io.Writer
	tmpLogFileHandle  *os.File
}

func (t *Deployer) Run() error {
	var result *model.ApiResult
	var deployApiReturnData returndata.DeployApi

	//change log file to tmp file
	if err := t.changeLogFile(); err != nil {
		return t.reportResult(err)
	}

	//call deploy api
	param := &api.EntranceApiParam{ServerUniqueName: config.C.ServerUniqueName}
	util.Logger.Info("Start to call deploy api...")
	result, err := util.ApiClient.CallApi(t.DeployApi.Name, t.DeployApi.Version, param)
	util.Logger.Info("Successfully called deploy api.")
	util.Logger.Debug("Product version: " + deployApiReturnData.ProductVersion)
	util.Logger.Debug("Server name: " + deployApiReturnData.ServerName)

	//deploy api returns error
	if err != nil {
		errMsg := "Failed to call deploy api: " + err.Error()
		util.Logger.Error(errMsg)
		return t.reportResult(errors.New(errMsg))
	}
	if result.Success == false {
		errMsg := "Deploy api return error: " + result.ErrorCode + ": " + result.ErrorMessage
		util.Logger.Error(errMsg)
		return t.reportResult(errors.New(errMsg))
	}

	//deploy api returns none data filed or empty api list, do nothing
	if result.Data == nil {
		util.Logger.Debug("Deploy api return none data field.")
		return t.reportResult(nil)
	}
	result.ConvertDataTo(&deployApiReturnData)
	if len(deployApiReturnData.ApiList) == 0 {
		util.Logger.Debug("Deploy api return empty api list.")
		return t.reportResult(nil)
	}

	//call dynamic api
	for _, dynamicApi := range deployApiReturnData.ApiList {
		if err := t.processDynamicApi(dynamicApi); err != nil {
			//report failure result
			return t.reportResult(err)
			break
		}
	}

	//report success result
	return t.reportResult(nil)
}

func (t *Deployer) changeLogFile() error {
	tmpLogFileName := "/tmp/ok_agent-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".log"
	tmpLogFileHandle, err := os.OpenFile(tmpLogFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		errMsg := "Failed to create log file [" + tmpLogFileName + "]: " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	t.mainLogFileHandle = util.Logger.Out
	t.tmpLogFileHandle = tmpLogFileHandle
	util.Logger.Out = tmpLogFileHandle
	return nil
}

func (t *Deployer) processDynamicApi(dynamicApi returndata.DynamicApi) error {
	var itemList []map[string]interface{}

	//call dynamic api
	param := &api.EntranceApiParam{ServerUniqueName: config.C.ServerUniqueName}
	result, err := util.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, param)
	if err != nil {
		errMsg := "Failed to call api: " + dynamicApi.Name + ": " + dynamicApi.Version + ": " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Success == false {
		errMsg := "Dynamic api return error: " + result.ErrorCode + ": " + result.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Data == nil {
		util.Logger.Debug("Dynamic api returns none data field, nothing to do, go to next api")
		return nil
	}
	result.ConvertDataTo(&itemList)
	util.Logger.Info("Successfully called dynamic api: " + dynamicApi.Name + "/" + dynamicApi.Version)

	//cast item list to native go type
	for _, mapItem := range itemList {
		var item adapter.AdapterInterface
		var itemTypeName string
		switch dynamicApi.ReturnDataType {
		case returndata.AugeasList:
			item = &adapter.Augeas{}
			itemTypeName = "Augeas"

		case returndata.CommandList:
			item = &adapter.Command{}
			itemTypeName = "Command"

		case returndata.FileList:
			item = &adapter.File{}
			itemTypeName = "File"

		default:
			errMsg := "Unsupported list: " + dynamicApi.ReturnDataType
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}

		//data type casting with json
		err = util.JsonConvert(mapItem, &item)
		if err != nil {
			errMsg := "Failed to convert item data type"
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}

		util.Logger.Info("Start to process [ " + itemTypeName + " ]: " + item.GetBrief())
		//process item
		if err := item.Check(); err != nil {
			errMsg := "Failed to check item: " + item.GetBrief() + ": " + err.Error()
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		if err := item.Parse(); err != nil {
			errMsg := "Failed to parse item: " + item.GetBrief() + ": " + err.Error()
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		if err := item.Process(); err != nil {
			errMsg := "Failed to process item: " + item.GetBrief() + ": " + err.Error()
			util.Logger.Error(errMsg)
			util.Logger.Debug(item)
			return errors.New(errMsg)
		}
	}
	return nil
}

func (t *Deployer) reportResult(err error) error {
	defer func() {
		tmpReader, _ := os.Open(t.tmpLogFileHandle.Name())
		io.Copy(t.mainLogFileHandle, tmpReader)
		util.Logger.Out = t.mainLogFileHandle
		t.tmpLogFileHandle.Close()
		os.Remove(t.tmpLogFileHandle.Name())
	}()

	param := &api.DeployResultParam{ServerUniqueName: config.C.ServerUniqueName}
	if err != nil {
		param.ErrorMessage = err.Error()
	} else {
		param.Success = true
	}
	//read tmp log content as result data
	if t.tmpLogFileHandle != nil {
		logMsg, _ := ioutil.ReadFile(t.tmpLogFileHandle.Name())
		param.Data = string(logMsg)
	}

	result, err := util.ApiClient.CallApi(t.ReportResultApi.Name, t.ReportResultApi.Version, param)
	if err != nil {
		errMsg := "Failed to call result report api: " + t.ReportResultApi.Name + ": " + t.ReportResultApi.Version + ": " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Success == false {
		errMsg := "Result report api return error: " + result.ErrorCode + ": " + result.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	return nil
}
