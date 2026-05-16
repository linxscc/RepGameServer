package logic

import (

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
)

// GameEndProcessor 游戏结束处理器
type GameEndProcessor struct {
	Name string
}

// NewGameEndProcessor 创建新的游戏结束处理器
func NewGameEndProcessor() *GameEndProcessor {
	return &GameEndProcessor{
		Name: "GameEndProcessor",
	}
}

// ProcessGameEnd 处理游戏结束逻辑
func (gep *GameEndProcessor) ProcessGameEnd(data interface{}) error {
	eventData, ok := data.(*events.EventData)
	if !ok {
		return nil
	}

	// 步骤1: 获取房间信息
	roomManager := service.GetRoomManager()
	room, err := roomManager.GetRoom(eventData.RoomID)
	if err != nil {
		return err
	}

	// 步骤2: 为房间内玩家发送各自的玩家信息 (消息码1101)
	err = gep.sendPlayerInfoToAll(room)
	if err != nil {
		return err
	}

	room.Status = "finished"

	// 步骤3: 清理房间信息
	err = gep.cleanupRoomInfo(room)
	if err != nil {
		return err
	}
	// 步骤4: 更新玩家状态信息
	err = gep.updatePlayerStates(room)
	if err != nil {
		return err
	}

	// 步骤5: 删除该房间
	err = gep.deleteRoom(room)
	if err != nil {
		return err
	}

	// 步骤6: 清理全局计时器
	GlobalRoomTimerProcessor.StopRoomTimer(room.RoomID)
	return nil
}

// sendPlayerInfoToAll 为房间内所有玩家发送各自的玩家信息
func (gep *GameEndProcessor) sendPlayerInfoToAll(room *types.RoomInfo) error {
	connManager := service.GetConnectionManager()
	roomManager := service.GetRoomManager()


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

		// 获取该玩家的游戏信息
		playerGameInfo, err := roomManager.GetPlayerGameInfo(room.RoomID, playerName)
		if err != nil {
			failCount++
			continue
		}

		if playerGameInfo == nil {
			failCount++
			continue
		}

		// 创建游戏结束消息 (消息码1101)
		response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(1101, playerGameInfo)

		// 发送消息
		broadcaster := NewGameStateBroadcaster()
		if err := broadcaster.SendTCPMessage(clientInfo.Conn, response); err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	return nil
}

// cleanupRoomInfo 清理房间信息
func (gep *GameEndProcessor) cleanupRoomInfo(room *types.RoomInfo) error {

	// 清理房间状态
	room.UpdateRoomStatus("finished")

	// 清理房间内的卡牌池
	room.Level1CardPool = []models.Card{}
	room.Level2CardPool = []models.Card{}
	room.Level3CardPool = []models.Card{}

	// 重置所有玩家的游戏内状态
	for _, player := range room.Players {
		// 清理手牌
		player.HandCards = []models.Card{}

		// 重置游戏状态
		player.IsReady = false
		player.Round = ""
		player.OtherPlayers = []models.OtherPlayerGameInfo{}
		player.DamageInfo = []models.DamageInfo{}
	}

	return nil
}

// updatePlayerStates 更新玩家状态信息
func (gep *GameEndProcessor) updatePlayerStates(room *types.RoomInfo) error {

	connManager := service.GetConnectionManager()

	for playerName := range room.Players {
		// 获取玩家连接信息
		clientInfo, exists := connManager.GetConnectionByUsername(playerName)
		if !exists {
			continue
		}

		// 更新玩家状态为非游戏状态
		if clientInfo != nil {
			clientInfo.Status = types.StatusLoggedIn
		}
	}

	return nil
}

// deleteRoom 删除该房间
func (gep *GameEndProcessor) deleteRoom(room *types.RoomInfo) error {

	roomManager := service.GetRoomManager()

	// 从房间管理器中删除房间
	err := roomManager.RemoveRoom(room.RoomID)
	if err != nil {
		return err
	}

	return nil
}
