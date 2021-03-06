package sms

import aliSMS "github.com/denverdino/aliyungo/sms"
import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"gopkg.in/redis.v4"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"log"
)

const (
	VCodeLength4 = 4
	VCodeLength6 = 6
)

var (
	SignLimited = errors.New("sms: sign limited.")
)

type SafeClient struct {
	Redis           *redis.Client     `sm:"(.redis)"`
	Prefix          string            `sm:"(.prefix)"`
	AccessKeyID     string            `sm:"(.ali).id"`
	AccessKeySecret string            `sm:"(.ali).secret"`
	SignName        string            `sm:"(.opts).SignName"`
	TemplateCode    string            `sm:"(.opts).TemplateCode"`
	TemplateParam   string            `sm:"(.opts).TemplateParam"`
	SmsUpExtendCode string            `sm:"(.opts).SmsUpExtendCode"`
	OutId           string            `sm:"(.opts).OutId"`
	dySmsClient     *aliSMS.DYSmsClient
	r               *rand.Rand
	moduleName string
}

func (c *SafeClient) Ready() {
	c.dySmsClient = aliSMS.NewDYSmsClient(c.AccessKeyID, c.AccessKeySecret)
	//c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
}
func (c *SafeClient)GetCommonClient()*aliSMS.DYSmsClient{
	return c.dySmsClient
}

// 增加一个签名，限定了改签名的调用次数及重置时间
func (c *SafeClient) Sign(msg string, limit int, ttl time.Duration) (sign string, err error) {
	h := sha1.New()
	h.Write([]byte(msg))
	s := hex.EncodeToString(h.Sum(nil))
	k := c.Prefix + "/sign/" + s
	err = c.Redis.Get(k).Err()
	if err == nil {
		return s, nil
	}
	if err != redis.Nil {
		return
	}
	err = c.Redis.Set(k, limit, ttl).Err()
	if err != nil {
		return
	}
	return s, nil
}
func (c *SafeClient) dec(sign string) (success bool, err error) {
	k := c.Prefix + "/sign/" + sign
	limit, err := c.Redis.Get(k).Result()
	if err != nil {
		return
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		return
	}
	if l <= 0 {
		return false, nil
	}
	err = c.Redis.Decr(k).Err()
	if err != nil {
		return
	}
	return true, nil
}

func (c *SafeClient) vCode(length int) string {
	i := rand.Int()
	return fmt.Sprintf("%0"+strconv.Itoa(length)+"."+strconv.Itoa(length)+"s", strconv.Itoa(i))
}

func (c *SafeClient) SendVCode(tel string, length int, ttl time.Duration, sign string) (vCode string, err error) {
	ok, err := c.dec(sign)
	if err != nil {
		return
	}
	if !ok {
		err = SignLimited
		return
	}
	k := c.Prefix + "/vCode/" + tel
	v := c.vCode(length)
	err = c.Redis.Set(k, v, ttl).Err()
	if err != nil {
		return
	}
	log.Printf("sendVCode phone: %s vcode:%s \n",tel, v)
	resp, err := c.dySmsClient.SendSms(&aliSMS.SendSmsArgs{
		PhoneNumbers:    tel,
		SignName:        c.SignName,
		TemplateCode:    c.TemplateCode,
		TemplateParam:   strings.Replace(c.TemplateParam, `${vCode}`, v, -1),
		SmsUpExtendCode: c.SmsUpExtendCode,
		OutId:           c.OutId,
	})
	if err != nil {
		return
	}

	if resp.Code != "OK" {
		err = errors.New(resp.RequestId + ";" + resp.Code + ";" + resp.Message + ";" + resp.BizId)
		return
	}
	vCode = v
	return vCode, err
}
func (c *SafeClient) VerifyVCode(tel string, vCode string, sign string) (success bool, err error) {
	ok, err := c.dec(sign)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, SignLimited
	}
	k := c.Prefix + "/vCode/" + tel
	r, err := c.Redis.Get(k).Result()
	if err != nil {
		return false, err
	}
	if r == vCode {
		success = true
		c.Redis.Del(k)
	}
	return success, nil
}

func (c *SafeClient) SetModuleName(name string){
	c.moduleName = name
}
