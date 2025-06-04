package tcpserver

import (
	"log"
	"net"

	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
)

// HandleUserReady 处理用户准备
func HandleUserReady(conn net.Conn, clientID string, connManager *service.ConnectionManager) {
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

	// 获取准备就绪的玩家统计
	readyPlayers := connManager.GetConnectionsByStatus(types.StatusReady)
	stats := connManager.GetConnectionStats()

	// 匹配逻辑（这里可以根据需要实现具体的匹配逻辑）
	if stats["ready"] >= 2 {
		log.Printf("有足够的玩家准备就绪，可以开始匹配")
		// 这里可以调用匹配逻辑
		// logic.StartMatchMaking()
	}

	// 构建响应数据
	var readyPlayerNames []string
	for _, player := range readyPlayers {
		if player.Username != "" {
			readyPlayerNames = append(readyPlayerNames, player.Username)
		}
	}

	responseData := map[string]interface{}{
		"ready_players": readyPlayerNames,
		"player_count":  stats["ready"],
	}
	SendTCPResponse(conn, tools.GlobalResponseHelper.CreateSuccessTcpResponse(4001, responseData))
}
