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
	High      float64 `json:"high"`
	Vol       float64 `json:"vol"`
	Last      float64 `json:"last"`
	Low       float64 `json:"low"`
	Buy       float64 `json:"buy"`
	Change    float64 `json:"change"`
	Sell      float64 `json:"sell"`
	DayLow    float64 `json:"dayLow"`
	DayHigh   float64 `json:"dayHigh"`
	Open      float64 `json:"open"`
	Timestamp int64   `json:"timestamp"`
}

type AddChannelResult struct {
	Result  bool   `json:"result"`
	Channel string `json:"channel"`
}

var Done = make(chan int)

func ConnectOkex() {
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	kill := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	signal.Notify(kill, os.Kill)
	var dialer = websocket.Dialer{ReadBufferSize: 10000, WriteBufferSize: 1000}
	dialer.EnableCompression = true
	//dialer.Proxy = nil
	//dialer.HandshakeTimeout = time.Duration(10)*time.Second

	c, _, err := dialer.Dial(WS_API_URL_OKEX, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Dial ok")
	defer c.Close()

	c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	go onRecv(c)

	addChannel(c, "ok_sub_spot_btc_usdt_ticker")
	for {
		select {
		case <-ticker.C:
			c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
		case <-kill:
			log.Println("kill")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			//c.Wri
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func addChannel(c *websocket.Conn, channel string) {
	cmd := OKCmd{}
	cmd.Event = "addChannel"
	cmd.Channel = channel
	cmd.Binary = 0
	err := c.WriteJSON(cmd)
	if err != nil {
		log.Println("write close:", err)
		Done <- 1
		return
	}
}

func onRecv(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			Done <- 1
			return
		}
		log.Printf("recv: %s", message)
		if json.Valid(message) {
			log.Println("Is a valid json data")
		} else {
			log.Println("Is not a valid json data")
			return
		}
		if message[0] != byte('[') {
			var pongEvent EventData
			err = json.Unmarshal(message, &pongEvent)
			if err != nil {
				log.Println(err)
			} else {
				log.Println(pongEvent)
			}
		} else {
			//var j []map[string]interface{}
			//err = json.Unmarshal(message, &j)
			var channels []ChannelData
			err = json.Unmarshal(message, &channels)
			if err != nil {
				log.Println(err)
			} else {
				for i := 0; i < len(channels); i++ {
					c := channels[i]
					if c.Channel == "addChannel" {
						onAddChannelResult(c.Data)
					} else {
						onChannelData(c.Data)
					}
				}
			}
		}

		//<-time.After(time.Second * 5)
		//Done <- 1
	}
}

func onChannelData(d json.RawMessage) {
	var r SpotData
	err := json.Unmarshal(d, &r)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(r)
	}
}

func onAddChannelResult(d json.RawMessage) {
	var r AddChannelResult
	err := json.Unmarshal(d, &r)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(r)
	}
}
