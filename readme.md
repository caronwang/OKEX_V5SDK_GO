# 简介
OKEX go版本的v5sdk，仅供学习交流使用。
(文档持续完善中)
# 项目说明

## REST调用
``` go
    // 设置您的APIKey
	apikey := APIKeyInfo{
		ApiKey:     "xxxx",
		SecKey:     "xxxx",
		PassPhrase: "xxxx",
	}

	// 第三个参数代表是否为模拟环境，更多信息查看接口说明
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

## websocket订阅

### 私有频道
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
	defer r.Stop()

	var res bool
	// 私有频道需要登录
	res, _, err = r.Login(apikey, secretKey, passphrase)
	if res {
		fmt.Println("登录成功！")
	} else {
		fmt.Println("登录失败！", err)
		return
	}

	
	var args []map[string]string
	arg := make(map[string]string)
	arg["ccy"] = "BTC"
	args = append(args, arg)

	start := time.Now()
	// 订阅账户频道
	res, _, err = r.PrivAccout(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！耗时:", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
	}

	time.Sleep(100 * time.Second)
	start = time.Now()
	// 取消订阅账户频道
	res, _, err = r.PrivAccout(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
	}
```
更多示例请查看ws/ws_priv_channel_test.go  

### 公有频道
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

	defer r.Stop()

	
	var args []map[string]string
	arg := make(map[string]string)
	arg["instType"] = FUTURES
	//arg["instType"] = OPTION
	args = append(args, arg)

	start := time.Now()

	// 订阅产品频道
	res, _, err := r.PubInstruemnts(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
	}

	time.Sleep(30 * time.Second)

	start = time.Now()

	// 取消订阅产品频道
	res, _, err = r.PubInstruemnts(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
	}
```
更多示例请查看ws/ws_pub_channel_test.go  

## websocket交易
```go
    ep := "wss://ws.okex.com:8443/ws/v5/private?brokerId=9999"

	// 填写您自己的APIKey信息
	apikey := "xxxx"
	secretKey := "xxxxx"
	passphrase := "xxxxx"

	var res bool
	var req_id string

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

	defer r.Stop()

	res, _, err = r.Login(apikey, secretKey, passphrase)
	if res {
		fmt.Println("登录成功！")
	} else {
		fmt.Println("登录失败！", err)
		return
	}

	start := time.Now()
	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["tdMode"] = "cash"
	param["side"] = "buy"
	param["ordType"] = "market"
	param["sz"] = "200"
	req_id = "00001"

	// 单个下单
	res, _, err = r.PlaceOrder(req_id, param)
	if res {
		usedTime := time.Since(start)
		fmt.Println("下单成功！", usedTime.String())
	} else {
		usedTime := time.Since(start)
		fmt.Println("下单失败！", usedTime.String(), err)
	}

```
更多示例请查看ws/ws_jrpc_test.go  

# 联系方式
邮箱:caron_co@163.com  
微信:caron_co
