package logic

import (
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// finishGame 对局结束：推送最终状态、重置玩家状态并删除房间。
// 调用方必须持有房间操作锁（见 room_sync.go）。
func finishGame(room *types.RoomInfo) {
	roomID := room.RoomID
	GlobalRoomTimerProcessor.stopRoomTimer(roomID)

	// 最终状态（含致命一击的伤害信息）
	for _, name := range room.PlayerNames() {
		sendRoomState(room, name, pb.GameStateReason_GAME_STATE_REASON_GAME_OVER)
	}

	connManager := service.GetConnectionManager()
	for _, name := range room.PlayerNames() {
		info, ok := connManager.GetConnectionByUsername(name)
		if !ok {
			continue
		}
		if info.GetStatus() == types.StatusWaitingReconnect {
			// 断线未归的玩家没有对局可恢复了，释放其座位记录
			cancelReconnectGrace(name)
			connManager.RemoveConnection(info.ClientID)
			continue
		}
		info.SetStatus(types.StatusLoggedIn)
		info.SetGameRoom("")
	}

	room.ResetForGameEnd()
	_ = service.GetRoomManager().RemoveRoom(roomID)
	forgetRoomLock(roomID)
}
