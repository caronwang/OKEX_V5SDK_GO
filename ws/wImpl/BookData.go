/*
	订阅频道后收到的推送数据
*/

package wImpl

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"strconv"
)

// 普通推送
type MsgData struct {
	Arg  map[string]string `json:"arg"`
	Data []interface{}     `json:"data"`
}

// 深度数据
type DepthData struct {
	Arg    map[string]string `json:"arg"`
	Action string            `json:"action"`
	Data   []DepthDetail     `json:"data"`
}

type DepthDetail struct {
	Asks     [][]string `json:"asks"`
	Bids     [][]string `json:"bids"`
	Ts       string     `json:"ts"`
	Checksum int32      `json:"checksum"`
}

/*
	深度数据校验
*/
func (this *DepthData) CheckSum(snap *DepthDetail) (pDepData *DepthDetail, err error) {

	if len(this.Data) != 1 {
		err = errors.New("深度数据错误！")
		return
	}

	if this.Action == DEPTH_SNAPSHOT {
		_, cs := CalCrc32(this.Data[0].Asks, this.Data[0].Bids)

		if cs != this.Data[0].Checksum {
			err = errors.New("校验失败！")
			return
		}
		pDepData = &this.Data[0]
		log.Println("snapshot校验成功", this.Data[0].Checksum)

	}

	if this.Action == DEPTH_UPDATE {
		if snap == nil {
			err = errors.New("深度快照数据不可为空！")
			return
		}

		pDepData, err = MergDepthData(*snap, this.Data[0], this.Data[0].Checksum)
		if err != nil {
			return
		}
		log.Println("update校验成功", this.Data[0].Checksum)
	}

	return
}

func CalCrc32(askDepths [][]string, bidDepths [][]string) (bytes.Buffer, int32) {

	crc32BaseBuffer := bytes.Buffer{}
	crcAskDepth, crcBidDepth := 25, 25

	if len(askDepths) < 25 {
		crcAskDepth = len(askDepths)
	}
	if len(bidDepths) < 25 {
		crcBidDepth = len(bidDepths)
	}
	if crcAskDepth == crcBidDepth {
		for i := 0; i < crcAskDepth; i++ {
			if crc32BaseBuffer.Len() > 0 {
				crc32BaseBuffer.WriteString(":")
			}
			crc32BaseBuffer.WriteString(
				fmt.Sprintf("%v:%v:%v:%v",
					(bidDepths)[i][0], (bidDepths)[i][1],
					(askDepths)[i][0], (askDepths)[i][1]))
		}
	} else {

		var crcArr []string
		for i, j := 0, 0; i < crcBidDepth || j < crcAskDepth; {

			if i < crcBidDepth {
				crcArr = append(crcArr, fmt.Sprintf("%v:%v", (bidDepths)[i][0], (bidDepths)[i][1]))
				i++
			}

			if j < crcAskDepth {
				crcArr = append(crcArr, fmt.Sprintf("%v:%v", (askDepths)[j][0], (askDepths)[j][1]))
				j++
			}
		}

		crc32BaseBuffer.WriteString(strings.Join(crcArr, ":"))
	}

	expectCrc32 := int32(crc32.ChecksumIEEE(crc32BaseBuffer.Bytes()))

	return crc32BaseBuffer, expectCrc32
}

/*
	深度合并的内部方法
	返回结果：
	res：合并后的深度
	index: 最新的 ask1/bids1 的索引
*/
func mergeDepth(oldDepths [][]string, newDepths [][]string, method string) (res [][]string, err error) {

	oldIdx, newIdx := 0, 0

	for oldIdx < len(oldDepths) && newIdx < len(newDepths) {

		oldItem := oldDepths[oldIdx]
		newItem := newDepths[newIdx]
		var oldPrice, newPrice float64
		oldPrice, err = strconv.ParseFloat(oldItem[0], 10)
		if err != nil {
			return
		}
		newPrice, err = strconv.ParseFloat(newItem[0], 10)
		if err != nil {
			return
		}

		if oldPrice == newPrice {
			if newItem[1] != "0" {
				res = append(res, newItem)
			}

			oldIdx++
			newIdx++
		} else {
			switch method {
			// 降序
			case "bids":
				if oldPrice < newPrice {
					res = append(res, newItem)
					newIdx++
				} else {

					res = append(res, oldItem)
					oldIdx++
				}
			// 升序
			case "asks":
				if oldPrice > newPrice {
					res = append(res, newItem)
					newIdx++
				} else {

					res = append(res, oldItem)
					oldIdx++
				}
			}
		}

	}

	if oldIdx < len(oldDepths) {
		res = append(res, oldDepths[oldIdx:]...)
	}

	if newIdx < len(newDepths) {
		res = append(res, newDepths[newIdx:]...)
	}

	return
}

/*
	深度合并，并校验
*/
func MergDepthData(snap DepthDetail, update DepthDetail, expChecksum int32) (res *DepthDetail, err error) {

	newAskDepths, err1 := mergeDepth(snap.Asks, update.Asks, "asks")
	if err1 != nil {
		return
	}

	// log.Println("old Ask - ", snap.Asks)
	// log.Println("update Ask - ", update.Asks)
	// log.Println("new Ask - ", newAskDepths)
	newBidDepths, err2 := mergeDepth(snap.Bids, update.Bids, "bids")
	if err2 != nil {
		return
	}
	// log.Println("old Bids - ", snap.Bids)
	// log.Println("update Bids - ", update.Bids)
	// log.Println("new Bids - ", newBidDepths)

	cBuf, checksum := CalCrc32(newAskDepths, newBidDepths)
	if checksum != expChecksum {
		err = errors.New("校验失败！")
		log.Println("buffer:", cBuf.String())
		log.Fatal(checksum, expChecksum)
		return
	}

	res = &DepthDetail{
		Asks:     newAskDepths,
		Bids:     newBidDepths,
		Ts:       update.Ts,
		Checksum: update.Checksum,
	}

	return
}
