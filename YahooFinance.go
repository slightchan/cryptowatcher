// YahooFinance
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

type PriceList struct {
	XMLName   xml.Name       `xml:"list"` // 指定最外层的标签为config
	Version   string         `xml:"version,attr"`
	Resources PriceResources `xml:"resources"`
	//SmtpServer string `xml:"smtpServer"` // 读取smtpServer配置项，并将结果保存到SmtpServer变量中
	//SmtpPort int `xml:"smtpPort"`
	//Sender string `xml:"sender"`
	//SenderPasswd string `xml:"senderPasswd"`
	//Receivers SReceivers `xml:"receivers"` // 读取receivers标签下的内容，以结构方式获取
}

type PriceResources struct {
	XMLName   xml.Name        `xml:"resources"`
	Start     uint32          `xml:"start,attr"`
	Count     uint32          `xml:"count,attr"`
	Resources []PriceResource `xml:"resource"`
}

type PriceResource struct {
	XMLName xml.Name     `xml:"resource"`
	Fields  []PriceField `xml:"field"`
}

type PriceField struct {
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

type PriceData struct {
	Name       string
	Price      float64
	UpdateTime time.Time
}

func UpdatePriceData() {
	body := GetYahooFinanceQuote()
	//fmt.Print(string(body))
	var result PriceList
	err := xml.Unmarshal(body, &result)
	//err = ioutil.WriteFile("/Users/logan/Downloads/financeYahoo.xml",body,0666)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(result.XMLName)
	//fmt.Println(result.Version)
	//fmt.Println(result.Resources.XMLName)
	//fmt.Println(result.Resources.Start)
	fmt.Printf("Price count:%d\n", result.Resources.Count)
	//fmt.Println(len(result.Resources.Resources))
	for i := range result.Resources.Resources {
		var data PriceData
		for j := range result.Resources.Resources[i].Fields {
			fieldName := result.Resources.Resources[i].Fields[j].Name
			fieldValue := result.Resources.Resources[i].Fields[j].Value
			if fieldName == "name" {
				//fmt.Println()
				data.Name = fieldValue
			}
			if fieldName == "price" {
				data.Price, _ = strconv.ParseFloat(fieldValue, 64)
			}
			if fieldName == "utctime" {
				data.UpdateTime, err = time.Parse("2006-01-02T15:04:05-0700", fieldValue)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		//fmt.Println(data)
		gPriceData[data.Name] = data
	}
}

func GetYahooFinanceQuote() []byte {
	cacheFile := path.Join(os.TempDir(), "financeYahoo.xml")
	fmt.Println(cacheFile)
	res, err := http.Get("https://finance.yahoo.com/webservice/v1/symbols/allcurrencies/quote")
	if err == nil {
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err == nil {
			err = ioutil.WriteFile(cacheFile, body, 0666)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Updated from Yahoo")
			return body
		}
	}
	log.Println(err)
	log.Println("Load from cache file")
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	return data
}
