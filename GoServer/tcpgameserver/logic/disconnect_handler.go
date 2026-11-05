package logic

import (
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// DisconnectHandler 断线处理器
type DisconnectHandler struct{}

// NewDisconnectHandler 创建断线处理器
func NewDisconnectHandler() *DisconnectHandler {
	return &DisconnectHandler{}
}

// HandlePlayerDisconnect 处理玩家断开连接逻辑
func (d *DisconnectHandler) HandlePlayerDisconnect(clientID, username, reason string) error {

	// 处理已登录用户的断开连接
	if username != "" {
		return d.handleLoggedInUserDisconnect(clientID, username, reason)
	}

	// 处理未登录用户的断开连接
	return d.handleAnonymousUserDisconnect(clientID)
}

// handleLoggedInUserDisconnect 处理已登录用户断开连接
func (d *DisconnectHandler) handleLoggedInUserDisconnect(clientID, username, reason string) error {
	// 获取连接管理器和房间管理器
	connManager := service.GetConnectionManager()
	roomManager := service.GetRoomManager()

	// 检查玩家是否在游戏房间中
	clientInfo, exists := connManager.GetConnectionByClientID(clientID)
	if !exists {
		// 连接信息不存在，直接移除
		connManager.RemoveConnection(clientID)
		return nil
	}

	roomID := clientInfo.GetGameRoom()
	status := clientInfo.GetStatus()
	// 如果玩家在游戏中，设置为等待重连状态
	if status == types.StatusInGame && roomID != "" {
		clientInfo.SetStatus(types.StatusWaitingReconnect)
		return d.handleInGamePlayerDisconnect(clientID, username, reason, roomID, roomManager)
	}

	// 玩家不在游戏中，完全移除连接
	return d.handleNonGamePlayerDisconnect(clientID, username, reason, roomID, connManager)
}

// handleInGamePlayerDisconnect 处理游戏中玩家断开连接
func (d *DisconnectHandler) handleInGamePlayerDisconnect(clientID, username, reason, roomID string, roomManager *service.RoomManager) error {
	notifyRoomPeers(roomID, username, presenceMessage(username, roomID, pb.PresenceStatus_PRESENCE_STATUS_DISCONNECTED, reason))
	return nil
}

// handleNonGamePlayerDisconnect 处理非游戏状态玩家断开连接
func (d *DisconnectHandler) handleNonGamePlayerDisconnect(clientID, username, reason, roomID string, connManager *service.ConnectionManager) error {
	// 玩家不在游戏中，完全移除连接
	connManager.RemoveConnection(clientID)

	return nil
}

// handleAnonymousUserDisconnect 处理未登录用户断开连接
func (d *DisconnectHandler) handleAnonymousUserDisconnect(clientID string) error {
	// 未登录用户直接移除连接
	connManager := service.GetConnectionManager()
	connManager.RemoveConnection(clientID)
	return nil
}

// 全局断线处理器实例
var GlobalDisconnectHandler = NewDisconnectHandler()
