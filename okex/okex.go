// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package okex

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WS_API_URL_OKCOIN = "wss://real.okcoin.com:10440/websocket/okcoinapi"
	WS_API_URL_OKEX   = "wss://real.okex.com:10441/websocket"
	WS_PROXY          = "ws://139.162.74.36:9158/websocket"
)

type OKCmd struct {
	Event   string      `json:"event"`
	Channel string      `json:"channel"`
	Binary  int         `json:"binary,omitempty"`
	Params  *OKCmdParam `json:"parameters,omitempty"`
}

type OKCmdParam struct {
	ApiKey string `json:api_key`
	Sign   string `json:sign`
}

type ChannelData struct {
	Binary  int             `json:"binary"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type EventData struct {
	Event string `json:"event"`
}

type SpotData struct {
	High      float64 `json:"high,string"`
	Vol       float64 `json:"vol,string"`
	Last      float64 `json:"last,string"`
	Low       float64 `json:"low,string"`
	Buy       float64 `json:"buy,string"`
	Change    float64 `json:"change,string"`
	Sell      float64 `json:"sell,string"`
	DayLow    float64 `json:"dayLow,string"`
	DayHigh   float64 `json:"dayHigh,string"`
	Open      float64 `json:"open,string"`
	Timestamp int64   `json:"timestamp"`
}

type AddChannelResult struct {
	Result  bool   `json:"result"`
	Channel string `json:"channel"`
}

var Done = make(chan int)
var cryptoPrices map[string]SpotData = map[string]SpotData{}

func ConnectOkex() {
	log.SetFlags(0)

	var dialer = websocket.Dialer{ReadBufferSize: 4096, WriteBufferSize: 4096}
	dialer.EnableCompression = true
	dialer.HandshakeTimeout = time.Duration(10) * time.Second
	//dialer.Proxy = http.ProxyFromEnvironment
	log.Println("Dialing...")
	c, _, err := dialer.Dial(WS_PROXY, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Dial ok")

	err = addChannel(c, "ok_sub_spot_btc_usdt_ticker")
	if err != nil {
		closeConnection(c)
	} else {
		go onRecv(c)
	}
}

func closeConnection(c *websocket.Conn) {
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	<-time.After(time.Second)
	Done <- 1
}

func addChannel(c *websocket.Conn, channel string) error {
	cmd := OKCmd{}
	cmd.Event = "addChannel"
	cmd.Channel = channel
	cmd.Binary = 0
	err := c.WriteJSON(cmd)
	if err != nil {
		log.Println("write close:", err)
	}
	return err
}

func onRecv(c *websocket.Conn) {
	interrupt := make(chan os.Signal, 1)
	kill := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(kill, os.Kill)
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		var message []byte
		var err error
		var msgType int
		select {
		case <-ticker.C:
			c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
			continue
		case <-kill:
			log.Println("kill")
			closeConnection(c)
			return
		case <-interrupt:
			log.Println("interrupt")
			closeConnection(c)
			return
		default:
			msgType, message, err = c.ReadMessage()
		}
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("recv: %s", message)
		if !json.Valid(message) {
			log.Printf("Invalid Data Format,Message Type:%d", msgType)
			log.Println(string(message))
			continue
		}
		if message[0] != byte('[') {
			var pongEvent EventData
			err = json.Unmarshal(message, &pongEvent)
			if err != nil {
				log.Println(err)
				break
			}
			if pongEvent.Event == "pong" {
				log.Println("Pong Event From Server")
			} else {
				log.Println("Unhandled Message Type:")
				log.Println(string(message))
				break
			}
		} else {
			var channels []ChannelData
			err = json.Unmarshal(message, &channels)
			if err != nil {
				log.Println(err)
				break
			} else {
				for i := 0; i < len(channels); i++ {
					c := channels[i]
					if c.Channel == "addChannel" {
						onAddChannelResult(c)
					} else {
						onChannelData(c)
					}
				}
			}
		}
	}
	log.Println("Receive Process Ended")
	closeConnection(c)
}

func onChannelData(c ChannelData) {
	var r SpotData
	err := json.Unmarshal(c.Data, &r)
	if err != nil {
		log.Println(err)
	} else {
		r.Timestamp = r.Timestamp / 1000
		cryptoPrices[c.Channel] = r
		log.Println(r)
	}
}

func onAddChannelResult(c ChannelData) {
	var r AddChannelResult
	err := json.Unmarshal(c.Data, &r)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(r)
	}
}
