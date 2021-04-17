package ws

import (
	"fmt"
	"testing"
	"time"
	. "v5sdk_go/ws/wImpl"
)

func PrintDetail(d *ProcessDetail) {
	fmt.Println("[详细信息]")
	fmt.Println("请求地址:", d.EndPoint)
	fmt.Println("请求内容:", d.ReqInfo)
	fmt.Println("发送时间:", d.SendTime.Format("2006-01-02 15:04:05.000"))
	fmt.Println("响应时间:", d.RecvTime.Format("2006-01-02 15:04:05.000"))
	fmt.Println("耗时:", d.UsedTime.String())
	fmt.Printf("接受到 %v 条消息:\n", len(d.Data))
	for _, v := range d.Data {
		fmt.Printf("[%v] %v\n", v.Timestamp.Format("2006-01-02 15:04:05.000"), v.Info)
	}
}

func (r *WsClient) makeOrder(instId string, tdMode string, side string, ordType string, px string, sz string) (orderId string, err error) {

	var res bool
	var data *ProcessDetail

	param := map[string]interface{}{}
	param["instId"] = instId
	param["tdMode"] = tdMode
	param["side"] = side
	param["ordType"] = ordType
	if px != "" {
		param["px"] = px
	}
	param["sz"] = sz

	res, data, err = r.PlaceOrder("0011", param)
	if err != nil {
		return
	}
	if res && len(data.Data) == 1 {
		rsp := data.Data[0].Info.(JRPCRsp)
		if len(rsp.Data) == 1 {
			val, ok := rsp.Data[0]["ordId"]
			if !ok {
				return
			}
			orderId = val.(string)
			return
		}
	}

	return
}

/*
	单个下单
*/
func TestPlaceOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	//r := prework_pri(TRADE_ACCOUNT)
	var res bool
	var err error
	var data *ProcessDetail

	start := time.Now()
	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["tdMode"] = "cash"
	param["side"] = "buy"
	param["ordType"] = "market"
	//param["px"] = "1"
	param["sz"] = "200"

	res, data, err = r.PlaceOrder("0011", param)
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
	批量下单
*/
func TestPlaceBatchOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)
	var res bool
	var err error
	var data *ProcessDetail

	start := time.Now()
	var params []map[string]interface{}
	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["tdMode"] = "cash"
	param["side"] = "sell"
	param["ordType"] = "market"
	param["sz"] = "0.001"
	params = append(params, param)
	param = map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["tdMode"] = "cash"
	param["side"] = "buy"
	param["ordType"] = "market"
	param["sz"] = "100"
	params = append(params, param)
	res, data, err = r.BatchPlaceOrders("001", params)
	usedTime := time.Since(start)
	if err != nil {
		fmt.Println("下单失败！", err, usedTime.String())
		t.Fail()
	}
	if res {
		fmt.Println("下单成功！", usedTime.String())
		PrintDetail(data)
	} else {

		fmt.Println("下单失败！", usedTime.String())
		t.Fail()
	}

}

/*
	撤销订单
*/
func TestCancelOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)

	// 用户自定义limit限价价格
	ordId, _ := r.makeOrder("BTC-USDT", "cash", "sell", "limit", "57000", "0.01")
	if ordId == "" {
		t.Fatal()
	}

	t.Log("生成挂单：orderId=", ordId)

	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["ordId"] = ordId
	start := time.Now()
	res, _, _ := r.CancelOrder("1", param)
	if res {
		usedTime := time.Since(start)
		fmt.Println("撤单成功！", usedTime.String())
	} else {
		t.Fatal("撤单失败！")
	}
}

/*
	修改订单
*/
func TestAmendlOrder(t *testing.T) {
	r := prework_pri(CROSS_ACCOUNT)

	// 用户自定义limit限价价格
	ordId, _ := r.makeOrder("BTC-USDT", "cash", "sell", "limit", "57000", "0.01")
	if ordId == "" {
		t.Fatal()
	}

	t.Log("生成挂单：orderId=", ordId)

	param := map[string]interface{}{}
	param["instId"] = "BTC-USDT"
	param["ordId"] = ordId
	// 调整修改订单的参数
	//param["newSz"] = "0.02"
	param["newPx"] = "57001"

	start := time.Now()
	res, _, _ := r.AmendOrder("1", param)
	if res {
		usedTime := time.Since(start)
		fmt.Println("修改订单成功！", usedTime.String())
	} else {
		t.Fatal("修改订单失败！")
	}
}
