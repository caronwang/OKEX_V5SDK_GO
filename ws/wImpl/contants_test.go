package wImpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEventId(t *testing.T) {

	id1 := GetEventId("index-candle30m")

	assert.True(t, id1 == EVENT_BOOK_KLINE_INDEX)

	id2 := GetEventId("candle1Y")

	assert.True(t, id2 == EVENT_BOOK_KLINE)

	id3 := GetEventId("orders-algo")
	assert.True(t, id3 == EVENT_BOOK_ALG_ORDER)

	id4 := GetEventId("balance_and_position")
	assert.True(t, id4 == EVENT_BOOK_B_AND_P)

	id5 := GetEventId("index-candle1m")
	assert.True(t, id5 == EVENT_BOOK_KLINE_INDEX)

}
