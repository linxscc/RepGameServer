package tcpserver

import (
	"encoding/json"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// HandleUserPlayCard 处理用户出牌请求（仅负责数据解析验证和事件发布）
func HandleUserPlayCard(req models.TcpRequest, conn net.Conn, clientID string, connManager *service.ConnectionManager) {
	// 获取客户端信息
	clientInfo, _ := connManager.GetConnectionByClientID(clientID)

	// 解析出牌请求数据 - 使用PlayerGameInfo
	var playCardData models.PlayerGameInfo
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4002)) // 数据格式错误
		return
	}
	if err := json.Unmarshal(dataBytes, &playCardData); err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4002)) // 数据格式错误
		return
	}

	// 基本数据验证
	if len(playCardData.SelfCards) == 0 {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(5009)) // 没有提供出牌信息
		return
	}

	// 发布出牌事件，让事件系统处理实际的游戏逻辑和验证
	playCardEventData := events.NewEventData(events.EventCardPlay, "user_play_card_handler", map[string]interface{}{
		"player":     clientInfo.Username,
		"self_cards": playCardData.SelfCards,
		"room_id":    playCardData.RoomId,
		"client_id":  clientID,
		"connection": conn,
	})
	events.Publish(events.EventCardPlay, playCardEventData)
}
