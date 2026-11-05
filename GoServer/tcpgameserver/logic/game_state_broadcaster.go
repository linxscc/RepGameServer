package logic

import (
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// GameStateBroadcaster 游戏状态广播器
type GameStateBroadcaster struct{}

// NewGameStateBroadcaster 创建新的游戏状态广播器
func NewGameStateBroadcaster() *GameStateBroadcaster {
	return &GameStateBroadcaster{}
}

// SendTCPMessage 发送消息到连接
func (gsb *GameStateBroadcaster) SendTCPMessage(conn net.Conn, msg *pb.ServerMessage) error {
	return sendTo(conn, msg)
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

		response := protocol.GameStateMessage(playerGameInfo, reasonForSource(eventData.Source))
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

// reasonForSource 根据触发广播的业务来源确定推送原因
func reasonForSource(source string) pb.GameStateReason {
	switch source {
	case "card_compose_processor":
		return pb.GameStateReason_GAME_STATE_REASON_CARDS_COMPOSED
	case "force_cardplay_processor":
		return pb.GameStateReason_GAME_STATE_REASON_FORCED_PLAY
	default:
		return pb.GameStateReason_GAME_STATE_REASON_CARDS_PLAYED
	}
}
