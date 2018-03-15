package main

import (
	"cryptowatcher/okex"
)

var gPriceData map[string]PriceData

func main() {
	gPriceData = map[string]PriceData{}
	go UpdateCurrencyPrices()
	go okex.ConnectOkex()
	<-okex.Done
}
