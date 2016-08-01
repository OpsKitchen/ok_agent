package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var B = &Base{
	AgentVersion:       "1.1.1",
	AppMarketId:        "520",
	CredentialFile:     "/root/.ok_agent/credential.json",
	EntranceApiName:    "ops.agent.entrance",
	EntranceApiVersion: "1.0",
	GatewayHost:        "api.OpsKitchen.com",
	LogDir:             "/var/log/ok_agent",
}

var C = &Credential{}

func ParseBaseConfig(file string) error {
	return parseJsonFile(file, B)
}

func ParseCredential() error {
	return parseJsonFile(B.CredentialFile, C)
}

func parseJsonFile(file string, out interface{}) error {
	if _, err := os.Stat(file); err != nil {
		return errors.New("util: file not found: " + err.Error())
	}

	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New("util: file not readable: " + err.Error())
	}

	if err = json.Unmarshal(jsonBytes, out); err != nil {
		return errors.New("util: json decode failed: " + err.Error())
	}
	return nil
}
