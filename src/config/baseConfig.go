/*
* go http server config file
*
 */

package config

import ()

type baseConfig struct {
	GW_HOST      string
	Content_Type string
	OA_Session_Id    string
	OA_App_Market_ID string
	OA_App_Version   string
	OA_Device_Id     string

	API_VERSION    string
	JsonConfPath   string
	BASE_API_NAME  string
	LOGER_CWD      string
	LOGER_FILE     string
	TEMP_JSON_PATH string
	TEMP_JSON_EX   string

	TEMP_JSON_FILE_PATH string
	AGENT_VERSION  string
}

// init default server config
var BaseConfig = &baseConfig{
	GW_HOST: "https://api.opskitchen.com/gw/json/",
	BASE_API_NAME: "ops.server.listAgentApi",

	Content_Type:     "application/x-www-form-urlencoded",
	OA_Session_Id:    "",
	OA_App_Market_ID: "678",
	OA_App_Version:   "1.0",
	OA_Device_Id:     "282459052",
	API_VERSION:      "1.0",
	JsonConfPath:     "/root/.ok_agent/agent.json",

	LOGER_CWD:      "/var/log/ok_agent/",
	LOGER_FILE:     "agent.log",
	TEMP_JSON_PATH: "/root/.ok_agent/",
	TEMP_JSON_EX:   "temp_",

	TEMP_JSON_FILE_PATH: "/root/.tmp_json/",
	AGENT_VERSION: "1.0.0",
}

