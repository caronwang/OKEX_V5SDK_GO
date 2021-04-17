package ws

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
	. "v5sdk_go/config"
	"v5sdk_go/rest"
	. "v5sdk_go/utils"
	. "v5sdk_go/ws/wImpl"
	. "v5sdk_go/ws/wInterface"
)

/*
	Ping服务端保持心跳。
	timeOut:超时时间(毫秒)，如果不填默认为5000ms
*/
func (a *WsClient) Ping(timeOut ...int) (res bool, detail *ProcessDetail, err error) {
	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}
	res = true

	detail = &ProcessDetail{
		EndPoint: a.WsEndPoint,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	ctx = context.WithValue(ctx, "detail", detail)
	msg, err := a.process(ctx, EVENT_PING, nil)
	if err != nil {
		res = false
		log.Println("处理请求失败!", err)
		return
	}
	detail.Data = msg

	if len(msg) == 0 {
		res = false
		return
	}

	str := string(msg[0].Info.([]byte))
	if str != "pong" {
		res = false
		return
	}

	return
}

/*
	登录私有频道
*/
func (a *WsClient) Login(apiKey, secKey, passPhrase string, timeOut ...int) (res bool, detail *ProcessDetail, err error) {

	if apiKey == "" {
		err = errors.New("ApiKey cannot be null")
		return
	}

	if secKey == "" {
		err = errors.New("SecretKey cannot be null")
		return
	}

	if passPhrase == "" {
		err = errors.New("Passphrase cannot be null")
		return
	}

	a.WsApi = &ApiInfo{
		ApiKey:     apiKey,
		SecretKey:  secKey,
		Passphrase: passPhrase,
	}

	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}
	res = true

	timestamp := EpochTime()

	preHash := PreHashString(timestamp, rest.GET, "/users/self/verify", "")
	//fmt.Println("preHash:", preHash)
	var sign string
	if sign, err = HmacSha256Base64Signer(preHash, secKey); err != nil {
		log.Println("处理签名失败！", err)
		return
	}

	args := map[string]string{}
	args["apiKey"] = apiKey
	args["passphrase"] = passPhrase
	args["timestamp"] = timestamp
	args["sign"] = sign
	req := &ReqData{
		Op:   OP_LOGIN,
		Args: []map[string]string{args},
	}

	detail = &ProcessDetail{
		EndPoint: a.WsEndPoint,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	ctx = context.WithValue(ctx, "detail", detail)

	msg, err := a.process(ctx, EVENT_LOGIN, req)
	if err != nil {
		res = false
		log.Println("处理请求失败!", req, err)
		return
	}
	detail.Data = msg

	if len(msg) == 0 {
		res = false
		return
	}

	info, _ := msg[0].Info.(ErrData)

	if info.Code == "0" && info.Event == OP_LOGIN {
		log.Println("登录成功!")
	} else {
		log.Println("登录失败!")
		res = false
		return
	}

	return
}

/*
	等待结果响应
*/
func (a *WsClient) waitForResult(e Event, timeOut int) (data interface{}, err error) {

	if _, ok := a.regCh[e]; !ok {
		a.lock.Lock()
		a.regCh[e] = make(chan *Msg)
		a.lock.Unlock()
		//log.Println("注册", e, "事件成功")
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	ch := a.regCh[e]
	//log.Println(e, "等待响应！")
	select {
	case <-time.After(time.Duration(timeOut) * time.Millisecond):
		log.Println(e, "超时未响应！")
		err = errors.New(e.String() + "超时未响应！")
		return
	case data = <-ch:
		//log.Println(data)
	}

	return
}

/*
	发送消息到服务端
*/
func (a *WsClient) Send(ctx context.Context, op WSReqData) (err error) {
	select {
	case <-ctx.Done():
		log.Println("发生失败退出！")
		err = errors.New("发送超时退出！")
	case a.sendCh <- op.ToString():
	}

	return
}

func (a *WsClient) process(ctx context.Context, e Event, op WSReqData) (data []*Msg, err error) {
	defer func() {
		_ = recover()
	}()

	var detail *ProcessDetail
	if val := ctx.Value("detail"); val != nil {
		detail = val.(*ProcessDetail)
	} else {
		detail = &ProcessDetail{
			EndPoint: a.WsEndPoint,
		}
	}
	defer func() {
		//fmt.Println("处理完成,", e.String())
		detail.UsedTime = detail.RecvTime.Sub(detail.SendTime)
	}()

	//查看事件是否被注册
	if _, ok := a.regCh[e]; !ok {
		a.lock.Lock()
		a.regCh[e] = make(chan *Msg)
		a.lock.Unlock()
		//log.Println("注册", e, "事件成功")
	} else {
		//log.Println("事件", e, "已注册！")
		err = errors.New("事件" + e.String() + "尚未处理完毕")
		return
	}

	//预期请求响应的条数
	expectCnt := 1
	if op != nil {
		expectCnt = op.Len()
	}
	recvCnt := 0

	//等待完成通知
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(ctx context.Context) {
		defer func() {
			a.lock.Lock()
			delete(a.regCh, e)
			//log.Println("事件已注销!",e)
			a.lock.Unlock()
			wg.Done()
		}()

		a.lock.RLock()
		ch := a.regCh[e]
		a.lock.RUnlock()

		//log.Println(e, "等待响应！")
		done := false
		ok := true
		for {
			var item *Msg
			select {
			case <-ctx.Done():
				log.Println(e, "超时未响应！")
				err = errors.New(e.String() + "超时未响应！")
				return
			case item, ok = <-ch:
				if !ok {
					return
				}
				detail.RecvTime = time.Now()
				//log.Println(e, "接受到数据", item)
				data = append(data, item)
				recvCnt++
				//log.Println(data)
				if recvCnt == expectCnt {
					done = true
					break
				}
			}
			if done {
				break
			}
		}
		if ok {
			close(ch)
		}

	}(ctx)

	switch e {
	case EVENT_PING:
		msg := "ping"
		detail.ReqInfo = msg
		a.sendCh <- msg
		detail.SendTime = time.Now()
	default:
		detail.ReqInfo = op.ToString()
		err = a.Send(ctx, op)
		if err != nil {
			log.Println("发送[", e, "]消息失败！", err)
			return
		}
		detail.SendTime = time.Now()
	}

	wg.Wait()
	return
}

/*
	根据args请求参数判断请求类型
	如：{"channel": "account","ccy": "BTC"} 类型为 EVENT_BOOK_ACCOUNT
*/
func GetEventByParam(param map[string]string) (evtId Event) {
	evtId = EVENT_UNKNOWN
	channel, ok := param["channel"]
	if !ok {
		return
	}

	evtId = GetEventId(channel)
	return
}

/*
	订阅频道。
	req：请求json字符串
*/
func (a *WsClient) Subscribe(param map[string]string, timeOut ...int) (res bool, detail *ProcessDetail, err error) {
	res = true
	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}

	evtid := GetEventByParam(param)
	if evtid == EVENT_UNKNOWN {
		err = errors.New("非法的请求参数！")
		return
	}

	var args []map[string]string
	args = append(args, param)

	req := ReqData{
		Op:   OP_SUBSCRIBE,
		Args: args,
	}

	detail = &ProcessDetail{
		EndPoint: a.WsEndPoint,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	ctx = context.WithValue(ctx, "detail", detail)

	msg, err := a.process(ctx, evtid, req)
	if err != nil {
		res = false
		log.Println("处理请求失败!", req, err)
		return
	}
	detail.Data = msg

	//检查所有频道是否都更新成功
	res, err = checkResult(req, msg)
	if err != nil {
		res = false
		return
	}

	return
}

/*
	取消订阅频道。
	req：请求json字符串
*/
func (a *WsClient) UnSubscribe(param map[string]string, timeOut ...int) (res bool, detail *ProcessDetail, err error) {
	res = true
	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}

	evtid := GetEventByParam(param)
	if evtid == EVENT_UNKNOWN {
		err = errors.New("非法的请求参数！")
		return
	}

	var args []map[string]string
	args = append(args, param)

	req := ReqData{
		Op:   OP_UNSUBSCRIBE,
		Args: args,
	}

	detail = &ProcessDetail{
		EndPoint: a.WsEndPoint,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	ctx = context.WithValue(ctx, "detail", detail)
	msg, err := a.process(ctx, evtid, req)
	if err != nil {
		res = false
		log.Println("处理请求失败!", req, err)
		return
	}
	detail.Data = msg
	//检查所有频道是否都更新成功
	res, err = checkResult(req, msg)
	if err != nil {
		res = false
		return
	}

	return
}

/*
	jrpc请求
*/
func (a *WsClient) Jrpc(id, op string, params []map[string]interface{}, timeOut ...int) (res bool, detail *ProcessDetail, err error) {
	res = true
	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}

	evtid := GetEventId(op)
	if evtid == EVENT_UNKNOWN {
		err = errors.New("非法的请求参数！")
		return
	}

	req := JRPCReq{
		Id:   id,
		Op:   op,
		Args: params,
	}
	detail = &ProcessDetail{
		EndPoint: a.WsEndPoint,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	ctx = context.WithValue(ctx, "detail", detail)
	msg, err := a.process(ctx, evtid, req)
	if err != nil {
		res = false
		log.Println("处理请求失败!", req, err)
		return
	}
	detail.Data = msg

	//检查所有频道是否都更新成功
	res, err = checkResult(req, msg)
	if err != nil {
		res = false
		return
	}

	return
}

func (a *WsClient) PubChannel(evtId Event, op string, params []map[string]string, pd Period, timeOut ...int) (res bool, msg []*Msg, err error) {

	// 参数校验
	pa, err := checkParams(evtId, params, pd)
	if err != nil {
		return
	}

	res = true
	tm := 5000
	if len(timeOut) != 0 {
		tm = timeOut[0]
	}

	req := ReqData{
		Op:   op,
		Args: pa,
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Duration(tm)*time.Millisecond)
	msg, err = a.process(ctx, evtId, req)
	if err != nil {
		res = false
		log.Println("处理请求失败!", req, err)
		return
	}

	//检查所有频道是否都更新成功

	res, err = checkResult(req, msg)
	if err != nil {
		res = false
		return
	}

	return
}

// 参数校验
func checkParams(evtId Event, params []map[string]string, pd Period) (res []map[string]string, err error) {

	channel := evtId.GetChannel(pd)
	if channel == "" {
		err = errors.New("参数校验失败!未知的类型:" + evtId.String())
		return
	}
	log.Println(channel)
	if params == nil {
		tmp := make(map[string]string)
		tmp["channel"] = channel
		res = append(res, tmp)
	} else {
		//log.Println(params)
		for _, param := range params {

			tmp := make(map[string]string)
			for k, v := range param {
				tmp[k] = v
			}

			val, ok := tmp["channel"]
			if !ok {
				tmp["channel"] = channel
			} else {
				if val != channel {
					err = errors.New("参数校验失败!channel应为" + channel + val)
					return
				}
			}

			res = append(res, tmp)
		}
	}

	return
}
