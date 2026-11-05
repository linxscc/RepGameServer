package tcpserver

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	messagehandle "GoServer/tcpgameserver/messageHandle"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/setup"
	"GoServer/tcpgameserver/tlsutil"
)

// StartTCPServer 启动游戏服务器（默认启用 TLS）。
func StartTCPServer() {
	// 先加载并校验配置表，配置有误时在启动阶段就失败
	cfg := config.Game()
	log.Printf("config loaded: %d players/room, turn %ds, bot match after %ds (enabled=%v)",
		cfg.Rules.Match.PlayersPerRoom, cfg.Rules.Turn.TimeoutSeconds, cfg.Bot.MatchWaitSeconds, cfg.Bot.Enabled)

	setup.InitializeServer()

	ln, err := listen()
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

// listen 根据环境变量创建监听器。
//
//	TLS_CERT_FILE / TLS_KEY_FILE  证书路径，不存在时自动生成自签名证书
//	TLS_HOSTS                     自签名证书包含的 IP/域名，逗号分隔
//	TLS_DISABLED=true             仅限本地调试：使用明文 TCP
func listen() (net.Listener, error) {
	listenAddr := config.Game().Network.Connection.ListenAddr
	if strings.EqualFold(os.Getenv("TLS_DISABLED"), "true") {
		log.Println("WARNING: TLS disabled, traffic is NOT encrypted. Use only for local debugging.")
		ln, err := net.Listen("tcp", listenAddr)
		if err == nil {
			log.Printf("TCP server listening on %s (plaintext)", listenAddr)
		}
		return ln, err
	}

	cfg, fingerprint, err := tlsutil.ServerConfig(tlsutil.Options{
		CertFile: envOr("TLS_CERT_FILE", "certs/server.crt"),
		KeyFile:  envOr("TLS_KEY_FILE", "certs/server.key"),
		Hosts:    strings.Split(envOr("TLS_HOSTS", "localhost,127.0.0.1"), ","),
	})
	if err != nil {
		return nil, err
	}

	ln, err := tls.Listen("tcp", listenAddr, cfg)
	if err != nil {
		return nil, err
	}
	log.Printf("TLS server listening on %s", listenAddr)
	log.Printf("TLS certificate SHA-256 fingerprint (set SERVER_CERT_SHA256 in the client .env): %s", fingerprint)
	return ln, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// handleConnection 处理单个客户端连接：完成握手后循环读取帧并分发。
func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		protocol.Forget(conn)
	}()

	if tlsConn, ok := conn.(*tls.Conn); ok {
		_ = conn.SetDeadline(time.Now().Add(config.Game().Network.HandshakeTimeout()))
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		_ = conn.SetDeadline(time.Time{})
	}

	clientID := messagehandle.HandleNewConnection(conn)
	clientAddr := conn.RemoteAddr().String()
	defer messagehandle.HandleConnectionClose(clientID)

	// 超过 network.yaml 的 idle_timeout_seconds 没有任何消息（含心跳）即断开
	idleTimeout := config.Game().Network.IdleTimeout()
	reader := bufio.NewReader(conn)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
		payload, err := protocol.ReadFrame(reader)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				events.Publish(events.EventClientDisconnect, events.NewEventData(events.EventClientDisconnect, "tcp_server", map[string]interface{}{
					"client_id":      clientID,
					"client_address": clientAddr,
					"error":          err.Error(),
				}))
			}
			return
		}

		messagehandle.HandleTCPMessage(payload, conn, clientID)
	}
}
