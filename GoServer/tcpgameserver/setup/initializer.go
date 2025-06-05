package setup

import (
	"GoServer/tcpgameserver/logic"
)

// InitializeServer 初始化TCP服务器的所有必要组件
func InitializeServer() {
	// 1. 初始化事件系统
	initializeEventSystem()
}

// 初始化事件系统
func initializeEventSystem() {
	logic.InitializeEventSystem()
}
