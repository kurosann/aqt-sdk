package okx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var config = KeyConfig{
	"",
	"",
	"",
}

func TestNewWsClient(t *testing.T) {
	client := NewWsClient(
		context.Background(),
		config,
		TestServer,
	)
	count := 2
	cond := sync.NewCond(&sync.RWMutex{})
	go func() {
		if err := client.Trades("BTC-USDT_BTC-USDT-240329", func(resp *WsResp[Trades]) {
			t.Log(*resp)
			cond.L.Lock()
			count--
			cond.Broadcast()
			cond.L.Unlock()
		}); err != nil {
			assert.Fail(t, err.Error())
			return
		}
	}()
	//go func() {
	//	ch, err := client.MarkPrice("BTC-USDT")
	//	if err != nil {
	//		assert.Fail(t, err.Error())
	//		return
	//	}
	//	for {
	//		resp, ok := <-ch
	//		if !ok {
	//			return
	//		}
	//		fmt.Println(*resp)
	//		cond.L.Lock()
	//		count--
	//		cond.Broadcast()
	//		cond.L.Unlock()
	//	}
	//}()
	isFailed := false
	go func() {
		<-time.After(time.Second * 20)

		cond.L.Lock()
		isFailed = true
		cond.Broadcast()
		cond.L.Unlock()
	}()
	cond.L.Lock()
	for count != 0 && !isFailed {
		cond.Wait()
	}
	cond.L.Unlock()
	assert.Equal(t, count, 0)
}
