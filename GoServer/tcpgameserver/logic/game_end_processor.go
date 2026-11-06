package logic

import (
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// finishGame 对局结束：重置玩家状态并删除房间（MatchOver 已由 chessRoom.finish 推送）。
// 调用方必须持有房间操作锁（见 room_sync.go）。
func finishGame(room *types.RoomInfo) {
	roomID := room.RoomID

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
