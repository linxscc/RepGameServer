package tcpserver

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"

	"GoServer/tcpgameserver/logic"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/tools"
)

type ClientInfo struct {
	ID   string
	Addr string
	// 可扩展更多字段，如游戏状态、上次活跃时间等
}

var (
	clientMap   = make(map[string]*ClientInfo) // key: 客户端IP:端口
	clientIDMap = make(map[string]string)      // key: 客户端ID, value: IP:端口
	clientLock  sync.Mutex
)

// 处理TCP消息
func HandleTCPMessage(msg string, conn net.Conn) models.TcpResponse {
	remoteAddr := conn.RemoteAddr().String()
	clientLock.Lock()
	defer clientLock.Unlock()

	msg = strings.TrimSpace(msg)

	// 尝试将msg解析为TcpRequest结构体
	var req models.TcpRequest
	err := json.Unmarshal([]byte(msg), &req)
	if err != nil {
		log.Printf("Failed to parse TcpRequest: %v, raw: %s", err, msg)
		return tools.GetResponseCodeByID("4001", nil)
	}

	switch req.Message {
	case "startconnect":
		// 客户端连接/重连，分配ID
		id := clientMap[remoteAddr]
		if id == nil {
			clientID := generateClientID(remoteAddr)
			clientMap[remoteAddr] = &ClientInfo{ID: clientID, Addr: remoteAddr}
			clientIDMap[clientID] = remoteAddr
			log.Printf("New client connected: %s, assigned ID: %s", remoteAddr, clientID)
			return tools.GetResponseCodeByID("2001", nil)
		} else {
			log.Printf("Client reconnected: %s, ID: %s", remoteAddr, id.ID)
			return tools.GetResponseCodeByID("2002", nil)
		}
	case "healthUpdateList":
		healthUpdates, _ := tools.ConvertJsontoList_Health(req.Data)
		if healthUpdates == nil {
			log.Printf("Failed to convert health updates from request data: %v", req.Data)
			return tools.GetResponseCodeByID("4001", nil)
		}

		logic.HandleHealthUpdateList(healthUpdates)
		return tools.GetResponseCodeByID("2003", nil)

	// 可扩展更多消息类型
	default:
		log.Printf("Unknown message type: %s", req.Message)
		return tools.GetResponseCodeByID("4000", nil)
	}
}

// 生成客户端ID（可自定义更复杂的生成方式）
func generateClientID(addr string) string {
	return strings.ReplaceAll(addr, ":", "_")
}
