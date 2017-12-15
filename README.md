#短信服务

> support summer

```golang

type OrderCtrl struct{
    SmsManager      *sms.AliManager `sm:"*"`
    SmsSignName     string          `sm:"#.order.smsSignName"`
    SmsTemplateCode string          `sm:"#.order.smsTemplateCode"`
    smsHandler      *sms.Client
}

func(c *OrderCtrl)Ready(){
    c.smsHandler = c.SMS.Handler(c.SmsSignName, c.SmsTemplateCode)
}

func(c *OrderCtrl)Index(){
    c.smsHandler.Send("18800001234", map[string]string{"verify":"1234"})
}
```