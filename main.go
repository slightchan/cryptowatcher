package main

import (
	"fmt"
)

var gPriceData map[string]PriceData

func main() {
	gPriceData = map[string]PriceData{}
	UpdatePriceData()
	fmt.Println(gPriceData["USD/JPY"])
	fmt.Println(gPriceData["USD/CNY"])
}
