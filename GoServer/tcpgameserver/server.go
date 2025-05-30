package tcpserver

import (
	messagehandle "GoServer/tcpgameserver/MessageHandle"
	"GoServer/tcpgameserver/tools"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
)

// 启动TCP服务器
func StartTCPServer() {
	// 启动时加载响应码（使用相对路径，便于部署）
	wd, _ := os.Getwd()
	jsonPath := filepath.Join(wd, "goserver", "tcpgameserver", "config", "response_codes.json")
	log.Println("Loading response codes from:", jsonPath)
	tools.LoadResponseCodes(jsonPath)
	ln, err := net.Listen("tcp", ":9060")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP server listening on :9060")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			for {
				n, err := c.Read(buf)
				if err != nil {
					break // 连接断开或出错，退出循环
				}
				msg := string(buf[:n])
				log.Printf("TCP received: %s", msg)
				resp := messagehandle.HandleTCPMessage(msg, c)
				jsonBytes, _ := json.Marshal(resp)
				jsonBytes = append(jsonBytes, '\n')
				c.Write(jsonBytes)
			}
			log.Printf("Connection closed: %s", c.RemoteAddr().String())
		}(conn)
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
