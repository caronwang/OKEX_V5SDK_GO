package wImpl

import (
	"regexp"
)

/*

 */

const (
	MSG_NORMAL = iota
	MSG_JRPC
)

//事件
type Event int

/*
	EventID
*/
const (
	EVENT_UNKNOWN Event = iota
	EVENT_ERROR
	EVENT_PING
	EVENT_LOGIN

	//订阅公共频道
	EVENT_BOOK_INSTRUMENTS
	EVENT_STATUS
	EVENT_BOOK_TICKERS
	EVENT_BOOK_OPEN_INTEREST
	EVENT_BOOK_KLINE
	EVENT_BOOK_TRADE
	EVENT_BOOK_ESTIMATE_PRICE
	EVENT_BOOK_MARK_PRICE
	EVENT_BOOK_MARK_PRICE_CANDLE_CHART
	EVENT_BOOK_LIMIT_PRICE
	EVENT_BOOK_ORDER_BOOK
	EVENT_BOOK_ORDER_BOOK5
	EVENT_BOOK_ORDER_BOOK_TBT
	EVENT_BOOK_ORDER_BOOK50_TBT
	EVENT_BOOK_OPTION_SUMMARY
	EVENT_BOOK_FUND_RATE
	EVENT_BOOK_KLINE_INDEX
	EVENT_BOOK_INDEX_TICKERS

	//订阅私有频道
	EVENT_BOOK_ACCOUNT
	EVENT_BOOK_POSTION
	EVENT_BOOK_ORDER
	EVENT_BOOK_ALG_ORDER
	EVENT_BOOK_B_AND_P

	// JRPC
	EVENT_PLACE_ORDER
	EVENT_PLACE_BATCH_ORDERS
	EVENT_CANCEL_ORDER
	EVENT_CANCEL_BATCH_ORDERS
	EVENT_AMEND_ORDER
	EVENT_AMEND_BATCH_ORDERS

	//订阅返回数据
	EVENT_BOOKED_DATA
	EVENT_DEPTH_DATA
)

/*
	EventID，事件名称，channel
	注： 带有周期参数的频道 如 行情频道 ，需要将channel写为 正则表达模式方便 类型匹配，如 "^candle*"
*/
var EVENT_TABLE = [][]interface{}{
	// 未知的消息
	{EVENT_UNKNOWN, "未知", ""},
	// 错误的消息
	{EVENT_ERROR, "错误", ""},
	// Ping
	{EVENT_PING, "ping", ""},
	// 登陆
	{EVENT_LOGIN, "登录", ""},

	/*
		订阅公共频道
	*/

	{EVENT_BOOK_INSTRUMENTS, "产品", "instruments"},
	{EVENT_STATUS, "status", "status"},
	{EVENT_BOOK_TICKERS, "行情", "tickers"},
	{EVENT_BOOK_OPEN_INTEREST, "持仓总量", "open-interest"},
	{EVENT_BOOK_KLINE, "K线", "candle"},
	{EVENT_BOOK_TRADE, "交易", "trades"},
	{EVENT_BOOK_ESTIMATE_PRICE, "预估交割/行权价格", "estimated-price"},
	{EVENT_BOOK_MARK_PRICE, "标记价格", "mark-price"},
	{EVENT_BOOK_MARK_PRICE_CANDLE_CHART, "标记价格K线", "mark-price-candle"},
	{EVENT_BOOK_LIMIT_PRICE, "限价", "price-limit"},
	{EVENT_BOOK_ORDER_BOOK, "400档深度", "books"},
	{EVENT_BOOK_ORDER_BOOK5, "5档深度", "books5"},
	{EVENT_BOOK_ORDER_BOOK_TBT, "tbt深度", "books-l2-tbt"},
	{EVENT_BOOK_ORDER_BOOK50_TBT, "tbt50深度", "books50-l2-tbt"},
	{EVENT_BOOK_OPTION_SUMMARY, "期权定价", "opt-summary"},
	{EVENT_BOOK_FUND_RATE, "资金费率", "funding-rate"},
	{EVENT_BOOK_KLINE_INDEX, "指数K线", "index-candle"},
	{EVENT_BOOK_INDEX_TICKERS, "指数行情", "index-tickers"},

	/*
		订阅私有频道
	*/
	{EVENT_BOOK_ACCOUNT, "账户", "account"},
	{EVENT_BOOK_POSTION, "持仓", "positions"},
	{EVENT_BOOK_ORDER, "订单", "orders"},
	{EVENT_BOOK_ALG_ORDER, "策略委托订单", "orders-algo"},
	{EVENT_BOOK_B_AND_P, "账户余额和持仓", "balance_and_position"},

	/*
		JRPC
	*/
	{EVENT_PLACE_ORDER, "下单", "order"},
	{EVENT_PLACE_BATCH_ORDERS, "批量下单", "batch-orders"},
	{EVENT_CANCEL_ORDER, "撤单", "cancel-order"},
	{EVENT_CANCEL_BATCH_ORDERS, "批量撤单", "batch-cancel-orders"},
	{EVENT_AMEND_ORDER, "改单", "amend-order"},
	{EVENT_AMEND_BATCH_ORDERS, "批量改单", "batch-amend-orders"},

	/*
		订阅返回数据
		注意：推送数据channle统一为""
	*/
	{EVENT_BOOKED_DATA, "普通推送", ""},
	{EVENT_DEPTH_DATA, "深度推送", ""},
}

/*
	获取事件名称
*/
func (e Event) String() string {
	for _, v := range EVENT_TABLE {
		eventId := v[0].(Event)
		if e == eventId {
			return v[1].(string)
		}
	}

	return ""
}

/*
	通过事件获取对应的channel信息
	对于频道名称有时间周期的 通过参数 pd 传入，拼接后返回完整channel信息
*/
func (e Event) GetChannel(pd Period) string {

	channel := ""

	for _, v := range EVENT_TABLE {
		eventId := v[0].(Event)
		if e == eventId {
			channel = v[2].(string)
			break
		}
	}

	if channel == "" {
		return ""
	}

	return channel + string(pd)
}

/*
	通过channel信息匹配获取事件类型
*/
func GetEventId(raw string) Event {
	evt := EVENT_UNKNOWN

	for _, v := range EVENT_TABLE {
		channel := v[2].(string)
		if raw == channel {
			evt = v[0].(Event)
			break
		}

		regexp := regexp.MustCompile(`^(.*)([1-9][0-9]?[\w])$`)
		//regexp := regexp.MustCompile(`^http://www.flysnow.org/([\d]{4})/([\d]{2})/([\d]{2})/([\w-]+).html$`)

		substr := regexp.FindStringSubmatch(raw)
		//fmt.Println(substr)
		if len(substr) >= 2 {
			if substr[1] == channel {
				evt = v[0].(Event)
				break
			}
		}
	}

	return evt
}

// 时间维度
type Period string

const (
	// 年
	PERIOD_1YEAR Period = "1Y"

	// 月
	PERIOD_6Mon Period = "6M"
	PERIOD_3Mon Period = "3M"
	PERIOD_1Mon Period = "1M"

	// 周
	PERIOD_1WEEK Period = "1W"

	// 天
	PERIOD_5DAY Period = "5D"
	PERIOD_3DAY Period = "3D"
	PERIOD_2DAY Period = "2D"
	PERIOD_1DAY Period = "1D"

	// 小时
	PERIOD_12HOUR Period = "12H"
	PERIOD_6HOUR  Period = "6H"
	PERIOD_4HOUR  Period = "4H"
	PERIOD_2HOUR  Period = "2H"
	PERIOD_1HOUR  Period = "1H"

	// 分钟
	PERIOD_30MIN Period = "30m"
	PERIOD_15MIN Period = "15m"
	PERIOD_5MIN  Period = "5m"
	PERIOD_3MIN  Period = "3m"
	PERIOD_1MIN  Period = "1m"

	// 缺省
	PERIOD_NONE Period = ""
)

// 深度枚举
const (
	DEPTH_SNAPSHOT = "snapshot"
	DEPTH_UPDATE   = "update"
)
