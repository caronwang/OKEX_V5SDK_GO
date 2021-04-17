package wInterface

// 请求数据
type WSReqData interface {
	GetType() int
	Len() int
	ToString() string
}
