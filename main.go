// author: luokuanxing <346300265@qq.com>
// date: 2022/7/15

package main

func main() {
	// 创建webSocket实例P
	webSocket := NewWebsocket()

	// 设置
	webSocket.SetConfig(GetConfig())
	webSocket.SetMessageHandlerFunc(func(socket *WebSocket, message []byte) {

	})
	// 运行
	webSocket.Run()

}
