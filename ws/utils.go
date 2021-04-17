package ws

import (
	"errors"
	"log"
	"runtime/debug"
	. "v5sdk_go/ws/wImpl"
	. "v5sdk_go/ws/wInterface"
)

// 判断返回结果成功失败
func checkResult(wsReq WSReqData, wsRsps []*Msg) (res bool, err error) {
	defer func() {
		a := recover()
		if a != nil {
			log.Printf("Receive End. Recover msg: %+v", a)
			debug.PrintStack()
		}
		return
	}()

	res = false
	if len(wsRsps) == 0 {
		return
	}

	for _, v := range wsRsps {
		switch v.Info.(type) {
		case ErrData:
			return
		}
		if wsReq.GetType() != v.Info.(WSRspData).MsgType() {
			err = errors.New("消息类型不一致")
			return
		}
	}

	//检查所有频道是否都更新成功
	if wsReq.GetType() == MSG_NORMAL {
		req, ok := wsReq.(ReqData)
		if !ok {
			log.Println("类型转化失败", req)
			err = errors.New("类型转化失败")
			return
		}

		for idx, _ := range req.Args {
			ok := false
			i_req := req.Args[idx]
			//fmt.Println("检查",i_req)
			for i, _ := range wsRsps {
				info, _ := wsRsps[i].Info.(RspData)
				//fmt.Println("<<",info)
				if info.Event == req.Op && info.Arg["channel"] == i_req["channel"] && info.Arg["instType"] == i_req["instType"] {
					ok = true
					continue
				}
			}
			if !ok {
				err = errors.New("未得到所有的期望的返回结果")
				return
			}
		}
	} else {
		for i, _ := range wsRsps {
			info, _ := wsRsps[i].Info.(JRPCRsp)
			if info.Code != "0" {
				return
			}
		}
	}

	res = true
	return
}
