package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
	. "v5sdk_go/config"
	. "v5sdk_go/utils"
	. "v5sdk_go/ws/wImpl"

	"github.com/gorilla/websocket"
)

// 全局回调函数
type ReceivedDataCallback func(*Msg) error

// 普通订阅推送数据回调函数
type ReceivedMsgDataCallback func(time.Time, MsgData) error

// 深度订阅推送数据回调函数
type ReceivedDepthDataCallback func(time.Time, DepthData) error

// websocket
type WsClient struct {
	WsEndPoint string
	WsApi      *ApiInfo
	conn       *websocket.Conn
	sendCh     chan string //发消息队列
	resCh      chan *Msg   //收消息队列

	errCh chan *Msg
	regCh map[Event]chan *Msg //请求响应队列

	quitCh chan struct{}
	lock   sync.RWMutex

	onMessageHook ReceivedDataCallback      //全局消息回调函数
	onBookMsgHook ReceivedMsgDataCallback   //普通订阅消息回调函数
	onDepthHook   ReceivedDepthDataCallback //深度订阅消息回调函数
	OnErrorHook   ReceivedDataCallback      //错误处理回调函数

	// 记录深度信息
	DepthDataList map[string]DepthDetail
	autoDepthMgr  bool // 深度数据管理（checksum等）
	DepthDataLock sync.RWMutex

	isStarted   bool //防止重复启动和关闭
	dailTimeout time.Duration
}

/*
	服务端响应详细信息
	Timestamp: 接受到消息的时间
	Info: 接受到的消息字符串
*/
type Msg struct {
	Timestamp time.Time   `json:"timestamp"`
	Info      interface{} `json:"info"`
}

func (this *Msg) Print() {
	fmt.Println("【消息时间】", this.Timestamp.Format("2006-01-02 15:04:05.000"))
	str, _ := json.Marshal(this.Info)
	fmt.Println("【消息内容】", string(str))
}

/*
	订阅结果封装后的消息结构体
*/
type ProcessDetail struct {
	EndPoint string        `json:"endPoint"`
	ReqInfo  string        `json:"ReqInfo"`  //订阅请求
	SendTime time.Time     `json:"sendTime"` //发送订阅请求的时间
	RecvTime time.Time     `json:"recvTime"` //接受到订阅结果的时间
	UsedTime time.Duration `json:"UsedTime"` //耗时
	Data     []*Msg        `json:"data"`     //订阅结果数据
}

func (p *ProcessDetail) String() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// 创建ws对象
func NewWsClient(ep string) (r *WsClient, err error) {
	if ep == "" {
		err = errors.New("websocket endpoint cannot be null")
		return
	}

	r = &WsClient{
		WsEndPoint: ep,
		sendCh:     make(chan string),
		resCh:      make(chan *Msg),
		errCh:      make(chan *Msg),
		regCh:      make(map[Event]chan *Msg),
		//cbs:        make(map[Event]ReceivedDataCallback),
		quitCh:        make(chan struct{}),
		DepthDataList: make(map[string]DepthDetail),
		dailTimeout:   time.Second * 5,
		// 自动深度校验默认开启
		autoDepthMgr: true,
	}

	return
}

/*
	新增记录深度信息
*/
func (a *WsClient) addDepthDataList(key string, dd DepthDetail) error {
	a.DepthDataLock.Lock()
	defer a.DepthDataLock.Unlock()
	a.DepthDataList[key] = dd
	return nil
}

/*
	更新记录深度信息（如果没有记录不会更新成功）
*/
func (a *WsClient) updateDepthDataList(key string, dd DepthDetail) error {
	a.DepthDataLock.Lock()
	defer a.DepthDataLock.Unlock()
	if _, ok := a.DepthDataList[key]; !ok {
		return errors.New("更新失败！未发现记录" + key)
	}

	a.DepthDataList[key] = dd
	return nil
}

/*
	删除记录深度信息
*/
func (a *WsClient) deleteDepthDataList(key string) error {
	a.DepthDataLock.Lock()
	defer a.DepthDataLock.Unlock()
	delete(a.DepthDataList, key)
	return nil
}

/*
	设置是否自动深度管理，开启 true，关闭 false
*/
func (a *WsClient) EnableAutoDepthMgr(b bool) error {
	a.DepthDataLock.Lock()
	defer a.DepthDataLock.Unlock()

	if len(a.DepthDataList) != 0 {
		err := errors.New("当前有深度数据处于订阅中")
		return err
	}

	a.autoDepthMgr = b
	return nil
}

/*
	获取当前的深度快照信息(合并后的)
*/
func (a *WsClient) GetSnapshotByChannel(data DepthData) (snapshot *DepthDetail, err error) {
	key, err := json.Marshal(data.Arg)
	if err != nil {
		return
	}
	a.DepthDataLock.Lock()
	defer a.DepthDataLock.Unlock()
	val, ok := a.DepthDataList[string(key)]
	if !ok {
		return
	}
	snapshot = new(DepthDetail)
	raw, err := json.Marshal(val)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &snapshot)
	if err != nil {
		return
	}
	return
}

// 设置dial超时时间
func (a *WsClient) SetDailTimeout(tm time.Duration) {
	a.dailTimeout = tm
}

// 非阻塞启动
func (a *WsClient) Start() error {
	a.lock.RLock()
	if a.isStarted {
		a.lock.RUnlock()
		fmt.Println("ws已经启动")
		return nil
	} else {
		a.lock.RUnlock()
		a.lock.Lock()
		defer a.lock.Unlock()
		// 增加超时处理
		done := make(chan struct{})
		ctx, cancel := context.WithTimeout(context.Background(), a.dailTimeout)
		defer cancel()
		go func(ctx context.Context) {
			defer func() {
				close(done)
			}()
			var c *websocket.Conn
			c, _, err := websocket.DefaultDialer.Dial(a.WsEndPoint, nil)
			if err != nil {
				err = errors.New("dial error:" + err.Error())
				return
			}
			a.conn = c

		}(ctx)
		select {
		case <-ctx.Done():
			err := errors.New("连接超时退出！")
			return err
		case <-done:

		}

		go a.receive()
		go a.work()
		a.isStarted = true
		log.Println("客户端已启动!", a.WsEndPoint)
		return nil
	}
}

// 客户端退出消息channel
func (a *WsClient) IsQuit() <-chan struct{} {
	return a.quitCh
}

func (a *WsClient) work() {
	defer func() {
		a.Stop()
		err := recover()
		if err != nil {
			log.Printf("work End. Recover msg: %+v", a)
			debug.PrintStack()
		}

	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C: // 保持心跳
			// go a.Ping(1000)
			go func() {
				_, _, err := a.Ping(1000)
				if err != nil {
					fmt.Println("心跳检测失败！", err)
					a.Stop()
					return
				}

			}()

		case <-a.quitCh: // 保持心跳
			return
		case data, ok := <-a.resCh: //接收到服务端发来的消息
			if !ok {
				return
			}
			//log.Println("接收到来自resCh的消息:", data)
			if a.onMessageHook != nil {
				err := a.onMessageHook(data)
				if err != nil {
					log.Println("执行onMessageHook函数错误！", err)
				}
			}
		case errMsg, ok := <-a.errCh: //错误处理
			if !ok {
				return
			}
			if a.OnErrorHook != nil {
				err := a.OnErrorHook(errMsg)
				if err != nil {
					log.Println("执行OnErrorHook函数错误！", err)
				}
			}
		case req, ok := <-a.sendCh: //从发送队列中取出数据发送到服务端
			if !ok {
				return
			}
			//log.Println("接收到来自req的消息:", req)
			err := a.conn.WriteMessage(websocket.TextMessage, []byte(req))
			if err != nil {
				log.Printf("发送请求失败: %s\n", err)
				return
			}
			log.Printf("[发送请求] %v\n", req)
		}
	}

}

/*
	处理接受到的消息
*/
func (a *WsClient) receive() {
	defer func() {
		a.Stop()
		err := recover()
		if err != nil {
			log.Printf("Receive End. Recover msg: %+v", a)
			debug.PrintStack()
		}

	}()

	for {
		messageType, message, err := a.conn.ReadMessage()
		if err != nil {
			if a.isStarted {
				log.Println("receive message error!" + err.Error())
			}

			break
		}

		txtMsg := message
		switch messageType {
		case websocket.TextMessage:
		case websocket.BinaryMessage:
			txtMsg, err = GzipDecode(message)
			if err != nil {
				log.Println("解压失败！")
				continue
			}
		}

		log.Println("[收到消息]", string(txtMsg))

		//发送结果到默认消息处理通道

		timestamp := time.Now()
		msg := &Msg{Timestamp: timestamp, Info: string(txtMsg)}

		a.resCh <- msg

		evt, data, err := a.parseMessage(txtMsg)
		if err != nil {
			log.Println("解析消息失败！", err)
			continue
		}

		//log.Println("解析消息成功!消息类型 =", evt)

		a.lock.RLock()
		ch, ok := a.regCh[evt]
		a.lock.RUnlock()
		if !ok {
			//只有推送消息才会主动创建通道和消费队列
			if evt == EVENT_BOOKED_DATA || evt == EVENT_DEPTH_DATA {
				//log.Println("channel不存在！event:", evt)
				//a.lock.RUnlock()
				a.lock.Lock()
				a.regCh[evt] = make(chan *Msg)
				ch = a.regCh[evt]
				a.lock.Unlock()

				//log.Println("创建", evt, "通道")

				// 创建消费队列
				go func(evt Event) {
					//log.Println("创建goroutine  evt:", evt)

					for msg := range a.regCh[evt] {
						//log.Println(msg)
						// msg.Print()
						switch evt {
						// 处理普通推送数据
						case EVENT_BOOKED_DATA:
							fn := a.onBookMsgHook
							if fn != nil {
								err = fn(msg.Timestamp, msg.Info.(MsgData))
								if err != nil {
									log.Println("订阅数据回调函数执行失败！", err)
								}
								//log.Println("函数执行成功！", err)
							}
						// 处理深度推送数据
						case EVENT_DEPTH_DATA:
							fn := a.onDepthHook

							depData := msg.Info.(DepthData)

							// 开启深度数据管理功能的，会合并深度数据
							if a.autoDepthMgr {
								a.MergeDepth(depData)
							}

							// 运行用户定义回调函数
							if fn != nil {
								err = fn(msg.Timestamp, msg.Info.(DepthData))
								if err != nil {
									log.Println("深度回调函数执行失败！", err)
								}

							}
						}

					}
					//log.Println("退出goroutine  evt:", evt)
				}(evt)

				//continue
			} else {
				//log.Println("程序异常！通道已关闭", evt)
				continue
			}

		}

		//log.Println(evt,"事件已注册",ch)

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond*1000)
		select {
		/*
			丢弃消息容易引发数据处理处理错误
		*/
		// case <-ctx.Done():
		// 	log.Println("等待超时，消息丢弃 - ", data)
		case ch <- &Msg{Timestamp: timestamp, Info: data}:
		}
		cancel()
	}
}

/*
	开启了深度数据管理功能后，系统会自动合并深度信息
*/
func (a *WsClient) MergeDepth(depData DepthData) (err error) {
	if !a.autoDepthMgr {
		return
	}

	key, err := json.Marshal(depData.Arg)
	if err != nil {
		err = errors.New("数据错误")
		return
	}

	// books5 不需要做checksum
	if depData.Arg["channel"] == "books5" {
		a.addDepthDataList(string(key), depData.Data[0])
		return
	}

	if depData.Action == "snapshot" {

		_, err = depData.CheckSum(nil)
		if err != nil {
			log.Println("校验失败", err)
			return
		}

		a.addDepthDataList(string(key), depData.Data[0])

	} else {

		var newSnapshot *DepthDetail
		a.DepthDataLock.RLock()
		oldSnapshot, ok := a.DepthDataList[string(key)]
		if !ok {
			log.Println("深度数据错误，全量数据未发现！")
			err = errors.New("数据错误")
			return
		}
		a.DepthDataLock.RUnlock()
		newSnapshot, err = depData.CheckSum(&oldSnapshot)
		if err != nil {
			log.Println("深度校验失败", err)
			err = errors.New("校验失败")
			return
		}

		a.updateDepthDataList(string(key), *newSnapshot)
	}
	return
}

/*
	通过ErrorCode判断事件类型
*/
func GetInfoFromErrCode(data ErrData) Event {
	switch data.Code {
	case "60001":
		return EVENT_LOGIN
	case "60002":
		return EVENT_LOGIN
	case "60003":
		return EVENT_LOGIN
	case "60004":
		return EVENT_LOGIN
	case "60005":
		return EVENT_LOGIN
	case "60006":
		return EVENT_LOGIN
	case "60007":
		return EVENT_LOGIN
	case "60008":
		return EVENT_LOGIN
	case "60009":
		return EVENT_LOGIN
	case "60010":
		return EVENT_LOGIN
	case "60011":
		return EVENT_LOGIN
	}

	return EVENT_UNKNOWN
}

/*
   从error返回中解析出对应的channel
   error信息样例
 {"event":"error","msg":"channel:index-tickers,instId:BTC-USDT1 doesn't exist","code":"60018"}
*/
func GetInfoFromErrMsg(raw string) (channel string) {
	reg := regexp.MustCompile(`channel:(.*?),`)
	if reg == nil {
		fmt.Println("MustCompile err")
		return
	}
	//提取关键信息
	result := reg.FindAllStringSubmatch(raw, -1)
	for _, text := range result {
		channel = text[1]
	}
	return
}

/*
	解析消息类型
*/
func (a *WsClient) parseMessage(raw []byte) (evt Event, data interface{}, err error) {
	evt = EVENT_UNKNOWN
	//log.Println("解析消息")
	//log.Println("消息内容:", string(raw))
	if string(raw) == "pong" {
		evt = EVENT_PING
		data = raw
		return
	}
	//log.Println(0, evt)
	var rspData = RspData{}
	err = json.Unmarshal(raw, &rspData)
	if err == nil {
		op := rspData.Event
		if op == OP_SUBSCRIBE || op == OP_UNSUBSCRIBE {
			channel := rspData.Arg["channel"]
			evt = GetEventId(channel)
			data = rspData
			return
		}
	}

	//log.Println("ErrData")
	var errData = ErrData{}
	err = json.Unmarshal(raw, &errData)
	if err == nil {
		op := errData.Event
		switch op {
		case OP_LOGIN:
			evt = EVENT_LOGIN
			data = errData
			//log.Println(3, evt)
			return
		case OP_ERROR:
			data = errData
			// TODO:细化报错对应的事件判断

			//尝试从msg字段中解析对应的事件类型
			evt = GetInfoFromErrCode(errData)
			if evt != EVENT_UNKNOWN {
				return
			}
			evt = GetEventId(GetInfoFromErrMsg(errData.Msg))
			if evt == EVENT_UNKNOWN {
				evt = EVENT_ERROR
				return
			}
			return
		}
		//log.Println(5, evt)
	}

	//log.Println("JRPCRsp")
	var jRPCRsp = JRPCRsp{}
	err = json.Unmarshal(raw, &jRPCRsp)
	if err == nil {
		data = jRPCRsp
		evt = GetEventId(jRPCRsp.Op)
		if evt != EVENT_UNKNOWN {
			return
		}
	}

	var depthData = DepthData{}
	err = json.Unmarshal(raw, &depthData)
	if err == nil {
		evt = EVENT_DEPTH_DATA
		data = depthData
		//log.Println("-->>EVENT_DEPTH_DATA", evt)
		//log.Println(evt, data)
		//log.Println(6)
		switch depthData.Arg["channel"] {
		case "books":
			return
		case "books-l2-tbt":
			return
		case "books50-l2-tbt":
			return
		case "books5":
			return
		default:

		}
	}

	//log.Println("MsgData")
	var msgData = MsgData{}
	err = json.Unmarshal(raw, &msgData)
	if err == nil {
		evt = EVENT_BOOKED_DATA
		data = msgData
		//log.Println("-->>EVENT_BOOK_DATA", evt)
		//log.Println(evt, data)
		//log.Println(6)
		return
	}

	evt = EVENT_UNKNOWN
	err = errors.New("message unknown")
	return
}

func (a *WsClient) Stop() error {

	a.lock.Lock()
	defer a.lock.Unlock()
	if !a.isStarted {
		return nil
	}

	a.isStarted = false
	
	if a.conn != nil {
		a.conn.Close()
	}
	close(a.errCh)
	close(a.sendCh)
	close(a.resCh)
	close(a.quitCh)

	for _, ch := range a.regCh {
		close(ch)
	}

	log.Println("ws客户端退出!")
	return nil
}

/*
	添加全局消息处理的回调函数
*/
func (a *WsClient) AddMessageHook(fn ReceivedDataCallback) error {
	a.onMessageHook = fn
	return nil
}

/*
	添加订阅消息处理的回调函数
*/
func (a *WsClient) AddBookMsgHook(fn ReceivedMsgDataCallback) error {
	a.onBookMsgHook = fn
	return nil
}

/*
	添加深度消息处理的回调函数
	例如:
	cli.AddDepthHook(func(ts time.Time, data DepthData) error { return nil })
*/
func (a *WsClient) AddDepthHook(fn ReceivedDepthDataCallback) error {
	a.onDepthHook = fn
	return nil
}

/*
	添加错误类型消息处理的回调函数
*/
func (a *WsClient) AddErrMsgHook(fn ReceivedDataCallback) error {
	a.OnErrorHook = fn
	return nil
}

/*
	判断连接是否存活
*/
func (a *WsClient) IsAlive() bool {
	res := false
	if a.conn == nil {
		return res
	}
	res, _, _ = a.Ping(500)
	return res
}
