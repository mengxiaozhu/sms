package sms

import (
	"errors"
	"gopkg.in/redis.v4"
	"os"
	"testing"
	"time"
)

var c *SafeClient
var tel string = os.Getenv("tel")

func init() {
	c = &SafeClient{
		Redis: redis.NewClient(&redis.Options{
			Addr:     os.Getenv("redis_addr"),
			Password: os.Getenv("redis_password"),
			PoolSize: 10,
		}),
		Prefix:          os.Getenv("redis_prefix"),
		AccessKeyID:     os.Getenv("ali_key"),
		AccessKeySecret: os.Getenv("ali_secret"),
		SignName:        os.Getenv("SignName"),
		TemplateCode:    os.Getenv("TemplateCode"),
		TemplateParam:   os.Getenv("TemplateParam"),
	}
	c.Ready()
}
func TestSafeClient_Sign(t *testing.T) {
	signMsg := "TestSafeClient_Sign"
	t.Log(signMsg, "begin")
	defer t.Log(signMsg, "\n")
	sign, err := c.Sign(signMsg, 10, 2*time.Second)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	t.Log("OK", sign)
	var vCode string
	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
	if err != nil {
		t.Error(err)
		t.Error(errors.New("签名有效却没有通过"))
		t.FailNow()
		return
	}
	t.Log("OK", vCode)

	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign+"1")
	if err == nil {
		t.Error(errors.New("签名无效却通过了"))
		t.FailNow()
		return
	}
	t.Log("OK", "无效签名")

	time.Sleep(3 * time.Second)
	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
	if err == nil {
		t.Error(errors.New("签名超时却通过了"))
		t.FailNow()
		return
	}
	t.Log("OK", "无效签名")

}

func TestSafeClient_SignLimit(t *testing.T) {
	signMsg := "TestSafeClient_SignLimit"
	t.Log(signMsg, "begin")
	defer t.Log(signMsg, "\n")
	sign, err := c.Sign(signMsg, 10, 2*time.Second)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	t.Log("OK", sign)
	var vCode string
	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
	if err != nil {
		t.Error(err)
		t.Error(errors.New("签名有效却没有通过"))
		t.FailNow()
		return
	}
	t.Log("OK", vCode)
	for i := 0; i < 10; i++ {
		vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
		if err == SignLimited {

			return
		}
	}
	t.Error(errors.New("签名超过了限制却通过了"))
	t.FailNow()

}

func TestSafeClient_VCode(t *testing.T) {
	signMsg := "TestSafeClient_VCode"
	t.Log(signMsg, "begin")
	defer t.Log(signMsg, "\n")
	sign, err := c.Sign(signMsg, 10, 2*time.Second)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	t.Log("OK", sign)
	var vCode string
	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
	if err != nil {
		t.Error(err)
		t.Error(errors.New("签名有效却没有通过"))
		t.FailNow()
		return
	}
	t.Log("OK", vCode)
	success, err := c.VerifyVCode(tel, vCode, sign)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	if !success {
		t.Error("vCode未效验通过")
		t.FailNow()
	}

	t.Log("OK", success)

	return

}

func TestSafeClient_VCodeFalse(t *testing.T) {
	signMsg := "TestSafeClient_VCodeFalse"
	t.Log(signMsg, "begin")
	defer t.Log(signMsg, "\n")
	sign, err := c.Sign(signMsg, 10, 2*time.Second)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	t.Log("OK", sign)
	var vCode string
	vCode, err = c.SendVCode(tel, 6, 20*time.Minute, sign)
	if err != nil {
		t.Error(err)
		t.Error(errors.New("签名有效却没有通过"))
		t.FailNow()
		return
	}
	t.Log("OK", vCode)

	success, err := c.VerifyVCode(tel, vCode+"1", sign)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	if success {
		t.Error("错误的vCode验证码效验通过")
		t.FailNow()
	}
	t.Log("OK", success)

	success, err = c.VerifyVCode(tel, vCode, sign)
	if err == redis.Nil {
		t.Log("OK", err)
		return
	}
	if success {
		t.Error("第二次vCode尝试通过")
		t.FailNow()
	}
	t.Log("OK", success)

	return

}
