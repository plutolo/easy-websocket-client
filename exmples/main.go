// author: luokuanxing <346300265@qq.com>
// date: 2022/7/15

package main

import (
	websocket "github.com/plutolo/easy-websocket-client"
)

func GetConfig() map[string]string {
	config := make(map[string]string)
	config["HOST"] = "localhost"
	config["PORT"] = "8088"
	config["PATH"] = "/echo"
	config["RECONNECT"] = "YES"
	config["EXPIRE_TIME"] = "0" // 重连时间（秒），0为一直重连

	return config

}

func main() {
	// 创建webSocket实例P
	webSocket := websocket.NewWebsocket()

	// 设置
	webSocket.SetConfig(GetConfig())
	webSocket.SetMessageHandlerFunc(func(socket *websocket.WebSocket, message []byte) {

	})
	// 运行
	webSocket.Run()

}
