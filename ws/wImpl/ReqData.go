/*
	普通订阅请求和响应的数据格式
*/

package wImpl

import (
	"encoding/json"
	. "v5sdk_go/utils"
)

// 客户端请求消息格式
type ReqData struct {
	Op   string              `json:"op"`
	Args []map[string]string `json:"args"`
}

func (r ReqData) GetType() int {
	return MSG_NORMAL
}

func (r ReqData) ToString() string {
	data, err := Struct2JsonString(r)
	if err != nil {
		return ""
	}
	return data
}

func (r ReqData) Len() int {
	return len(r.Args)
}

// 服务端请求响应消息格式
type RspData struct {
	Event string            `json:"event"`
	Arg   map[string]string `json:"arg"`
}

func (r RspData) MsgType() int {
	return MSG_NORMAL
}

func (r RspData) String() string {
	raw, _ := json.Marshal(r)
	return string(raw)
}
