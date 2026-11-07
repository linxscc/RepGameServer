package match

import (
	"errors"
	"log"
	"sync"
	"time"

	"GoServer/tcpgameserver/data/config"
	"GoServer/tcpgameserver/model"
	"GoServer/tcpgameserver/network/transport"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
)

// 玩家在线状态：断线、重连、顶号
// ------------------------------------------------------------
// 对局中断线的玩家保留座位记录（StatusWaitingReconnect，连接为 nil），
// 在 network.yaml 的 reconnect_grace_seconds 内用会话令牌或密码重新登录即可回到对局；
// 期间对局照常进行，各阶段到时由计时器推进。宽限期过后记录被清理，不再能恢复。
// 不在对局中的玩家断线时直接清理记录。
//
// 这些函数都在连接协程内同步调用，保证"标记断线"一定发生在"重连恢复"之前。

var (
	graceMu     sync.Mutex
	graceTimers = make(map[string]*time.Timer) // username -> 宽限计时器
)

// PlayerDisconnected 连接关闭时调用（读循环退出、被顶号）。
func PlayerDisconnected(clientID, reason string) {
	connManager := service.GetConnectionManager()
	info, ok := connManager.GetConnectionByClientID(clientID)
	if !ok {
		return // 已被新连接替换（顶号/重连）或已清理
	}

	switch info.GetStatus() {
	case model.StatusWaitingReconnect:
		return // 已处理过（顶号时先标记，随后旧连接的读循环退出又会调用一次）
	case model.StatusInGame:
		username, roomID := info.GetUsername(), info.GetGameRoom()
		if username != "" && roomID != "" && roomHasPlayer(roomID, username) {
			connManager.MarkWaitingReconnect(clientID)
			startReconnectGrace(username, clientID)
			notifyRoomPeers(roomID, username, presenceMessage(username, roomID, pb.PresenceStatus_PRESENCE_STATUS_DISCONNECTED, reason))
			chessPresenceChanged(roomID)
			log.Printf("player %s disconnected from room %s (%s), waiting for reconnect", username, roomID, reason)
			return
		}
	}
	connManager.RemoveConnection(clientID)
}

// KickDuplicateLogin 同一账号在新连接上登录：通知并断开旧连接。
// 旧连接若在对局中，会先转为等待重连，随后由新连接恢复对局。
func KickDuplicateLogin(old *model.ClientInfo) {
	conn := old.GetConn()
	if conn != nil {
		// 先入队通知，随后的关闭会等它发出（见 transport.CloseAfterFlush）
		_ = sendTo(conn, transport.NewError(0, pb.ErrorCode_ERROR_CODE_SESSION_EXPIRED, config.Msg(config.MsgLoggedInElsewhere)))
	}
	PlayerDisconnected(old.ClientID, "logged_in_elsewhere")
	if conn != nil {
		transport.CloseAfterFlush(conn)
	}
}

// CanResumeGame 用户是否有一局可以恢复（断线后仍在宽限期内）
func CanResumeGame(username string) bool {
	info, ok := service.GetConnectionManager().GetConnectionByUsername(username)
	return ok && info.GetStatus() == model.StatusWaitingReconnect
}

// ResumeGame 把等待重连的玩家绑定到新连接，并推送完整对局状态。
// 返回 false 表示没有可恢复的对局（调用方按普通登录处理）。
func ResumeGame(clientID, username string) bool {
	connManager := service.GetConnectionManager()
	old, ok := connManager.GetConnectionByUsername(username)
	if !ok || old.GetStatus() != model.StatusWaitingReconnect {
		return false
	}
	roomID := old.GetGameRoom()

	err := withRoom(roomID, func(room *model.RoomInfo) error {
		if !room.HasPlayer(username) || room.GetStatus() != model.RoomStatusPlaying {
			return errNoGameToResume
		}
		cancelReconnectGrace(username)
		// BindUser 会移除旧的座位记录，并把用户名映射到新连接
		if err := connManager.BindUser(clientID, username); err != nil {
			return err
		}
		_ = connManager.SetPlayerStatus(clientID, model.StatusInGame)
		_ = connManager.SetPlayerGameRoom(clientID, roomID)

		chessResend(roomID, username)
		notifyRoomPeers(roomID, username, presenceMessage(username, roomID, pb.PresenceStatus_PRESENCE_STATUS_RECONNECTED, ""))
		return nil
	})
	if err != nil {
		log.Printf("resume game for %s in room %s: %v", username, roomID, err)
		return false
	}
	log.Printf("player %s reconnected to room %s", username, roomID)
	return true
}

var errNoGameToResume = errors.New("no game to resume")

func roomHasPlayer(roomID, username string) bool {
	room, err := service.GetRoomManager().GetRoom(roomID)
	return err == nil && room.HasPlayer(username) && room.GetStatus() == model.RoomStatusPlaying
}

// startReconnectGrace 宽限期结束仍未重连：清理座位记录（对局继续，由阶段计时器推进）
func startReconnectGrace(username, clientID string) {
	graceMu.Lock()
	defer graceMu.Unlock()
	if old, ok := graceTimers[username]; ok {
		old.Stop()
	}
	var timer *time.Timer
	timer = time.AfterFunc(config.Game().Network.ReconnectGrace(), func() {
		graceMu.Lock()
		if graceTimers[username] == timer {
			delete(graceTimers, username)
		}
		graceMu.Unlock()

		connManager := service.GetConnectionManager()
		if info, ok := connManager.GetConnectionByClientID(clientID); ok && info.GetStatus() == model.StatusWaitingReconnect {
			connManager.RemoveConnection(clientID)
			log.Printf("player %s did not reconnect in time", username)
		}
	})
	graceTimers[username] = timer
}

// cancelReconnectGrace 玩家已重连或对局已结束
func cancelReconnectGrace(username string) {
	graceMu.Lock()
	defer graceMu.Unlock()
	if t, ok := graceTimers[username]; ok {
		t.Stop()
		delete(graceTimers, username)
	}
}
