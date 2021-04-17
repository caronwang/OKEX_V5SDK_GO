package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	. "v5sdk_go/utils"
)

type RESTAPI struct {
	EndPoint string `json:"endPoint"`
	// GET/POST
	Method     string                 `json:"method"`
	Uri        string                 `json:"uri"`
	Param      map[string]interface{} `json:"param"`
	Timeout    time.Duration
	ApiKeyInfo *APIKeyInfo
	isSimulate bool
}

type APIKeyInfo struct {
	ApiKey     string
	PassPhrase string
	SecKey     string
	UserId     string
}

type RESTAPIResult struct {
	Url    string `json:"url"`
	Param  string `json:"param"`
	Header string `json:"header"`
	Code   int    `json:"code"`
	// 原始返回信息
	Body string `json:"body"`
	// okexV5返回的数据
	V5Response    Okexv5APIResponse `json:"v5Response"`
	ReqUsedTime   time.Duration     `json:"reqUsedTime"`
	TotalUsedTime time.Duration     `json:"totalUsedTime"`
}

type Okexv5APIResponse struct {
	Code string                   `json:"code"`
	Msg  string                   `json:"msg"`
	Data []map[string]interface{} `json:"data"`
}

/*
	endPoint:请求地址
	apiKey
	isSimulate: 是否为模拟环境
*/
func NewRESTClient(endPoint string, apiKey *APIKeyInfo, isSimulate bool) *RESTAPI {

	res := &RESTAPI{
		EndPoint:   endPoint,
		ApiKeyInfo: apiKey,
		isSimulate: isSimulate,
		Timeout:    5 * time.Second,
	}
	return res
}

func NewRESTAPI(ep, method, uri string, param *map[string]interface{}) *RESTAPI {
	//TODO:参数校验
	reqParam := make(map[string]interface{})

	if param != nil {
		reqParam = *param
	}
	res := &RESTAPI{
		EndPoint: ep,
		Method:   method,
		Uri:      uri,
		Param:    reqParam,
		Timeout:  150 * time.Second,
	}
	return res
}

func (this *RESTAPI) SetSimulate(b bool) *RESTAPI {
	this.isSimulate = b
	return this
}

func (this *RESTAPI) SetAPIKey(apiKey, secKey, passPhrase string) *RESTAPI {
	if this.ApiKeyInfo == nil {
		this.ApiKeyInfo = &APIKeyInfo{
			ApiKey:     apiKey,
			PassPhrase: passPhrase,
			SecKey:     secKey,
		}
	} else {
		this.ApiKeyInfo.ApiKey = apiKey
		this.ApiKeyInfo.PassPhrase = passPhrase
		this.ApiKeyInfo.SecKey = secKey
	}
	return this
}

func (this *RESTAPI) SetUserId(userId string) *RESTAPI {
	if this.ApiKeyInfo == nil {
		fmt.Println("ApiKey为空")
		return this
	}

	this.ApiKeyInfo.UserId = userId
	return this
}

func (this *RESTAPI) SetTimeOut(timeout time.Duration) *RESTAPI {
	this.Timeout = timeout
	return this
}

// GET请求
func (this *RESTAPI) Get(ctx context.Context, uri string, param *map[string]interface{}) (res *RESTAPIResult, err error) {
	this.Method = GET
	this.Uri = uri

	reqParam := make(map[string]interface{})

	if param != nil {
		reqParam = *param
	}
	this.Param = reqParam
	return this.Run(ctx)
}

// POST请求
func (this *RESTAPI) Post(ctx context.Context, uri string, param *map[string]interface{}) (res *RESTAPIResult, err error) {
	this.Method = POST
	this.Uri = uri

	reqParam := make(map[string]interface{})

	if param != nil {
		reqParam = *param
	}
	this.Param = reqParam

	return this.Run(ctx)
}

func (this *RESTAPI) Run(ctx context.Context) (res *RESTAPIResult, err error) {

	if this.ApiKeyInfo == nil {
		err = errors.New("APIKey不可为空")
		return
	}

	procStart := time.Now()

	defer func() {
		if res != nil {
			res.TotalUsedTime = time.Since(procStart)
		}
	}()

	client := &http.Client{
		Timeout: this.Timeout,
	}

	uri, body, err := this.GenReqInfo()
	if err != nil {
		return
	}

	url := this.EndPoint + uri
	bodyBuf := new(bytes.Buffer)
	bodyBuf.ReadFrom(strings.NewReader(body))

	req, err := http.NewRequest(this.Method, url, bodyBuf)
	if err != nil {
		return
	}

	res = &RESTAPIResult{
		Url:   url,
		Param: body,
	}

	// Sign and set request headers
	timestamp := IsoTime()
	preHash := PreHashString(timestamp, this.Method, uri, body)
	//log.Println("preHash:", preHash)
	sign, err := HmacSha256Base64Signer(preHash, this.ApiKeyInfo.SecKey)
	if err != nil {
		return
	}
	//log.Println("sign:", sign)
	headStr := this.SetHeaders(req, timestamp, sign)
	res.Header = headStr

	this.PrintRequest(req, body, preHash)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败！", err)
		return
	}
	defer resp.Body.Close()

	res.ReqUsedTime = time.Since(procStart)

	resBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("获取请求结果失败！", err)
		return
	}

	res.Body = string(resBuff)
	res.Code = resp.StatusCode

	// 解析结果
	var v5rsp Okexv5APIResponse
	err = json.Unmarshal(resBuff, &v5rsp)
	if err != nil {
		fmt.Println("解析v5返回失败！", err)
		return
	}

	res.V5Response = v5rsp

	return
}

/*
	生成请求对应的参数
*/
func (this *RESTAPI) GenReqInfo() (uri string, body string, err error) {
	uri = this.Uri

	switch this.Method {
	case GET:
		getParam := []string{}

		if len(this.Param) == 0 {
			return
		}

		for k, v := range this.Param {
			getParam = append(getParam, fmt.Sprintf("%v=%v", k, v))
		}
		uri = uri + "?" + strings.Join(getParam, "&")

	case POST:

		var rawBody []byte
		rawBody, err = json.Marshal(this.Param)
		if err != nil {
			return
		}
		body = string(rawBody)
	default:
		err = errors.New("request type unknown!")
		return
	}

	return
}

/*
   Set http request headers:
   Accept: application/json
   Content-Type: application/json; charset=UTF-8  (default)
   Cookie: locale=en_US        (English)
   OK-ACCESS-KEY: (Your setting)
   OK-ACCESS-SIGN: (Use your setting, auto sign and add)
   OK-ACCESS-TIMESTAMP: (Auto add)
   OK-ACCESS-PASSPHRASE: Your setting
*/
func (this *RESTAPI) SetHeaders(request *http.Request, timestamp string, sign string) (header string) {

	request.Header.Add(ACCEPT, APPLICATION_JSON)
	header += ACCEPT + ":" + APPLICATION_JSON + "\n"

	request.Header.Add(CONTENT_TYPE, APPLICATION_JSON_UTF8)
	header += CONTENT_TYPE + ":" + APPLICATION_JSON_UTF8 + "\n"

	request.Header.Add(COOKIE, LOCALE+ENGLISH)
	header += COOKIE + ":" + LOCALE + ENGLISH + "\n"

	request.Header.Add(OK_ACCESS_KEY, this.ApiKeyInfo.ApiKey)
	header += OK_ACCESS_KEY + ":" + this.ApiKeyInfo.ApiKey + "\n"

	request.Header.Add(OK_ACCESS_SIGN, sign)
	header += OK_ACCESS_SIGN + ":" + sign + "\n"

	request.Header.Add(OK_ACCESS_TIMESTAMP, timestamp)
	header += OK_ACCESS_TIMESTAMP + ":" + timestamp + "\n"

	request.Header.Add(OK_ACCESS_PASSPHRASE, this.ApiKeyInfo.PassPhrase)
	header += OK_ACCESS_PASSPHRASE + ":" + this.ApiKeyInfo.PassPhrase + "\n"

	//模拟盘交易标记
	if this.isSimulate {
		request.Header.Add(X_SIMULATE_TRADING, "1")
		header += X_SIMULATE_TRADING + ":1" + "\n"
	}
	return
}

/*
	打印header信息
*/
func (this *RESTAPI) PrintRequest(request *http.Request, body string, preHash string) {
	if this.ApiKeyInfo.SecKey != "" {
		fmt.Println("  Secret-Key: " + this.ApiKeyInfo.SecKey)
	}
	fmt.Println("  Request(" + IsoTime() + "):")
	fmt.Println("\tUrl: " + request.URL.String())
	fmt.Println("\tMethod: " + strings.ToUpper(request.Method))
	if len(request.Header) > 0 {
		fmt.Println("\tHeaders: ")
		for k, v := range request.Header {
			if strings.Contains(k, "Ok-") {
				k = strings.ToUpper(k)
			}
			fmt.Println("\t\t" + k + ": " + v[0])
		}
	}
	fmt.Println("\tBody: " + body)
	if preHash != "" {
		fmt.Println("  PreHash: " + preHash)
	}
}
