package task

import (
	"errors"
	"github.com/OpsKitchen/ok_agent/adapter"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/util"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
)

type Deployer struct {
	Api *returndata.DynamicApi
}

func (t *Deployer) Run() error {
	util.Logger.Info("Calling deploy api")
	var result *model.ApiResult
	var apiResultData returndata.DeployApi

	result, err := util.ApiClient.CallApi(t.Api.Name, t.Api.Version, nil)
	if err != nil {
		util.Logger.Debug("Failed to call deploy api.")
		return err
	}
	if result.Success == false {
		errMsg := "deploy api return error: " + result.ErrorCode + "\t" + result.ErrorMessage
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
	util.Logger.Info("Succeed to call deploy api.")
	util.Logger.Info("Product version: " + apiResultData.ProductVersion)
	util.Logger.Info("Server name: " + apiResultData.ServerName)
	for _, dynamicApi := range apiResultData.ApiList {
		if err := t.processDynamicApi(dynamicApi); err != nil {
			return err
		}
	}
	util.Logger.Info("Succeed to run all deploy task.")
	return nil
}

func (t *Deployer) processDynamicApi(dynamicApi returndata.DynamicApi) error {
	util.Logger.Info("Calling dynamic api: ", dynamicApi.Name)
	var itemList []map[string]interface{}

	//call dynamic api
	result, err := util.ApiClient.CallApi(dynamicApi.Name, dynamicApi.Version, nil)
	if err != nil {
		util.Logger.Error("Failed to call api: " + dynamicApi.Name + "\t" + dynamicApi.Version)
		return err
	}
	if result.Success == false {
		errMsg := "api return error: " + result.ErrorCode + "\t" + result.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if result.Data == nil {
		util.Logger.Info("Api returns empty data, nothing to do, go to next api")
		return nil
	}
	result.ConvertDataTo(&itemList)

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
		util.Logger.Info("Processing..." + item.Brief())
		if item.Check() != nil && item.Parse() != nil && item.Process() != nil {
			errMsg := "Failed to process item"
			util.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
	} //end for "range apiResultData"
	util.Logger.Info("Succeed to process dynamic api: ", dynamicApi.Name+"\t"+dynamicApi.Version)
	return nil
}
