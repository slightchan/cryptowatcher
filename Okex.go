// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "real.okex.com:10441", "http service address")

const (
	WS_API_URL_OKCOIN = "wss://real.okcoin.com:10440/websocket/okcoinapi"
	WS_API_URL_OKEX   = "wss://real.okex.com:10441/websocket"
)

func connectOkex() {
	flag.Parse()
	log.SetFlags(0)
	log.Println(websocket.DefaultDialer)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/websocket"}
	log.Printf("connecting to %s", u.String())
	var dialer = websocket.Dialer{ReadBufferSize: 10000, WriteBufferSize: 1000}
	dialer.EnableCompression = true
	//dialer.Proxy = nil
	//dialer.HandshakeTimeout = time.Duration(10)*time.Second

	c, _, err := dialer.Dial(WS_API_URL_OKEX, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Dial OK")
	defer c.Close()

	done := make(chan struct{})
	c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
	ticker := time.NewTicker(time.Second)
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			//err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			//if err != nil {
			//log.Println("write:", err)
			//return
			//}
			c.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
			log.Println(t)
			//			return
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
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
