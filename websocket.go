// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type WebSocket struct {
	conn   *websocket.Conn   // websocket连接
	config map[string]string // 配置项
	//Reconnect      bool                                    // 断线是否自动重连
	//Status         int                                     // 当前连接状态
	//IsLock         bool                                    // 当前锁状态
	//Locker         sync.Mutex                              // 锁
	//Path           string                                  // 路径
	interrupt chan os.Signal // 中断/关闭处理
	//Message        chan []byte                             // 消息队列
	messageHandler func(webSocket *WebSocket, message []byte) // 消息处理器
	//Errs           chan string                             // 错误消息队列
	ticker *time.Ticker // 定时器
	//Exit           chan bool                               // 退出
	//Expire         chan bool                               // 有效
	//ExpireTime     int                                     // 有效检测时间（秒）
	url string
}

func (o *WebSocket) SetConfig(config map[string]string) {
	o.config = config
}

func (o *WebSocket) GetConn() *websocket.Conn {
	return o.conn
}

// NewWebsocket 实例化WebSocket
func NewWebsocket() *WebSocket {
	o := &WebSocket{}
	return o
}

func (o *WebSocket) initialize() {
	flag.Parse()
	log.SetFlags(0)

	o.interrupt = make(chan os.Signal, 1)
	signal.Notify(o.interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: o.config["HOST"] + ":" + o.config["PORT"], Path: o.config["PATH"]}
	o.url = u.String()
	glog.Infof("connecting to %s", o.url)

	o.reconnect()

	o.ticker = time.NewTicker(time.Second)
}

func (o *WebSocket) reconnect() {
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	i := 0
	for {
		err := o.connect()
		if err != nil {
			time.Sleep(time.Second * 2)
		} else {
			break
		}
		i++
	}
}

func (o *WebSocket) connect() error {
	var err error
	glog.Infof("reconnecting to %s", o.url)

	o.conn, _, err = websocket.DefaultDialer.Dial(o.url, nil)
	if err != nil {
		glog.Errorf("dial: %v", err)
	}
	return err
}

func (o *WebSocket) SetMessageHandler(messageHandler func(webSocket *WebSocket, message []byte)) {
	o.messageHandler = messageHandler
}

func (o *WebSocket) Run() {
	o.initialize()

	defer o.conn.Close()

	defer o.ticker.Stop()

	go func() {
		for {
			if o.conn != nil {
				_, message, err := o.conn.ReadMessage()
				if err != nil {
					glog.Errorf("read: %v", err)
					o.reconnect()
					continue
				}
				o.messageHandler(o, message)

			}

		}
	}()

	for {
		select {
		case _ = <-o.ticker.C:
			if o.conn != nil {
				err := o.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
				if err != nil {
					glog.Errorf("write: %v", err)
					o.reconnect()
				}
			}

		case <-o.interrupt:
			glog.Infof("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			if o.conn != nil {
				err := o.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					glog.Errorf("write close: %v", err)
					return
				}
			}
			return

		}
	}
}
