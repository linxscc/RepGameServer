package logic

import (
	"log"

	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// 对局状态推送。调用方必须持有房间操作锁（见 room_sync.go），
// 保证推送的是一次操作完成后的完整状态。

// broadcastRoomState 向房间内每个在线的真人玩家推送各自视角的状态，然后清空本次的伤害信息
func broadcastRoomState(room *types.RoomInfo, reason pb.GameStateReason) {
	for _, name := range room.PlayerNames() {
		sendRoomState(room, name, reason)
	}
	for _, name := range room.PlayerNames() {
		_ = room.CleanPlayerDamage(name)
	}
}

// sendRoomState 向一个玩家推送其视角的状态；机器人或不在线的玩家跳过
func sendRoomState(room *types.RoomInfo, username string, reason pb.GameStateReason) {
	if IsBot(username) {
		return
	}
	info, ok := service.GetConnectionManager().GetConnectionByUsername(username)
	if !ok {
		return
	}
	conn := info.GetConn()
	if conn == nil {
		return // 断线中：重连时会收到完整状态
	}
	snapshot, err := room.Snapshot(username)
	if err != nil {
		log.Printf("room %s: snapshot for %s: %v", room.RoomID, username, err)
		return
	}
	_ = sendTo(conn, protocol.GameStateMessage(snapshot, reason))
}
