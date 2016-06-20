package agentAugeas

import (
	"errors"
	"logger"
    "honnef.co/go/augeas"
    "fmt"
)

const AugeasLoadPath = "/augeas/load/"
const AugeasLens = "/lens"
const AugeasLensSuffix = ".lns"
const AugeasIncl = "/incl[last()+1]"

type httpAugeas struct {
	FilePath    string
	SetKey 		string
	SetValue    string
	Context 	string
	Lens   string
	Action  string
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
		case "context":
			if confFileMap["context"] != nil {
				ha.Context  = mapStr
			}
		case "lens":
			if confFileMap["lens"] != nil {
				ha.Lens = mapStr
			}
		case "action":
			if confFileMap["action"] != nil {
					ha.Action = mapStr
				}
		default:
		}
	}
	return ha
}


func DoAugeas(confFileMap map[string]interface{}) error {
	var ha = loadHttpAugeas(confFileMap)
	if ha.Context == "" {
		logger.Info("The augeas file path can't be empty, please check your configuration")
		return errors.New("empty of augeas file path")
	}
	if ha.Lens != "" {
		err :=doAugeasNoLoad(ha); if err != nil {
			return err
		}
	} else {
		err :=doAugeasNone(ha); if err != nil {
			return err
		}
	}
	return nil 
}


func doAugeasNone(ha *httpAugeas) error {
	ag, err := augeas.New("/", "" , augeas.None); if err != nil {
		logger.Info(err)
		return err
	}
	if ha.Action == "ADD" || ha.Action == "SET" || ha.Action == "" {
		fmt.Println("setting ", ha.Context+ ha.SetKey, ha.SetValue ," now ... ")
		err = ag.Set(ha.Context + ha.SetKey ,ha.SetValue); if err != nil {
			logger.Info(err)
			return err
		}
	}
	if ha.Action == "REMOVE" {
		fmt.Println("remove ", ha.Context , ha.SetKey ," now ... ")
		ag.Remove(ha.Context + ha.SetKey)
	}

	err = ag.Save(); if err != nil {
		logger.Info(err)
		return err
	}
	ag.Close()
	return nil
}

func doAugeasNoLoad(ha *httpAugeas) error {
	ag, err := augeas.New("/", "" , augeas.NoLoad); if err != nil {
		logger.Info(err)
		return err
	}
	err = ag.Set(AugeasLoadPath + ha.Lens + AugeasLens, ha.Lens + AugeasLensSuffix );if err != nil {
		logger.Info(err)
		return err
	}
	err = ag.Set(AugeasLoadPath + ha.Lens + AugeasIncl, ha.FilePath);if err != nil {
		logger.Info(err)
		return err
	}
	err = ag.Load(); if err != nil {
		logger.Info(err)
		return err
	}

	if ha.Action == "ADD" || ha.Action == "SET" || ha.Action == "" {
		fmt.Println("setting ", ha.Context+ ha.SetKey, ha.SetValue ," now ... ")
		err = ag.Set(ha.Context + ha.SetKey ,ha.SetValue); if err != nil {
			logger.Info(err)
			return err
		}
	}
	if ha.Action == "REMOVE" {
		fmt.Println("remove ", ha.Context , ha.SetKey ," now ... ")
		ag.Remove(ha.Context + ha.SetKey)
	}

	err = ag.Save(); if err != nil {
		logger.Info(err)
		return err
	}
	ag.Close()
	return nil
}

