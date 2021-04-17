# 简介
OKEX go版本的v5sdk。

# 项目说明

## REST调用
``` go
    // 设置您的APIKey
	apikey := APIKeyInfo{
		ApiKey:     "xxxx",
		SecKey:     "xxxx",
		PassPhrase: "xxxx",
	}

	// 模拟环境表示是否为模拟环境
	cli := NewRESTClient("https://www.okex.win", &apikey, true)
	rsp, err := cli.Get(context.Background(), "/api/v5/account/balance", nil)
	if err != nil {
		return
	}

	fmt.Println("Response:")
	fmt.Println("\thttp code: ", rsp.Code)
	fmt.Println("\t总耗时: ", rsp.TotalUsedTime)
	fmt.Println("\t请求耗时: ", rsp.ReqUsedTime)
	fmt.Println("\t返回消息: ", rsp.Body)
	fmt.Println("\terrCode: ", rsp.V5Response.Code)
	fmt.Println("\terrMsg: ", rsp.V5Response.Msg)
	fmt.Println("\tdata: ", rsp.V5Response.Data)
 ```
更多示例请查看rest/rest_test.go  

## WS订阅

### 公有频道
```go
    ep := "wss://ws.okex.com:8443/ws/v5/private?brokerId=9999"

	// 填写您自己的APIKey信息
	apikey := "xxxx"
	secretKey := "xxxxx"
	passphrase := "xxxxx"

	// 创建ws客户端
	r, err := NewWsClient(ep)
	if err != nil {
		log.Println(err)
		return
	}

	// 设置连接超时
	r.SetDailTimeout(time.Second * 2)
	err = r.Start()
	if err != nil {
		log.Println(err)
		return
	}

	var res bool

	res, _, err = r.Login(apikey, secretKey, passphrase)
	if res {
		fmt.Println("登录成功！")
	} else {
		fmt.Println("登录失败！", err)
		return
	}

	// 订阅账户频道
	var args []map[string]string
	arg := make(map[string]string)
	arg["ccy"] = "BTC"
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivAccout(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！耗时:", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
	}

	time.Sleep(100 * time.Second)
	start = time.Now()
	res, _, err = r.PrivAccout(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
	}
```
更多示例请查看ws/ws_pub_channel_test.go

### 私有频道
```go
    ep := "wss://ws.okex.com:8443/ws/v5/public?brokerId=9999"

	// 创建ws客户端
	r, err := NewWsClient(ep)
	if err != nil {
		log.Println(err)
		return
	}

	// 设置连接超时
	r.SetDailTimeout(time.Second * 2)
	err = r.Start()
	if err != nil {
		log.Println(err)
		return
	}
```
更多示例请查看ws/ws_priv_channel_test.go

# 联系方式
邮箱:caron_co@163.com  
微信:caron_co
