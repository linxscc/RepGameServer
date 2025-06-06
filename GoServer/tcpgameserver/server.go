package tcpserver

import (
	"GoServer/tcpgameserver/events"
	messagehandle "GoServer/tcpgameserver/messageHandle"
	"GoServer/tcpgameserver/setup"
	"log"
	"net"
)

// 启动TCP服务器
func StartTCPServer() {
	// 初始化服务器组件（响应码、卡牌池、游戏开始回调等）
	setup.InitializeServer()

	// 启动TCP监听
	ln, err := net.Listen("tcp", ":9060")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP server listening on :9060")

	// 处理连接请求
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// 处理客户端连接
func handleConnection(conn net.Conn) {
	// 使用连接管理器处理新连接
	clientID := messagehandle.HandleNewConnection(conn)
	clientAddr := conn.RemoteAddr().String()
	defer func() {
		messagehandle.HandleConnectionClose(clientID)

		conn.Close()
	}()

	buf := make([]byte, 4096) // 8KB buffer to handle 6KB data with safety margin
	for {
		n, err := conn.Read(buf)
		if err != nil {
			// 发布连接超时或错误事件
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				timeoutData := events.NewEventData(events.EventClientDisconnect, "tcp_server", map[string]interface{}{
					"client_id":      clientID,
					"client_address": clientAddr,
					"error":          err.Error(),
				})
				events.Publish(events.EventClientDisconnect, timeoutData)
			}
			break // 连接断开或出错，退出循环
		}

		// 更新客户端活动时间
		messagehandle.UpdateClientActivity(clientID)

		msg := string(buf[:n])
		messagehandle.HandleTCPMessage(msg, conn, clientID)
	}
}

// 启动UDP服务器
func StartUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", ":9060")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("UDP server listening on :9060")
	for {
		buf := make([]byte, 1024)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err == nil {
			log.Printf("UDP received: %s", string(buf[:n]))
			conn.WriteToUDP([]byte("pong"), remoteAddr)
		}
	}
}
