/*
	JRPC请求/响应数据
*/
package wImpl

import (
	"encoding/json"
	. "v5sdk_go/utils"
)

// jrpc请求结构体
type JRPCReq struct {
	Id   string                   `json:"id"`
	Op   string                   `json:"op"`
	Args []map[string]interface{} `json:"args"`
}

func (r JRPCReq) GetType() int {
	return MSG_JRPC
}

func (r JRPCReq) ToString() string {
	data, err := Struct2JsonString(r)
	if err != nil {
		return ""
	}
	return data
}

func (r JRPCReq) Len() int {
	return 1
}

// jrpc响应结构体
type JRPCRsp struct {
	Id   string                   `json:"id"`
	Op   string                   `json:"op"`
	Data []map[string]interface{} `json:"data"`
	Code string                   `json:"code"`
	Msg  string                   `json:"msg"`
}

func (r JRPCRsp) MsgType() int {
	return MSG_JRPC
}

func (r JRPCRsp) String() string {
	raw, _ := json.Marshal(r)
	return string(raw)
}
