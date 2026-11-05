package logic

import (
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"net"

	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// ReconnectionHandler 重连处理器
type ReconnectionHandler struct{}

// NewReconnectionHandler 创建重连处理器
func NewReconnectionHandler() *ReconnectionHandler {
	return &ReconnectionHandler{}
}

// HandlePlayerReconnection 处理玩家重连逻辑
func (r *ReconnectionHandler) HandlePlayerReconnection(clientID, username string) error {
	// 获取服务管理器
	connManager := service.GetConnectionManager()
	roomManager := service.GetRoomManager()

	// 检查玩家原本的状态是否为等待重连
	originalClientInfo, exists := connManager.GetConnectionByUsername(username)
	if !exists {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectPlayerNotFound), connManager)
	}

	// 验证玩家状态是否为等待重连
	if originalClientInfo.GetStatus() != types.StatusWaitingReconnect {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectNotWaiting), connManager)
	}

	// 获取玩家的游戏信息
	room, err := roomManager.FindRoomByPlayer(username)
	if err != nil || room == nil {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectNoGame), connManager)
	}
	playerGameInfo, err := roomManager.GetPlayerGameInfo(room.RoomID, username)
	if err != nil || playerGameInfo == nil {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectNoGame), connManager)
	}

	// 获取客户端连接信息
	clientInfo, exists := connManager.GetConnectionByClientID(clientID)
	if !exists {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectConnectionLost), connManager)
	}

	// 绑定用户到新连接
	err = connManager.BindUser(clientID, username)
	if err != nil {
		return r.sendReconnectionFailure(clientID, config.Msg(config.MsgReconnectBindFailed), connManager)
	}

	// 设置玩家状态为游戏中
	connManager.SetPlayerStatus(clientID, types.StatusInGame)
	connManager.SetPlayerGameRoom(clientID, playerGameInfo.RoomId)
	// 发送重连成功消息 (消息类型 6001)
	err = r.sendReconnectionSuccess(clientInfo.Conn, playerGameInfo)
	if err != nil {
		return err
	}

	// 通知房间内其他玩家该玩家已重连
	err = r.notifyRoomPlayersReconnection(username, playerGameInfo.RoomId, connManager)
	if err != nil {
	}

	return nil
}

// sendReconnectionSuccess 推送重连后的完整对局状态
func (r *ReconnectionHandler) sendReconnectionSuccess(conn net.Conn, playerGameInfo *models.PlayerGameInfo) error {
	return sendTo(conn, protocol.GameStateMessage(playerGameInfo, pb.GameStateReason_GAME_STATE_REASON_RECONNECTED))
}

// sendReconnectionFailure 告知客户端无法恢复对局
func (r *ReconnectionHandler) sendReconnectionFailure(clientID, reason string, connManager *service.ConnectionManager) error {
	notifyClientError(clientID, 0, pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE, reason)
	return nil
}

// notifyRoomPlayersReconnection 通知房间内其他玩家有玩家重连
func (r *ReconnectionHandler) notifyRoomPlayersReconnection(username, roomID string, connManager *service.ConnectionManager) error {
	notifyRoomPeers(roomID, username, presenceMessage(username, roomID, pb.PresenceStatus_PRESENCE_STATUS_RECONNECTED, ""))
	return nil
}

// 全局重连处理器实例
var GlobalReconnectionHandler = NewReconnectionHandler()
