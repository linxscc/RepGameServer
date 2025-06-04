package tcpserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// SendTCPResponse 发送TCP响应消息
func SendTCPResponse(conn net.Conn, resp *models.TcpResponse) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	_, err = conn.Write(jsonBytes)
	if err != nil {
		log.Printf("Failed to write response to connection: %v", err)
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
		log.Printf("Failed to parse TcpRequest: %v, raw: %s", err, msg)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4001))
		return
	}
	switch req.Message {
	case "UserLogin":
		HandleUserLogin(req, conn, clientID, connManager)
	case "UserRegister":
		HandleUserRegister(req, conn, clientID, connManager)
	case "UserReady":
		HandleUserReady(conn, clientID, connManager)
	default:
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(4001))
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
		log.Printf("Detected potential reconnection from %s, old clientID: %s, new clientID: %s",
			remoteAddr, existingInfo.ClientID, clientID)

		// 移除旧连接，为新连接让路
		connManager.RemoveConnection(existingInfo.ClientID)
	}

	// 添加到连接管理器
	clientInfo := connManager.AddConnection(conn, clientID)

	// 设置初始元数据
	clientInfo.SetMetadata("connection_type", "tcp")
	clientInfo.SetMetadata("user_agent", "game_client")
	clientInfo.SetMetadata("version", "1.0")
	clientInfo.SetMetadata("first_connect_time", time.Now().Unix())

	log.Printf("New client connected: clientID=%s, addr=%s, status=%s",
		clientID, clientInfo.RemoteAddr, clientInfo.GetStatus())

	// 发送欢迎消息
	welcomeResponse := tools.GlobalResponseHelper.CreateSuccessTcpResponse(1001, map[string]interface{}{
		"client_id":   clientID,
		"server_time": time.Now().Unix(),
		"status":      "connected",
	})
	SendTCPResponse(conn, welcomeResponse)

	return clientID
}

// HandleConnectionClose 处理连接关闭
func HandleConnectionClose(clientID string) {
	connManager := service.GetConnectionManager()
	if connManager.RemoveConnection(clientID) {
		log.Printf("Client disconnected: %s", clientID)
	}
}

// UpdateClientActivity 更新客户端活动时间
func UpdateClientActivity(clientID string) {
	connManager := service.GetConnectionManager()
	connManager.UpdateActivity(clientID)
}
