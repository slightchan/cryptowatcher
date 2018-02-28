// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
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

var gDone = make(chan int)

func connectOkex() {
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

func onRecv(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			gDone <- 1
			return
		}
		log.Printf("recv: %s", message)
		<-time.After(time.Second * 5)
		gDone <- 1
	}
}
