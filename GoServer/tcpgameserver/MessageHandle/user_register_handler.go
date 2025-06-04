package tcpserver

import (
	"encoding/json"
	"net"
	"strings"

	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
)

// HandleUserRegister 处理用户注册
func HandleUserRegister(req models.TcpRequest, conn net.Conn, clientID string, connManager *service.ConnectionManager) {
	var registerData models.UserAccount
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(3002))
		return
	}
	if err := json.Unmarshal(dataBytes, &registerData); err != nil {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(3002))
		return
	}

	// 验证参数
	if registerData.Username == "" || registerData.Password == "" {
		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(3003))
		return
	}

	// 使用数据库服务创建用户
	err = service.CreateUserAccount(registerData.Username, registerData.Password)
	if err != nil {

		// 根据错误类型返回不同的响应
		if strings.Contains(err.Error(), "user already exists") {
			SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(3004))
			return
		}

		SendTCPResponse(conn, tools.GlobalResponseHelper.CreateErrorTcpResponse(3005))
		return
	}

	// 创建成功
	SendTCPResponse(conn, tools.GlobalResponseHelper.CreateSuccessTcpResponse(3001, map[string]interface{}{
		"username": registerData.Username,
		"message":  "用户创建成功",
	}))
}
