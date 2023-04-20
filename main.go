package main

import (
	"fmt"
	"serviceascendex/ascendex"
	"time"
)

func main() {
	asc := ascendex.GetAPIClient()
	if err := asc.Connection(); err != nil {
		fmt.Println(err)
		return
	}
	if err := asc.SubscribeToChannel("USDT_BTC"); err != nil {
		fmt.Println(err)
		return
	}
	ch := make(chan ascendex.BestOrderBook)
	asc.WriteMessagesToChannel()
	asc.ReadMessagesFromChannel(ch)
	go func() {
		for bob := range ch {
			fmt.Println(bob)
		}
	}()
	time.Sleep(time.Minute * 5)
	asc.Disconnect()

}
