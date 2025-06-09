package tcpserver

import (
	"log"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
)

// HandleUserReady 处理用户准备
func HandleUserRestart(conn net.Conn, clientID string, connManager *service.ConnectionManager) {
	// 获取客户端信息
	clientInfo, exists := connManager.GetConnectionByClientID(clientID)
	if !exists {
		log.Printf("Client not found: %s", clientID)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4003))
		return
	}

	// 检查用户是否已登录
	if !clientInfo.IsLoggedIn {
		log.Printf("User not logged in for client: %s", clientID)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4002))
		return
	}

	// 设置玩家状态为准备就绪
	connManager.SetPlayerStatus(clientID, types.StatusReady)
	stats := connManager.GetConnectionStats() // 匹配逻辑 - 当有足够玩家准备就绪时发送游戏开始事件
	if stats["ready"] >= 2 {

		// 创建游戏开始事件数据
		gameStartData := events.CreateRoomEventData(events.EventGameStart, "new_room", int(stats["ready"]))
		gameStartData.AddData("message", "Players ready, starting matchmaking")
		gameStartData.AddData("trigger_source", "user_ready_handler")

		// 发布游戏开始事件，让事件监听器处理匹配逻辑
		events.Publish(events.EventGameStart, gameStartData)

	}
}
