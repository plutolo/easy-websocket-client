// author: luokuanxing <346300265@qq.com>
// date: 2022/7/16

package main

func GetConfig() map[string]string {
	config := make(map[string]string)
	config["HOST"] = "localhost"
	config["PORT"] = "8088"
	config["PATH"] = "/echo"
	config["RECONNECT"] = "YES"
	config["EXPIRE_TIME"] = "0" // 重连时间（秒），0为一直重连

	return config

}
