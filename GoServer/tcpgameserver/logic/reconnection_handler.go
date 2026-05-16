package logic

import (
	"encoding/json"
	"net"
	"time"

	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
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
		return r.sendReconnectionFailure(clientID, "Player not found", connManager)
	}

	// 验证玩家状态是否为等待重连
	if originalClientInfo.GetStatus() != types.StatusWaitingReconnect {
		return r.sendReconnectionFailure(clientID, "Player not waiting for reconnection", connManager)
	}

	// 获取玩家的游戏信息
	room, _ := roomManager.FindRoomByPlayer(username)
	playerGameInfo, err := roomManager.GetPlayerGameInfo(room.RoomID, username)
	if err != nil || playerGameInfo == nil {
		return r.sendReconnectionFailure(clientID, "No active game found", connManager)
	}

	// 获取客户端连接信息
	clientInfo, exists := connManager.GetConnectionByClientID(clientID)
	if !exists {
		return r.sendReconnectionFailure(clientID, "Connection not found", connManager)
	}

	// 绑定用户到新连接
	err = connManager.BindUser(clientID, username)
	if err != nil {
		return r.sendReconnectionFailure(clientID, "Failed to bind user", connManager)
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

// sendReconnectionSuccess 发送重连成功消息 (消息类型 6001)
func (r *ReconnectionHandler) sendReconnectionSuccess(conn net.Conn, playerGameInfo *models.PlayerGameInfo) error {
	// 创建重连成功响应
	response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(6001, playerGameInfo)

	// 序列化并发送消息
	messageData, err := json.Marshal(response)
	if err != nil {
		return err
	}
	messageData = append(messageData, '\n')

	_, err = conn.Write(messageData)
	if err != nil {
		return err
	}

	return nil
}

// sendReconnectionFailure 发送重连失败消息
func (r *ReconnectionHandler) sendReconnectionFailure(clientID, reason string, connManager *service.ConnectionManager) error {
	clientInfo, exists := connManager.GetConnectionByClientID(clientID)
	if !exists || clientInfo.Conn == nil {
		return nil // 连接已不存在
	}

	response := tools.GlobalResponseHelper.CreateErrorTcpResponse(6002)

	// 序列化并发送消息
	messageData, err := json.Marshal(response)
	if err != nil {
		return err
	}
	messageData = append(messageData, '\n')

	_, err = clientInfo.Conn.Write(messageData)
	if err != nil {
		return err
	}

	return nil
}

// notifyRoomPlayersReconnection 通知房间内其他玩家有玩家重连
func (r *ReconnectionHandler) notifyRoomPlayersReconnection(username, roomID string, connManager *service.ConnectionManager) error {

	// 获取房间内的所有玩家连接
	allConnections := connManager.GetAllConnections()
	var roomPlayers []*types.ClientInfo

	// 筛选出同一房间内的其他玩家
	for _, clientInfo := range allConnections {
		if clientInfo.GetGameRoom() == roomID && clientInfo.Username != username && clientInfo.Username != "" {
			roomPlayers = append(roomPlayers, clientInfo)
		}
	}

	// 创建重连通知消息 (消息类型 7002)
	reconnectNotification := map[string]interface{}{
		"message_type": "player_reconnect",
		"username":     username,
		"status":       "online",
		"room_id":      roomID,
		"timestamp":    time.Now().Unix(),
	}

	// 向房间内其他玩家发送重连通知
	notifiedCount := 0
	for _, player := range roomPlayers {
		if player.Conn != nil {
			response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(7002, reconnectNotification)

			if messageData, err := json.Marshal(response); err == nil {
				messageData = append(messageData, '\n')
				if _, writeErr := player.Conn.Write(messageData); writeErr != nil {
				} else {
					notifiedCount++
				}
			}
		}
	}
	return nil
}

// 全局重连处理器实例
var GlobalReconnectionHandler = NewReconnectionHandler()
