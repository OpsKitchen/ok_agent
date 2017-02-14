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
	Api *returndata.DynamicApi
}

func (t *Deployer) Run() error {
	util.Logger.Info("Calling deploy api...")
	var result *model.ApiResult
	var apiResultData returndata.DeployApi

	//call deploy api
	param := &api.EntranceApiParam{ServerUniqueName: config.C.ServerUniqueName}
	result, err := util.ApiClient.CallApi(t.Api.Name, t.Api.Version, param)
	if err != nil {
		errMsg := "Failed to call deploy api: " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Success == false {
		errMsg := "deploy api return error: " + result.ErrorCode + ": " + result.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Data == nil {
		errMsg := "deploy api return empty data."
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	result.ConvertDataTo(&apiResultData)
	if len(apiResultData.ApiList) == 0 {
		errMsg := "deploy api return empty api list."
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	//change log file to tmp file
	tmpLogFileName := "/tmp/ok_agent-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".log"
	tmpLogFileHandle, err := os.OpenFile(tmpLogFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		errMsg := "Failed to create log file [" + tmpLogFileName + "]: " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	mainLogFileHandle := util.Logger.Out
	util.Logger.Out = tmpLogFileHandle
	defer func() {
		tmpReader, _ := os.Open(tmpLogFileName)
		io.Copy(mainLogFileHandle, tmpReader)
		util.Logger.Out = mainLogFileHandle
		tmpLogFileHandle.Close()
		os.Remove(tmpLogFileName)
	}()
	util.Logger.Info("Succeed to call deploy api.")
	util.Logger.Debug("Product version: " + apiResultData.ProductVersion)
	util.Logger.Debug("Server name: " + apiResultData.ServerName)

	//call dynamic api
	for _, dynamicApi := range apiResultData.ApiList {
		if err := t.processDynamicApi(dynamicApi); err != nil {
			//report failure result
			return t.reportResult(apiResultData.ReportResultApi, err, tmpLogFileHandle)
			break
		}
	}

	//report success result
	return t.reportResult(apiResultData.ReportResultApi, nil, tmpLogFileHandle)
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
		util.Logger.Debug("Dynamic api returns no data field, nothing to do, go to next api")
		return nil
	}
	result.ConvertDataTo(&itemList)
	util.Logger.Info("Succeed to call dynamic api: ", dynamicApi.Name)

	//cast item list to native go type
	for _, mapItem := range itemList {
		var item adapter.AdapterInterface
		switch dynamicApi.ReturnDataType {
		case returndata.AugeasList:
			item = &adapter.Augeas{}

		case returndata.CommandList:
			item = &adapter.Command{}

		case returndata.FileList:
			item = &adapter.File{}

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
		util.Logger.Info("Succeed to process: " + item.GetBrief())
	}
	return nil
}

func (t *Deployer) reportResult(dynamicApi returndata.DynamicApi, err error, tmpLogFileHandle *os.File) error {
	param := &api.DeployResultParam{ServerUniqueName: config.C.ServerUniqueName}
	if err != nil {
		param.ErrorMessage = err.Error()
	} else {
		param.Success = true
	}
	//read tmp log content as result data
	if tmpLogFileHandle != nil {
		logMsg, _ := ioutil.ReadFile(tmpLogFileHandle.Name())
		param.Data = string(logMsg)
	}

	result, err := util.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, param)
	if err != nil {
		errMsg := "Failed to call result report api: " + dynamicApi.Name + ": " + dynamicApi.Version + ": " + err.Error()
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
