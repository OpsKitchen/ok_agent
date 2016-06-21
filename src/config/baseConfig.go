/*
* go http server config file
*
 */

package config

import ()

type baseConfig struct {
	GW_HOST          string
	CONTENT_TYPE     string
	OA_SESSION_ID    string
	OA_APP_MARKET_ID string
	OA_APP_VERSION   string
	OA_DEVICE_ID     string

	API_VERSION    string
	JSON_CONF_PATH string
	BASE_API_NAME  string
	LOGER_CWD      string
	LOGER_FILE     string
	TEMP_JSON_PATH string
	TEMP_JSON_EX   string

	AGENT_REPO_BACKUP_DIR string
	SOURCE_REPO_DIR       string

	TEMP_JSON_FILE_PATH string
	AGENT_VERSION       string
}

// init default server config
var BaseConfig = &baseConfig{
	GW_HOST:       "https://api.opskitchen.com/gw/json/",
	BASE_API_NAME: "ops.server.listAgentApi",

	CONTENT_TYPE:     "application/x-www-form-urlencoded",
	OA_SESSION_ID:    "",
	OA_APP_MARKET_ID: "678",
	OA_APP_VERSION:   "1.0",
	OA_DEVICE_ID:     "282459052",
	API_VERSION:      "1.0",
	JSON_CONF_PATH:   "/root/.ok_agent/agent.json",

	LOGER_CWD:      "/var/log/ok_agent/",
	LOGER_FILE:     "agent.log",
	TEMP_JSON_PATH: "/root/.ok_agent/",
	TEMP_JSON_EX:   "temp_",

	AGENT_REPO_BACKUP_DIR: "/root/backup-repo",
	SOURCE_REPO_DIR:       "/etc/yum.repos.d",

	TEMP_JSON_FILE_PATH: "/root/.tmp_json/",
	AGENT_VERSION:       "1.0.0",
}
