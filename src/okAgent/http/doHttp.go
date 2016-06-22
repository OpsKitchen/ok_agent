package agentHttp

import (
	"config"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"logger"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type httpConf struct {
	SecretKey    string
	OA_App_Key   string
	AgentVersion string
	ApiName      string
	Params       string
}

/*Create an Http request */
func DoHttpRequest(requestType string, apiName string) (string, error) {
	client := &http.Client{}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	params := getParamString(apiName) //获取参数的拼接
	paramString := "api=" + apiName + "&version=" + string(config.BaseConfig.API_VERSION) +"&timestamp=" + string(timestamp) + "&params=" + params
	req, err := http.NewRequest(requestType, config.BaseConfig.GW_HOST, strings.NewReader(paramString))
	if err != nil {
		logger.Info("The request failed")
		return "", err
	}
	req.Header.Set("Content-Type", config.BaseConfig.CONTENT_TYPE)
	req.Header.Set("OA-Session-Id", config.BaseConfig.OA_SESSION_ID)
	req.Header.Set("OA-App-Market-ID", config.BaseConfig.OA_APP_MARKET_ID)
	req.Header.Set("OA-App-Version", config.BaseConfig.OA_APP_VERSION)
	req.Header.Set("OA-Device-Id", config.BaseConfig.OA_DEVICE_ID)

	HttpConf := ConfigLoadAndGet(apiName) //base for apiName
	req.Header.Set("OA-App-Key", HttpConf.OA_App_Key)
	req.Header.Set("OA-Sign", getOaSign(apiName, timestamp))

	resp, err := client.Do(req)
	if err != nil {
		logger.Info("Http request failed")
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Info("Your " + apiName + " request failed")
		return "", err
	}
	return string(body), nil
}

/* Method of MD5 */
func httpMd5(str string) (md5str string) {
	h := md5.New()
	io.WriteString(h, str)
	md5str = fmt.Sprintf("%x", h.Sum(nil))
	return
}

/* Use MD5 encryption parameters */
func getOaSign(apiName string, timestamp string) string {
	HttpConf := ConfigLoadAndGet(apiName)
	md5str := HttpConf.SecretKey + apiName + string(config.BaseConfig.API_VERSION)+getParamString(apiName) + string(timestamp)
	OASign := httpMd5(md5str)
	return OASign
}

/* Assembly parameters */
func getParamString(apiName string) string {
	HttpConf := ConfigLoadAndGet(apiName)
	var paramString = ""
	if apiName != config.BaseConfig.BASE_API_NAME || apiName == "" {
		paramString = HttpConf.Params
	}
	return paramString
}

/* load json config */
func ConfigLoadAndGet(apiName string) (HttpConf *httpConf) {
	bytes, err := ioutil.ReadFile(config.BaseConfig.JSON_CONF_PATH)
	if err != nil {
		exitAgent(err, "Failed to parse json file")
	}
	HttpConf = new(httpConf)
	if err := json.Unmarshal(bytes, &HttpConf); err != nil {
		exitAgent(err, "Failed to parse json file")
	}
	if HttpConf.AgentVersion != config.BaseConfig.AGENT_VERSION {
		err := errors.New("Running Failed!! Your agent version does't match the opskitchen agent version number!")
		exitAgent(err, "Agent version no match")
	}
	return HttpConf
}

func exitAgent(err error, errMsg string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	fmt.Println("** Exception error occur. "+errMsg+" **", err)
	os.Exit(1)
}
