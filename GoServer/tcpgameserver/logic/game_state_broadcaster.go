package logic

import (
	"encoding/json"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
)

// GameStateBroadcaster 游戏状态广播器
type GameStateBroadcaster struct{}

// NewGameStateBroadcaster 创建新的游戏状态广播器
func NewGameStateBroadcaster() *GameStateBroadcaster {
	return &GameStateBroadcaster{}
}

// SendTCPMessage 发送TCP消息到连接
func (gsb *GameStateBroadcaster) SendTCPMessage(conn net.Conn, response *models.TcpResponse) error {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	// 添加换行符作为消息结束标识
	jsonBytes = append(jsonBytes, '\n')

	_, err = conn.Write(jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

// BroadcastGameStateToRoom 向房间内所有玩家广播游戏状态（统一方法）
func (gsb *GameStateBroadcaster) BroadcastGameStateToRoom(eventData *events.EventData) {

	// 获取连接管理器
	connManager := service.GetConnectionManager()
	roomManager := service.GetRoomManager()

	// 获取房间信息
	room, err := roomManager.GetRoom(eventData.RoomID)
	if err != nil {
		return
	}

	// 向房间内每个玩家发送其个人游戏信息
	successCount := 0
	failCount := 0

	for playerName := range room.Players {
		// 获取该玩家的连接信息
		clientInfo, exists := connManager.GetConnectionByUsername(playerName)
		if !exists {
			failCount++
			continue
		}

		if clientInfo.Conn == nil {
			failCount++
			continue
		}

		// 从房间管理器获取该玩家的游戏信息
		playerGameInfo, err := roomManager.GetPlayerGameInfo(eventData.RoomID, playerName)
		if err != nil {
			failCount++
			continue
		}
		// 如果玩家不在游戏中，跳过
		if playerGameInfo == nil {
			continue
		}

		// 根据事件源确定消息码
		var messageCode int
		switch eventData.Source {
		case "card_compose_processor":
			messageCode = 9001
		case "play_card_processor":
			messageCode = 8001
		case "force_cardplay_processor":
			messageCode = 7001
		default:
			messageCode = 8001 // 默认消息码
		}

		// 创建游戏状态响应消息
		response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(messageCode, playerGameInfo)
		// 发送消息
		if err := gsb.SendTCPMessage(clientInfo.Conn, response); err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	// 在广播完成后重置所有玩家的战斗统计信息
	gsb.resetPlayerBattleStats(room)

}

// resetPlayerBattleStats 重置房间内所有玩家的战斗统计信息
func (gsb *GameStateBroadcaster) resetPlayerBattleStats(room *types.RoomInfo) {
	roomManager := service.GetRoomManager()
	for playerName := range room.Players {
		// 重置玩家的伤害统计信息
		err := roomManager.CleanPlayerDamage(room.RoomID, playerName)
		if err != nil {
			continue
		}
	}
}
