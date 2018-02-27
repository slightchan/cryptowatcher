package main

import (
	"fmt"
)

var gPriceData map[string]PriceData

func main() {
	gPriceData = map[string]PriceData{}
	UpdateCurrencyPrices()
	fmt.Println(gPriceData["USD/JPY"])
	fmt.Println(gPriceData["USD/CNY"])
	fmt.Printf("CNY/JPY,%.4f\n", gPriceData["USD/CNY"].Price/gPriceData["USD/JPY"].Price)
}
