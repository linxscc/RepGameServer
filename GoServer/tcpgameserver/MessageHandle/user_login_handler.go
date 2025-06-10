package tcpserver

import (
	"encoding/json"
	"log"
	"net"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
)

// HandleUserLogin 处理用户登录
func HandleUserLogin(req models.TcpRequest, conn net.Conn, clientID string, connManager *service.ConnectionManager) {
	var loginData models.UserAccount
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		log.Printf("Failed to marshal UserLogin data: %v, raw: %v", err, req.Data)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2002))
		return
	}
	if err := json.Unmarshal(dataBytes, &loginData); err != nil {
		log.Printf("Failed to parse UserLogin data: %v, raw: %v", err, req.Data)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2002))
		return
	}

	// 验证参数
	if loginData.Username == "" || loginData.Password == "" {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2003))
		return
	}

	// 使用数据库服务验证登录
	isValid, err := service.ValidateUserLogin(loginData.Username, loginData.Password)
	if err != nil {
		log.Printf("User login error: %v", err)
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2004))
		return
	}
	if !isValid {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(2005))
		return
	}
	// 检查用户是否已经登录
	if existingClient, isLoggedIn := connManager.GetConnectionByUsername(loginData.Username); isLoggedIn {
		// 检查用户状态是否在等待重连
		if existingClient.GetStatus() == types.StatusWaitingReconnect {
			log.Printf("User %s is waiting for reconnection, triggering reconnection event", loginData.Username)

			// 发送重连事件
			reconnectData := events.CreateUserConnectionEventData(
				events.EventClientReconnect, clientID, loginData.Username, conn.RemoteAddr().String())
			reconnectData.AddData("old_client_id", existingClient.ClientID)
			events.Publish(events.EventClientReconnect, reconnectData)

			// 直接返回，不继续执行登录逻辑
			return
		}

		// 用户已登录且不在等待重连状态，踢出原客户端连接
		log.Printf("User %s is already logged in from clientID: %s, kicking existing client", loginData.Username, existingClient.ClientID)

		// 发布踢出事件
		kickData := events.CreateUserConnectionEventData(
			events.EventClientKicked, existingClient.ClientID, loginData.Username, existingClient.RemoteAddr)
		kickData.AddData("kick_reason", "duplicate_login")
		kickData.AddData("kicked_by", "system")
		kickData.AddData("new_client_id", clientID)
		kickData.AddData("new_login_data", loginData)
		events.Publish(events.EventClientKicked, kickData)

		// 继续执行新客户端的登录逻辑
	}

	// 登录成功，绑定用户到连接
	err = connManager.BindUser(clientID, loginData.Username)
	if err != nil {
		log.Printf("Failed to bind user to connection: %v", err)
	}

	// 设置玩家状态为已登录
	connManager.SetPlayerStatus(clientID, types.StatusLoggedIn)

	SendTCPResponse(conn, tools.GlobalResponseHelper.CreateSuccessTcpResponse(2001, map[string]interface{}{
		"username": loginData.Username,
	}))
}
