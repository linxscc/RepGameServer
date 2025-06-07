package logic

import (
	"encoding/json"
	"log"
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
		log.Printf("Failed to marshal response: %v", err)
		return err
	}

	// 添加换行符作为消息结束标识
	jsonBytes = append(jsonBytes, '\n')

	_, err = conn.Write(jsonBytes)
	if err != nil {
		log.Printf("Failed to write response to connection: %v", err)
		return err
	}

	return nil
}

// BroadcastGameStateToRoom 向房间内所有玩家广播游戏状态（统一方法）
func (gsb *GameStateBroadcaster) BroadcastGameStateToRoom(room *types.RoomInfo, eventData *events.EventData, connManager *service.ConnectionManager) {
	log.Printf("Broadcasting game state to room %s with %d players", room.RoomID, len(room.Players))

	// 获取房间管理器
	roomManager := service.GetRoomManager()

	// 向房间内每个玩家发送其个人游戏信息
	successCount := 0
	failCount := 0

	for playerName := range room.Players {
		// 获取该玩家的连接信息
		clientInfo, exists := connManager.GetConnectionByUsername(playerName)
		if !exists {
			log.Printf("Player %s not found in connection manager", playerName)
			failCount++
			continue
		}

		if clientInfo.Conn == nil {
			log.Printf("Connection is nil for player %s", playerName)
			failCount++
			continue
		}

		// 从房间管理器获取该玩家的游戏信息
		playerGameInfo, err := roomManager.GetPlayerGameInfo(playerName)
		if err != nil {
			log.Printf("Failed to get game info for player %s: %v", playerName, err)
			failCount++
			continue
		}
		// 如果玩家不在游戏中，跳过
		if playerGameInfo == nil {
			log.Printf("Player %s is not in game state", playerName)
			continue
		}

		// 根据事件源确定消息码
		var messageCode int
		if eventData.Source == "card_compose_processor" {
			messageCode = 9001
		} else if eventData.Source == "play_card_processor" {
			messageCode = 8001
		} else {
			messageCode = 8001 // 默认消息码
		}

		// 创建游戏状态响应消息
		response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(messageCode, playerGameInfo)
		// 发送消息
		if err := gsb.SendTCPMessage(clientInfo.Conn, response); err != nil {
			log.Printf("Failed to send game state to player %s: %v", playerName, err)
			failCount++
		} else {
			log.Printf("Successfully sent game state to player %s", playerName)
			successCount++
		}
	}

	// 在广播完成后重置所有玩家的战斗统计信息
	gsb.resetPlayerBattleStats(room)

	log.Printf("Game state broadcast completed for room %s: %d success, %d failed",
		room.RoomID, successCount, failCount)
}

// resetPlayerBattleStats 重置房间内所有玩家的战斗统计信息
func (gsb *GameStateBroadcaster) resetPlayerBattleStats(room *types.RoomInfo) {
	log.Printf("Resetting battle stats for all players in room %s", room.RoomID)

	for playerName := range room.Players {
		// 重置玩家的伤害统计信息
		err := room.SetPlayerDamageStats(playerName, 0, 0)
		if err != nil {
			log.Printf("Failed to reset damage stats for player %s: %v", playerName, err)
			continue
		}

		// 重置玩家的触发羁绊列表
		err = room.SetPlayerBonds(playerName, []models.BondModel{})
		if err != nil {
			log.Printf("Failed to reset bonds for player %s: %v", playerName, err)
			continue
		}

		log.Printf("Reset battle stats for player %s", playerName)
	}

	log.Printf("Completed resetting battle stats for room %s", room.RoomID)
}
