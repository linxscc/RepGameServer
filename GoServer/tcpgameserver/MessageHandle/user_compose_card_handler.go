package tcpserver

import (
	"encoding/json"
	"log"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// HandleUserComposeCard 处理用户合成卡牌请求
func HandleUserComposeCard(req models.TcpRequest, conn net.Conn, clientID string, connManager *service.ConnectionManager) {
	log.Printf("Handling user compose card request from client %s", clientID)

	// 获取客户端信息
	clientInfo, _ := connManager.GetConnectionByClientID(clientID)

	// 解析PlayerGameInfo数据
	var gameInfo models.PlayerGameInfo
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		log.Printf("Failed to marshal request data: %v", err)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
		return
	}

	err = json.Unmarshal(dataBytes, &gameInfo)
	if err != nil {
		log.Printf("Failed to parse PlayerGameInfo for compose card: %v", err)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
		return
	}

	// 验证必要字段
	if gameInfo.RoomId == "" {
		log.Printf("Missing room_id in compose card request")
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2001))
		return
	}

	if len(gameInfo.SelfCards) == 0 {
		log.Printf("No cards provided for composition")
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2002))
		return
	}

	// 验证卡牌数量（必须是3的倍数）
	if len(gameInfo.SelfCards)%3 != 0 {
		log.Printf("Invalid card count for composition: %d (must be multiple of 3)", len(gameInfo.SelfCards))
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2003))
		return
	}

	// 发布卡牌合成事件
	composeEventData := events.NewEventData(events.EventCardCompose, "user_compose_card_handler", map[string]interface{}{
		"room_id":   gameInfo.RoomId,
		"player":    clientInfo.Username,
		"cards":     gameInfo.SelfCards,
		"client_id": clientID,
	})

	events.Publish(events.EventCardCompose, composeEventData)

}
