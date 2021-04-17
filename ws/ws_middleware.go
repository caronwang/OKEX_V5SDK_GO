package ws

import "fmt"

type ReqFunc func(...interface{}) (res bool, msg *Msg, err error)
type Decorator func(ReqFunc) ReqFunc

func handler(h ReqFunc, decors ...Decorator) ReqFunc {
	for i := range decors {
		d := decors[len(decors)-1-i]
		h = d(h)
	}
	return h
}

func preprocess() (res bool, msg *Msg, err error) {
	fmt.Println("preprocess")
	return
}
