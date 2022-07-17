// author: luokuanxing <346300265@qq.com>
// date: 2022/7/15

package main

import (
	"github.com/golang/glog"
	websocket "websocket-client"
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

func messageHandler(webSocket *websocket.WebSocket, message []byte) {
	if string(message) != "连接成功" {
		glog.Infof("收到消息: %s，正在处理...", message)
		//
	} else {
		glog.Infof("连接成功")
	}
}

func main() {
	// 创建webSocket实例P
	webSocket := websocket.NewWebsocket()

	webSocket.SetConfig(GetConfig())

	webSocket.SetMessageHandler(messageHandler)

	webSocket.Run()

}
