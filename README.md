# easy-websocket-client

一套基于 gorilla/websocket 封装的客户端应用

- 封装了 自动重连、有效性检测, 陆续添加新功能

## 安装步骤

### 安装govendor包管理工具

```
# go get -u -v github.com/plutolo/easy-websocket-client
```

### 使用 govendor 安装依赖包

```sh
# govendor sync
```

### 服务配置

```go
package main

// 服务配置 防止变量污染故用函数组织
func GetConfig() map[string]string {
	config := make(map[string]string)
	config["HOST"] = "localhost" // 服务器地址
	config["PORT"] = "8088"      // 服务器端口
	config["PATH"] = "/echo"     // 路由
	config["RECONNECT"] = "YES"  // 自动重连
	config["EXPIRE_TIME"] = "0"  // 重连时间（秒），0为一直重连

	return config

}

```

### 启动服务示例

```go
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
```
