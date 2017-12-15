package sms

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cocotyty/summer"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	summer.Put(&AliManager{})
}

func NewManager(appKey, appSecret string) (manager *AliManager) {
	return &AliManager{AppKey: appKey, AppSecret: appSecret}
}

type AliManager struct {
	AppKey    string `sm:"#.ali.appKey"`
	AppSecret string `sm:"#.ali.appSecret"`
}

func (m *AliManager) Handler(signName, templateCode string) Client {
	return NewAliSMSClient(signName, templateCode, m.AppKey, m.AppSecret)
}

const (
	URL     string = "https://dm.aliyuncs.com/"
	Action  string = "SingleSendSms"
	Version string = "2015-11-23"
)

type AliClient struct {
	client       *http.Client
	SignName     string
	TemplateCode string
	AppKey       string
	AppSecret    string
}

func NewAliSMSClient(signName, templateCode, appKey, appSecret string) (client *AliClient) {
	return &AliClient{
		client:       &http.Client{},
		SignName:     signName,
		TemplateCode: templateCode,
		AppKey:       appKey,
		AppSecret:    appSecret,
	}
}

type AliResult struct {
	Code string `json:"Code"`
}

var CodeMsg = map[string]string{
	"InvalidDayuStatus.Malformed":          "账户短信开通状态不正确",
	"InvalidSignName.Malformed":            "短信签名不正确或签名状态不正确",
	"InvalidTemplateCode.MalFormed":        "短信模板Code不正确或者模板状态不正确",
	"InvalidRecNum.Malformed":              "目标手机号不正确，单次发送数量不能超过100",
	"InvalidParamString.MalFormed":         "短信模板中变量不是json格式",
	"InvalidParamStringTemplate.Malformed": "短信模板中变量与模板内容不匹配",
	"InvalidSendSms":                       "触发业务流控",
	"":                                     "发送成功",
}

var CodeToErrCode = map[string]int64{
	"InvalidDayuStatus.Malformed":          7,
	"InvalidSignName.Malformed":            -1,
	"InvalidTemplateCode.MalFormed":        9,
	"InvalidRecNum.Malformed":              1,
	"InvalidParamString.MalFormed":         5,
	"InvalidParamStringTemplate.Malformed": 5,
	"InvalidSendSms":                       10,
	"":                                     0,
}

func (client *AliClient) Send(telephone string, msg map[string]string) (result *Result, err error) {

	params := make(map[string]string)
	params["Format"] = "json"
	params["Version"] = Version
	params["AccessKeyId"] = client.AppKey

	params["SignatureMethod"] = "HMAC-SHA1"
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = uuid.NewV4().String()

	params["Action"] = Action
	params["SignName"] = client.SignName
	params["TemplateCode"] = client.TemplateCode
	params["RecNum"] = telephone
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	params["ParamString"] = string(msgByte)
	body, err := client.doPost(params)
	if err != nil {
		return nil, err
	}
	aliResult := &AliResult{}
	fmt.Println(string(body))
	if err = json.Unmarshal(body, aliResult); err != nil {
		return nil, err
	}
	return &Result{ErrCode: CodeToErrCode[aliResult.Code], Msg: CodeMsg[aliResult.Code]}, nil

}

func (client *AliClient) doPost(m map[string]string) (result []byte, err error) {
	body, size := client.getRequestBody(m)
	req, _ := http.NewRequest("POST", URL, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = size
	resp, err := client.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (client *AliClient) getRequestBody(m map[string]string) (reader io.Reader, size int64) {
	v := url.Values{}
	for k := range m {
		v.Set(k, m[k])
	}
	result := "POST&%2F&" + url.QueryEscape(v.Encode())
	mac := hmac.New(sha1.New, []byte(client.AppSecret+"&"))
	mac.Write([]byte(result))
	sum := mac.Sum(nil)
	v.Set("Signature", base64.StdEncoding.EncodeToString(sum))
	return strings.NewReader(v.Encode()), int64(len(v.Encode()))
}
