package agentAugeas

import (
	"errors"
	"logger"
    "honnef.co/go/augeas"
    "fmt"
)


type httpAugeas struct {
	FilePath    string
	SetKey 		string
	SetValue    string
	SetKey2 	string
	SetValue2   string
	OptionType  string
}

func loadHttpAugeas(confFileMap map[string]interface{}) *httpAugeas {
	var ha = &httpAugeas{}

	for key, val := range confFileMap {
		mapStr := val.(string)
		switch key {
		case "filePath":
			if confFileMap["filePath"] != nil {
				ha.FilePath = mapStr
			}
		case "optionKey":
			if confFileMap["optionKey"] != nil {
				ha.SetKey  = mapStr
			}
		case "optionValue":
			if confFileMap["optionValue"] != nil {
				ha.SetValue = mapStr
			}
		case "optionKey2":
			if confFileMap["optionKey2"] != nil {
				ha.SetKey2  = mapStr
			}
		case "optionValue2":
			if confFileMap["optionValue2"] != nil {
				ha.SetValue2 = mapStr
			}
		case "optionType":
			if confFileMap["optionType"] != nil {
					ha.OptionType = mapStr
				}
		default:
		}
	}
	return ha
}


func DoAugeas(confFileMap map[string]interface{}) error {
	var ha = loadHttpAugeas(confFileMap)
	if ha.FilePath == "" {
		logger.Info("The augeas file path can't be empty, please check your configuration")
		return errors.New("empty of augeas file path")
	}
	ag, err := augeas.New("/", "", augeas.None); if err != nil {
		logger.Info(err)
		return err
	}
	if ha.OptionType == "ADD" || ha.OptionType == "SET" || ha.OptionType == "" {
		fmt.Println("setting ", ha.FilePath , ha.SetKey , ha.SetValue ," now ... ")
		err = ag.Set("/files/"+ha.FilePath+ "/"+ha.SetKey , ha.SetValue); if err != nil {
			logger.Info(err)
			return err
		}
		if ha.SetKey2 != "" && ha.SetValue2 != "" {
				err = ag.Set("/files/"+ha.FilePath+ "/"+ha.SetKey2 , ha.SetValue2); if err != nil {
				logger.Info(err)
				return err
			}
		}
	}
	if ha.OptionType == "REMOVE" {
		ag.Remove("/files/"+ha.FilePath+ "/"+ha.SetKey)
	}

	err = ag.Save(); if err != nil {
		logger.Info(err)
		return err
	}
	ag.Close()
	return nil 
}

