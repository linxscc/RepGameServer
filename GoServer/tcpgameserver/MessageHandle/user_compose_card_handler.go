package tcpserver

import (
	"encoding/json"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// HandleUserComposeCard 处理用户合成卡牌请求
func HandleUserComposeCard(req models.TcpRequest, conn net.Conn, clientID string, connManager *service.ConnectionManager) {

	// 获取客户端信息
	clientInfo, _ := connManager.GetConnectionByClientID(clientID)

	// 解析PlayerGameInfo数据
	var gameInfo models.PlayerGameInfo
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
		return
	}

	err = json.Unmarshal(dataBytes, &gameInfo)
	if err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
		return
	}

	// 验证必要字段
	if gameInfo.RoomId == "" {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2001))
		return
	}

	if len(gameInfo.SelfCards) == 0 {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2002))
		return
	}

	// 验证卡牌数量（必须是3的倍数）
	if len(gameInfo.SelfCards)%3 != 0 {
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
