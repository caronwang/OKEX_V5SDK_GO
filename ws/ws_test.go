package ws

import (
	"fmt"
	"log"
	"testing"
	"time"
	. "v5sdk_go/ws/wImpl"

	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

}

func TestPing(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)

	res, _, _ := r.Ping()
	assert.True(t, res, true)
}

func TestWsClient_SubscribeAndUnSubscribe(t *testing.T) {
	r := prework()
	var err error
	var res bool

	param := map[string]string{}
	param["channel"] = "opt-summary"
	param["uly"] = "BTC-USD"

	start := time.Now()
	res, _, err = r.Subscribe(param)
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
	res, _, err = r.UnSubscribe(param)
	if res {
		usedTime := time.Since(start)
		fmt.Println("取消订阅成功！", usedTime.String())
	} else {
		fmt.Println("取消订阅失败！", err)
		t.Fatal("取消订阅失败！", err)
	}

}

func TestWsClient_SubscribeAndUnSubscribe_priv(t *testing.T) {
	r := prework_pri(ISOLATE_ACCOUNT)
	var err error
	var res bool

	var params []map[string]string
	params = append(params, map[string]string{"channel": "orders", "instType": SPOT, "instId": "BTC-USDT"})
	//一个失败的订阅用例
	params = append(params, map[string]string{"channel": "positions", "instType": "any"})

	for _, v := range params {
		start := time.Now()
		var data *ProcessDetail
		res, data, err = r.Subscribe(v)
		if res {
			usedTime := time.Since(start)
			fmt.Println("订阅成功！", usedTime.String())
			PrintDetail(data)
		} else {
			fmt.Println("订阅失败！", err)
			//return
		}
		time.Sleep(60 * time.Second)
		//等待推送

		start = time.Now()
		res, _, err = r.UnSubscribe(v)
		if res {
			usedTime := time.Since(start)
			fmt.Println("取消订阅成功！", usedTime.String())
		} else {
			fmt.Println("取消订阅失败！", err)
		}

	}

}

func TestWsClient_Jrpc(t *testing.T) {
	//r := prework_pri(ISOLATE_ACCOUNT)
	r := prework_pri(CROSS_ACCOUNT)
	var res bool
	var err error
	var data *ProcessDetail

	start := time.Now()
	var args []map[string]interface{}

	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["clOrdId"] = "SIM0dcopy16069997808063455"
	param["tdMode"] = "cross"
	param["side"] = "sell"
	param["ordType"] = "limit"
	param["px"] = "19333.3"
	param["sz"] = "0.18605445"

	param1 := map[string]interface{}{}
	param1["instId"] = "BTC-USDT"
	param1["clOrdId"] = "SIM0dcopy16069997808063456"
	param1["tdMode"] = "cross"
	param1["side"] = "sell"
	param1["ordType"] = "limit"
	param1["px"] = "19334.2"
	param1["sz"] = "0.03508913"

	param2 := map[string]interface{}{}
	param2["instId"] = "BTC-USDT"
	param2["clOrdId"] = "SIM0dcopy16069997808063457"
	param2["tdMode"] = "cross"
	param2["side"] = "sell"
	param2["ordType"] = "limit"
	param2["px"] = "19334.8"
	param2["sz"] = "0.03658186"

	param3 := map[string]interface{}{}
	param3["instId"] = "BTC-USDT"
	param3["clOrdId"] = "SIM0dcopy16069997808063458"
	param3["tdMode"] = "cross"
	param3["side"] = "sell"
	param3["ordType"] = "limit"
	param3["px"] = "19334.9"
	param3["sz"] = "0.5"

	param4 := map[string]interface{}{}
	param4["instId"] = "BTC-USDT"
	param4["clOrdId"] = "SIM0dcopy16069997808063459"
	param4["tdMode"] = "cross"
	param4["side"] = "sell"
	param4["ordType"] = "limit"
	param4["px"] = "19335.2"
	param4["sz"] = "0.3"

	param5 := map[string]interface{}{}
	param5["instId"] = "BTC-USDT"
	param5["clOrdId"] = "SIM0dcopy16069997808063460"
	param5["tdMode"] = "cross"
	param5["side"] = "sell"
	param5["ordType"] = "limit"
	param5["px"] = "19335.9"
	param5["sz"] = "0.051"

	param6 := map[string]interface{}{}
	param6["instId"] = "BTC-USDT"
	param6["clOrdId"] = "SIM0dcopy16069997808063461"
	param6["tdMode"] = "cross"
	param6["side"] = "sell"
	param6["ordType"] = "limit"
	param6["px"] = "19336.4"
	param6["sz"] = "1"

	param7 := map[string]interface{}{}
	param7["instId"] = "BTC-USDT"
	param7["clOrdId"] = "SIM0dcopy16069997808063462"
	param7["tdMode"] = "cross"
	param7["side"] = "sell"
	param7["ordType"] = "limit"
	param7["px"] = "19336.8"
	param7["sz"] = "0.475"

	param8 := map[string]interface{}{}
	param8["instId"] = "BTC-USDT"
	param8["clOrdId"] = "SIM0dcopy16069997808063463"
	param8["tdMode"] = "cross"
	param8["side"] = "sell"
	param8["ordType"] = "limit"
	param8["px"] = "19337.3"
	param8["sz"] = "0.21299357"

	param9 := map[string]interface{}{}
	param9["instId"] = "BTC-USDT"
	param9["clOrdId"] = "SIM0dcopy16069997808063464"
	param9["tdMode"] = "cross"
	param9["side"] = "sell"
	param9["ordType"] = "limit"
	param9["px"] = "19337.5"
	param9["sz"] = "0.5"

	args = append(args, param)
	args = append(args, param1)
	args = append(args, param2)
	args = append(args, param3)
	args = append(args, param4)
	args = append(args, param5)
	args = append(args, param6)
	args = append(args, param7)
	args = append(args, param8)
	args = append(args, param9)

	res, data, err = r.Jrpc("okexv5wsapi001", "order", args)
	if res {
		usedTime := time.Since(start)
		fmt.Println("下单成功！", usedTime.String())
		PrintDetail(data)
	} else {
		usedTime := time.Since(start)
		fmt.Println("下单失败！", usedTime.String(), err)
	}
}

/*
	测试 添加全局消息回调函数
*/
func TestAddMessageHook(t *testing.T) {

	r := prework_pri(CROSS_ACCOUNT)

	r.AddMessageHook(func(msg *Msg) error {
		// 添加你的方法
		fmt.Println("这是自定义MessageHook")
		fmt.Println("当前数据是", msg)
		return nil
	})

	select {}
}

/*
	普通推送数据回调函数
*/
func TestAddBookedDataHook(t *testing.T) {
	var r *WsClient

	/*订阅私有频道*/
	{
		r = prework_pri(CROSS_ACCOUNT)
		var res bool
		var err error

		r.AddBookMsgHook(func(ts time.Time, data MsgData) error {
			// 添加你的方法
			fmt.Println("这是自定义AddBookMsgHook")
			fmt.Println("当前数据是", data)
			return nil
		})

		param := map[string]string{}
		param["channel"] = "account"
		param["ccy"] = "BTC"

		res, _, err = r.Subscribe(param)
		if res {
			fmt.Println("订阅成功！")
		} else {
			fmt.Println("订阅失败！", err)
			t.Fatal("订阅失败！", err)
			//return
		}

		time.Sleep(100 * time.Second)
	}

	//订阅公共频道
	{
		r = prework()
		var res bool
		var err error

		r.AddBookMsgHook(func(ts time.Time, data MsgData) error {
			// 添加你的方法
			fmt.Println("这是自定义AddBookMsgHook")
			fmt.Println("当前数据是", data)
			return nil
		})

		param := map[string]string{}
		param["channel"] = "instruments"
		param["instType"] = "FUTURES"

		res, _, err = r.Subscribe(param)
		if res {
			fmt.Println("订阅成功！")
		} else {
			fmt.Println("订阅失败！", err)
			t.Fatal("订阅失败！", err)
			//return
		}

		select {}
	}

}

func TestGetInfoFromErrMsg(t *testing.T) {
	a := assert.New(t)
	buf := `
"channel:index-tickers,instId:BTC-USDT1 doesn't exist"
    `
	ch := GetInfoFromErrMsg(buf)
	//t.Log(ch)
	a.Equal("index-tickers", ch)

	//assert.True(t,ch == "index-tickers")
}

/*

 */
func TestParseMessage(t *testing.T) {
	r := prework()
	var evt Event
	msg := `{"event":"error","msg":"Contract does not exist.","code":"51001"}`

	evt, _, _ = r.parseMessage([]byte(msg))
	assert.True(t, EVENT_ERROR == evt)

	msg = `{"event":"error","msg":"channel:positions,ccy:BTC doesn't exist","code":"60018"}`
	evt, _, _ = r.parseMessage([]byte(msg))
	assert.True(t, EVENT_BOOK_POSTION == evt)
}

/*
	原始方式 深度订阅 测试
*/
func TestSubscribeTBT(t *testing.T) {
	r := prework()
	var res bool
	var err error

	// 添加你的方法
	r.AddDepthHook(func(ts time.Time, data DepthData) error {
		//fmt.Println("这是自定义AddBookMsgHook")
		fmt.Println("当前数据是:", data)
		return nil
	})

	param := map[string]string{}
	param["channel"] = "books-l2-tbt"
	//param["channel"] = "books"
	param["instId"] = "BTC-USD-SWAP"

	res, _, err = r.Subscribe(param)
	if res {
		fmt.Println("订阅成功！")
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(60 * time.Second)
}

/*

 */
func TestSubscribeBalAndPos(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var res bool
	var err error

	param := map[string]string{}

	// 产品信息
	param["channel"] = "balance_and_position"

	res, _, err = r.Subscribe(param)
	if res {
		fmt.Println("订阅成功！")
	} else {
		fmt.Println("订阅失败！", err)
		t.Fatal("订阅失败！", err)
		//return
	}

	time.Sleep(60 * time.Second)
}
