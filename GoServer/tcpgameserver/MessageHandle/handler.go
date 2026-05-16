package tcpserver

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// SendTCPResponse 发送TCP响应消息
func SendTCPResponse(conn net.Conn, resp *models.TcpResponse) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	_, err = conn.Write(jsonBytes)
	if err != nil {
	}
}

// HandleTCPMessage 处理TCP消息
func HandleTCPMessage(msg string, conn net.Conn, clientID string) {
	// 更新客户端活动时间
	connManager := service.GetConnectionManager()
	connManager.UpdateActivity(clientID)
	msg = strings.TrimSpace(msg)
	// 尝试将msg解析为TcpRequest结构体
	var req models.TcpRequest
	err := json.Unmarshal([]byte(msg), &req)
	if err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
		return
	}
	switch req.Message {
	case "UserLogin":
		HandleUserLogin(req, conn, clientID, connManager)
	case "UserRegister":
		HandleUserRegister(req, conn, clientID, connManager)
	case "UserReady":
		HandleUserReady(conn, clientID, connManager)
	case "UserPlayCard":
		HandleUserPlayCard(req, conn, clientID, connManager)
	case "UserComposeCard":
		HandleUserComposeCard(req, conn, clientID, connManager)
	case "UserRestart":
		HandleUserRestart(conn, clientID, connManager)
	default:
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(9999))
	}
}

// HandleNewConnection 处理新的TCP连接
func HandleNewConnection(conn net.Conn) string {
	remoteAddr := conn.RemoteAddr().String()

	// 生成客户端ID（包含时间戳和地址，确保唯一性）
	clientID := fmt.Sprintf("client_%d_%s", time.Now().UnixNano(),
		strings.ReplaceAll(remoteAddr, ":", "_"))

	// 获取连接管理器
	connManager := service.GetConnectionManager()

	// 检查是否是重连的客户端（基于地址）
	if existingInfo, exists := connManager.GetConnectionByAddr(remoteAddr); exists {

		// 移除旧连接，为新连接让路
		connManager.RemoveConnection(existingInfo.ClientID)
	}
	// 添加到连接管理器
	connManager.AddConnection(conn, clientID)

	// 发布客户端连接事件
	connectData := events.CreateConnectionEventData(events.EventClientConnect, clientID, remoteAddr)
	connectData.AddData("connection_type", "tcp")
	connectData.AddData("user_agent", "game_client")
	connectData.AddData("version", "1.0")
	connectData.AddData("first_connect_time", time.Now().Unix())
	events.Publish(events.EventClientConnect, connectData)

	return clientID
}

// HandleConnectionClose 处理连接关闭
func HandleConnectionClose(clientID string) {
	connManager := service.GetConnectionManager()

	// 获取连接信息用于事件发布
	if clientInfo, exists := connManager.GetConnectionByClientID(clientID); exists {
		// 发布客户端断开连接事件
		disconnectData := events.CreateUserConnectionEventData(
			events.EventClientDisconnect, clientID, clientInfo.Username, clientInfo.RemoteAddr)
		disconnectData.AddData("reason", "connection_close")
		events.Publish(events.EventClientDisconnect, disconnectData)
	}

	if connManager.RemoveConnection(clientID) {
	}
}

// UpdateClientActivity 更新客户端活动时间
func UpdateClientActivity(clientID string) {
	connManager := service.GetConnectionManager()
	connManager.UpdateActivity(clientID)
}
