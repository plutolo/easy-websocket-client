// author: luokuanxing <346300265@qq.com>
// date: 2022/7/15

package websocket

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WebsocketStatusOnline  = 1  // websocket服务在线中
	WebsocketStatusOffline = 0  // websocket服务不在线
	UnKnowMessageType      = -1 // 未知消息类型
	ExpireAlways           = 0  // 一直有效
)

type WebSocket struct {
	conn           *websocket.Conn                         // websocket连接
	config         map[string]string                       // 配置项
	reconnect      bool                                    // 断线是否自动重连
	status         int                                     // 当前连接状态
	isLock         bool                                    // 当前锁状态
	locker         sync.Mutex                              // 锁
	addr           string                                  // 连接地址
	path           string                                  // 路径
	interrupt      chan os.Signal                          // 中断/关闭处理
	message        chan []byte                             // 消息队列
	messageHandler func(socket *WebSocket, message []byte) // 消息处理器
	errs           chan string                             // 错误消息队列
	ticker         *time.Ticker                            // 定时器
	done           chan struct{}                           // 结束
	exit           chan bool                               // 退出
	expire         chan bool                               // 有效
	expireTime     int                                     // 有效检测时间（秒）
}

func (o *WebSocket) GetConn() *websocket.Conn {
	return o.conn
}

func (o *WebSocket) SetConfig(config map[string]string) *WebSocket {
	o.config = config
	return o
}

func (o *WebSocket) GetConfig() map[string]string {
	return o.config
}

func (o *WebSocket) WriteErrorMsg(errorMsg string) {
	o.errs <- errorMsg
}

// NewWebsocket 实例化WebSocket
func NewWebsocket() *WebSocket {
	o := &WebSocket{}
	return o
}

// lock 加锁
func (o *WebSocket) lock() {
	o.locker.Lock()
	o.isLock = true
}

// unLock 解锁
func (o *WebSocket) unLock() {
	o.locker.Unlock()
	o.isLock = false
}

// initialize 初始化
func (o *WebSocket) initialize() *WebSocket {
	log.SetFlags(0)

	o.locker = sync.Mutex{}
	o.interrupt = make(chan os.Signal, 1)
	o.message = make(chan []byte)
	o.errs = make(chan string)
	o.done = make(chan struct{})
	o.ticker = time.NewTicker(time.Second)
	o.exit = make(chan bool)
	o.expire = make(chan bool)
	et, _ := strconv.Atoi(o.config["EXPIRE_TIME"])
	o.expireTime = et

	if strings.ToUpper(o.config["RECONNECT"]) == "YES" || strings.ToUpper(o.config["RECONNECT"]) == "Y" {
		o.reconnect = true
	}

	signal.Notify(o.interrupt, os.Interrupt)

	return o
}

// stopTicker 关闭定时器
func (o *WebSocket) stopTicker() *WebSocket {
	o.ticker.Stop()
	return o
}

// NewClient 创建新连接
func (o *WebSocket) NewClient() {
	var err error

	o.lock()
	defer o.unLock()

	i := 0
	for {
		if o.expireTime != ExpireAlways && i > o.expireTime {
			break
		}
		o.conn, err = o.createClient()
		if err != nil {
			time.Sleep(time.Second)
		} else {
			o.setPongHandler(o.pongHandler)
			break
		}
		i++
	}

	o.status = WebsocketStatusOnline

}

// createClient 创建连接
func (o *WebSocket) createClient() (*websocket.Conn, error) {
	var err error
	config := o.GetConfig()
	u := url.URL{Scheme: "ws", Host: config["HOST"] + ":" + config["PORT"], Path: config["PATH"]}

	log.Printf("connecting to %s", u.String())

	o.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		o.errs <- err.Error()
		return nil, err
	}
	return o.conn, nil
}

// readMessage 读消息
func (o *WebSocket) readMessage() *WebSocket {
	go func() {
		defer func() {
			close(o.done)
		}()
		for {
			if o.conn != nil {
				messageType, message, err := o.conn.ReadMessage()
				// 如果收到关闭或未知消息，判断是否返回
				if messageType == websocket.CloseMessage || messageType == UnKnowMessageType {
					if o.reconnect == false {
						return
					}
				}
				o.message <- message
				if err != nil {
					time.Sleep(time.Second)
					o.errs <- err.Error()
				}

			}

		}
	}()
	return o
}

// SetMessageHandlerFunc 设置消息处理器
func (o *WebSocket) SetMessageHandlerFunc(fun func(socket *WebSocket, message []byte)) {
	o.messageHandler = fun
}

// GetMessageHandlerFunc 获取当前消息处理器
func (o *WebSocket) GetMessageHandlerFunc() func(socket *WebSocket, message []byte) {
	return o.messageHandler
}

// ExecMessage 获取消息并调用消息处理器
func (o *WebSocket) ExecMessage() {
	go func(o *WebSocket) {
		for {
			select {
			case m, ok := <-o.message:
				if ok {
					o.messageHandler(o, m)
				}
			}
		}
	}(o)
}

// ErrHandler 错误消息处理
func (o *WebSocket) ErrHandler() {
	go func(o *WebSocket) {
		for {
			select {
			case e, ok := <-o.errs:
				if ok {
					log.Printf("error: %v", e)
				}
			}
		}
	}(o)
}

// expireHandler ping <==> pong 有效消息检测
func (o *WebSocket) expireHandler() {
	go func(socket *WebSocket) {
		var i = 0
		for {
			select {
			case e, ok := <-o.expire:
				if ok {
					if e == false {
						i++
						time.Sleep(time.Second)
					} else {
						i = 0
					}
				}
			default:
				if o.expireTime != ExpireAlways {
					if i > o.expireTime {
						o.status = WebsocketStatusOffline
						o.exit <- true
					}
					i++
					time.Sleep(time.Second)
				}
			}
		}
	}(o)
}

// pongHandler pong消息处理器
func (o *WebSocket) pongHandler(appData string) error {
	o.expire <- true
	return nil
}

// ping 发送ping消息
func (o *WebSocket) ping() {
	go func(o *WebSocket) {
		for {
			select {
			case _ = <-o.ticker.C:
				if o.conn != nil {
					err := o.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
					if err != nil {
						// 如果写入出错 将连接状态标识为下线
						// 发送错误消息
						o.status = WebsocketStatusOffline
						o.errs <- err.Error()
					}
				}
			}
		}
	}(o)
}

// interruptHandler 中断/关闭处理器
func (o *WebSocket) interruptHandler() {
	go func(o *WebSocket) {
		for {
			select {
			case <-o.done:
				o.exit <- true
			case <-o.interrupt:
				if o.status == WebsocketStatusOnline {
					o.message <- []byte("interrupt")
					// 发送关闭消息给服务器，等待服务器关闭连接
					err := o.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						o.errs <- errors.New(fmt.Sprintf("write close: %v", err)).Error()
						return
					}
					select {
					case <-o.done:
						o.exit <- true
					case <-time.After(time.Second):
					}
					return
				}
			}
		}
	}(o)
}

// start 健康检查
func (o *WebSocket) start() {
	go func(o *WebSocket) {
		o.NewClient()
		for {
			// 掉线重连
			if o.status == WebsocketStatusOffline {
				if o.reconnect == true {
					o.NewClient()
				} else {
					<-o.done
					return
				}
			}
		}
	}(o)

}

// SetPongHandler 设置Pong消息处理器
func (o *WebSocket) setPongHandler(fun func(appData string) error) {
	o.conn.SetPongHandler(fun)
}

// closeConn 关闭连接
func (o *WebSocket) closeConn() {
	if o.conn != nil {
		o.conn.Close()
	}
}

// WriteMessage 发消息
func (o *WebSocket) WriteMessage(messageType int, message []byte) error {
	err := o.conn.WriteMessage(messageType, message)
	if err != nil {
		o.errs <- err.Error()
	}
	return err
}

// defaultMessageHandlerFunc 默认消息处理器
func defaultMessageHandlerFunc(socket *WebSocket, message []byte) {
	fmt.Printf("%s\n", string(message))
}

func (o *WebSocket) Run() *WebSocket {
	// 初始化
	o.initialize()

	// 健康检查
	o.start()

	// 设置消息默认处理方法
	if o.GetMessageHandlerFunc() == nil {
		o.SetMessageHandlerFunc(defaultMessageHandlerFunc)
	}

	// 发送Ping消息
	o.ping()

	// 断电关闭
	o.interruptHandler()

	// 读取消息
	o.readMessage()

	// 处理普通消息
	o.ExecMessage()

	// 处理错误消息
	o.ErrHandler()

	// ping-pong 有效检测
	o.expireHandler()

	// 关闭连接
	defer o.closeConn()

	// 关闭定时器
	defer o.stopTicker()

	for {
		exit, ok := <-o.exit
		if ok {
			if exit == true {
				fmt.Printf("exit.")
				return o
			}
		}
	}
}
