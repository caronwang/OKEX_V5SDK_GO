package ws

import (
	"fmt"
	"log"
	"testing"
	"time"
)

const (
	TRADE_ACCOUNT = iota
	ISOLATE_ACCOUNT
	CROSS_ACCOUNT
	CROSS_ACCOUNT_B
)

func prework_pri(t int) *WsClient {
	// 模拟环境
	ep := "wss://ws.okex.com:8443/ws/v5/private?brokerId=9999"
	var apikey, passphrase, secretKey string

	switch t {
	case TRADE_ACCOUNT:
		apikey = "x"
		passphrase = "x"
		secretKey = "x"
	case ISOLATE_ACCOUNT:
		apikey = "x"
		passphrase = "x"
		secretKey = "x"
	case CROSS_ACCOUNT:
		apikey = "x"
		secretKey = "x"
		passphrase = "x"
	case CROSS_ACCOUNT_B:
		apikey = "x"
		passphrase = "x"
		secretKey = "x"
	}

	r, err := NewWsClient(ep)
	if err != nil {
		log.Fatal(err)
	}

	err = r.Start()
	if err != nil {
		log.Fatal(err)
	}

	var res bool
	//start := time.Now()
	res, _, err = r.Login(apikey, secretKey, passphrase)
	if res {
		//usedTime := time.Since(start)
		//fmt.Println("登录成功！",usedTime.String())
	} else {
		log.Fatal("登录失败！", err)
	}
	fmt.Println(apikey, secretKey, passphrase)
	return r
}

// 账户频道 测试
func TestAccout(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var res bool
	var err error

	var args []map[string]string
	arg := make(map[string]string)
	//arg["ccy"] = "BTC"
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivAccout(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅所有成功！", usedTime.String())
	} else {
		fmt.Println("订阅所有成功！", err)
		t.Fatal("订阅所有成功！", err)
	}

	time.Sleep(100 * time.Second)
	start = time.Now()
	res, _, err = r.PrivAccout(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅所有成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅所有失败！", err)
		t.Fatal("取消订阅所有失败！", err)
	}

}

// 持仓频道 测试
func TestPositon(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var err error
	var res bool

	var args []map[string]string
	arg := make(map[string]string)
	arg["instType"] = FUTURES
	arg["uly"] = "BTC-USD"
	//arg["instId"] = "BTC-USD-210319"
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivPostion(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(60000 * time.Second)
	//等待推送

	start = time.Now()
	res, _, err = r.PrivPostion(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())

	} else {
		fmt.Println("取消订阅失败！", err)
		t.Fatal("取消订阅失败！", err)
	}

}

// 订单频道 测试
func TestBookOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var err error
	var res bool

	var args []map[string]string
	arg := make(map[string]string)
	arg["instId"] = "BTC-USDT"
	arg["instType"] = "ANY"
	//arg["instType"] = "SWAP"
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivBookOrder(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(6000 * time.Second)
	//等待推送

	start = time.Now()
	res, _, err = r.PrivBookOrder(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
		t.Fatal("取消订阅失败！", err)
	}

}

// 策略委托订单频道 测试
func TestAlgoOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var err error
	var res bool

	var args []map[string]string
	arg := make(map[string]string)
	arg["instType"] = "SPOT"
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivBookAlgoOrder(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(60 * time.Second)
	//等待推送

	start = time.Now()
	res, _, err = r.PrivBookAlgoOrder(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
		t.Fatal("取消订阅失败！", err)
	}

}

// 账户余额和持仓频道 测试
func TestPrivBalAndPos(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var err error
	var res bool

	var args []map[string]string
	arg := make(map[string]string)
	args = append(args, arg)

	start := time.Now()
	res, _, err = r.PrivBalAndPos(OP_SUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("订阅成功！", usedTime.String())
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(600 * time.Second)
	//等待推送

	start = time.Now()
	res, _, err = r.PrivBalAndPos(OP_UNSUBSCRIBE, args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
		t.Fatal("取消订阅失败！", err)
	}

}
