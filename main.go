package main

import (
	"cryptowatcher/okex"
	"fmt"
	"time"
)

var gPriceData map[string]PriceData

func main() {
	gPriceData = map[string]PriceData{}
	go UpdateCurrencyPrices()
	<-time.After(time.Second * 5)
	fmt.Println(gPriceData["USD/JPY"])
	fmt.Println(gPriceData["USD/CNY"])
	fmt.Printf("CNY/JPY,%.4f\n", gPriceData["USD/CNY"].Price/gPriceData["USD/JPY"].Price)
	go okex.ConnectOkex()
	<-okex.Done
}
