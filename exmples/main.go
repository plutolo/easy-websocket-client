// author: luokuanxing <346300265@qq.com>
// date: 2022/7/15

package main

import websocket "websocket-client"

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
