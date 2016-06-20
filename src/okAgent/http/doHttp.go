package agentHttp

import (
	"crypto/md5"
	"fmt"
	"io"

	"io/ioutil"
	"net/http"
	"strings"

	"errors"
	"strconv"
	"time"

	"github.com/pantsing/goconf"
	"logger"
	"config"
)

type httpConfJson struct {
	SecretKey  string
	OA_App_Key string
	AgentVersion string
	Api_name   string
	Params     string
}

type HttpBody struct {
	paramString string
	timestamp   int64
}

var httpbody = &HttpBody{
	paramString: "",
	timestamp:   time.Now().Unix(),
}

/*Create an Http request */
func DoHttpRequest(requestType string, apiName string) (string, error) {
	client := &http.Client{}
	// timestamp := strconv.FormatInt(httpbody.timestamp, 10)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	params := getParamString(apiName) //获取参数的拼接
	paramString := "api=" + apiName + "&version=" + string(config.BaseConfig.API_VERSION) + "&timestamp=" + string(timestamp) + "&params=" + params
	req, err := http.NewRequest(requestType, config.BaseConfig.GW_HOST, strings.NewReader(paramString))
	if err != nil {
		logger.Info("The request failed")
		return "", err
	}
	req.Header.Set("Content-Type", config.BaseConfig.Content_Type)
	req.Header.Set("OA-Session-Id", config.BaseConfig.OA_Session_Id)
	req.Header.Set("OA-App-Market-ID", config.BaseConfig.OA_App_Market_ID)
	req.Header.Set("OA-App-Version", config.BaseConfig.OA_App_Version)
	req.Header.Set("OA-Device-Id", config.BaseConfig.OA_Device_Id) //new Fingerprint({canvas, true}).get()

	HttpConfJson := ConfigLoadAndGet(apiName) //根据apiName获得配置信息
	req.Header.Set("OA-App-Key", HttpConfJson.OA_App_Key)
	req.Header.Set("OA-Sign", getOaSign(apiName,timestamp)) // sessionStorage.getItem(signKey)

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
func getOaSign(apiName string,timestamp string) string {
	// timestamp := strconv.FormatInt(httpbody.timestamp, 10)
	HttpConfJson := ConfigLoadAndGet(apiName)
	md5str := HttpConfJson.SecretKey + apiName + string(config.BaseConfig.API_VERSION) + getParamString(apiName) + string(timestamp)
	OASign := httpMd5(md5str)
	return OASign
}

/* Assembly parameters */
func getParamString(apiName string) string {
	HttpConfJson := ConfigLoadAndGet(apiName)
	var paramString = ""
	if apiName != config.BaseConfig.BASE_API_NAME || apiName == "" {
		paramString = HttpConfJson.Params
	}
	return paramString
}

/* load json config */
func ConfigLoadAndGet(apiName string) (json *httpConfJson) {
	c, err := goconf.New(config.BaseConfig.JSON_CONF_PATH)
	if err != nil {
		fmt.Println("Error:", err)
	}
	json = new(httpConfJson)
	json.Api_name = apiName
	c.Get("/HttpConfig", json)
	if json.AgentVersion != config.BaseConfig.AGENT_VERSION {
		err := errors.New("Running Failed!! Your agent version isn’t match the opskitchen agent version number!")
		panic(err)
	}
	return
}