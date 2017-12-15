package sms

type Client interface {
	Send(telephone string, msg map[string]string) (result *Result, err error)
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	Msg     string `json:"msg"`
}
