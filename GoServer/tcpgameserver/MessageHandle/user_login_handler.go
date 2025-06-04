package tcpserver

import (
	"encoding/json"
	"log"
	"net"

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
