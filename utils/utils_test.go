package utils

import (
	"fmt"
	"testing"
)

func TestHmacSha256Base64Signer(t *testing.T) {
	raw := `2021-04-06T03:33:21.681ZPOST/api/v5/trade/order{"instId":"ETH-USDT-SWAP","ordType":"limit","px":"2300","side":"sell","sz":"1","tdMode":"cross"}`
	key := "1A9E86759F2D2AA16E389FD3B7F8273E"
	res, err := HmacSha256Base64Signer(raw, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
	t.Log(res)
}
