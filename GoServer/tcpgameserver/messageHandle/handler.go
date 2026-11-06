package tcpserver

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/logic"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/service"
)

// Context 单条请求的处理上下文。
type Context struct {
	Conn        net.Conn
	ClientID    string
	Seq         uint32
	ConnManager *service.ConnectionManager
}

// Reply 回复当前请求（自动填写 reply_to）。
func (c *Context) Reply(msg *pb.ServerMessage) {
	msg.ReplyTo = c.Seq
	if err := protocol.Send(c.Conn, msg); err != nil {
		log.Printf("send reply to %s failed: %v", c.ClientID, err)
	}
}

// Fail 以错误回复当前请求。
func (c *Context) Fail(code pb.ErrorCode, text string) {
	c.Reply(protocol.NewError(0, code, text))
}

// RemoteIP 客户端 IP（不含端口），用于限流
func (c *Context) RemoteIP() string {
	addr := c.Conn.RemoteAddr().String()
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

// Username 当前连接已登录的用户名；未登录时为空。
func (c *Context) Username() string {
	if info, ok := c.ConnManager.GetConnectionByClientID(c.ClientID); ok && info.IsLoggedIn {
		return info.Username
	}
	return ""
}

type handlerFunc func(ctx *Context, msg *pb.ClientMessage)

type route struct {
	requireAuth bool
	handle      handlerFunc
}

// routes 按 ClientMessage.body 的 oneof 分支名注册处理器。
// 新增请求类型时只需在 .proto 中加分支并在此注册，分发逻辑无需改动。
var routes = map[protoreflect.Name]route{
	"login":          {requireAuth: false, handle: HandleLogin},
	"register":       {requireAuth: false, handle: HandleRegister},
	"resume_session": {requireAuth: false, handle: HandleResumeSession},
	"ping":           {requireAuth: false, handle: HandlePing},
	"ready":          {requireAuth: true, handle: HandleReady},
	"game_action":    {requireAuth: true, handle: HandleGameAction},
	"cancel_match":   {requireAuth: true, handle: HandleCancelMatch},
}

var bodyOneof = (&pb.ClientMessage{}).ProtoReflect().Descriptor().Oneofs().ByName("body")

// HandleTCPMessage 解析一帧客户端数据并分发到对应处理器。
func HandleTCPMessage(payload []byte, conn net.Conn, clientID string) {
	connManager := service.GetConnectionManager()
	connManager.UpdateActivity(clientID)

	msg, err := protocol.DecodeClientMessage(payload)
	if err != nil {
		_ = protocol.Send(conn, protocol.NewError(0, pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, config.Msg(config.MsgMalformedMessage)))
		return
	}

	ctx := &Context{Conn: conn, ClientID: clientID, Seq: msg.GetSeq(), ConnManager: connManager}

	field := msg.ProtoReflect().WhichOneof(bodyOneof)
	if field == nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, config.Msg(config.MsgEmptyMessageBody))
		return
	}

	r, ok := routes[field.Name()]
	if !ok {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, config.Msg(config.MsgUnsupportedRequest, field.Name()))
		return
	}
	if r.requireAuth && ctx.Username() == "" {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_NOT_AUTHENTICATED, config.Msg(config.MsgLoginRequired))
		return
	}

	r.handle(ctx, msg)
}

// SendWelcome 连接建立后发送的第一条消息。
func SendWelcome(conn net.Conn) {
	_ = protocol.Send(conn, &pb.ServerMessage{Body: &pb.ServerMessage_Welcome{Welcome: &pb.Welcome{
		ServerTimeMs:    time.Now().UnixMilli(),
		ProtocolVersion: protocol.ProtocolVersion,
	}}})
}

// HandleNewConnection 处理新的TCP连接
func HandleNewConnection(conn net.Conn) string {
	remoteAddr := conn.RemoteAddr().String()

	// 生成客户端ID（包含时间戳和地址，确保唯一性）
	clientID := fmt.Sprintf("client_%d_%s", time.Now().UnixNano(),
		strings.ReplaceAll(remoteAddr, ":", "_"))

	connManager := service.GetConnectionManager()

	// 检查是否是重连的客户端（基于地址）
	if existingInfo, exists := connManager.GetConnectionByAddr(remoteAddr); exists {
		connManager.RemoveConnection(existingInfo.ClientID)
	}
	connManager.AddConnection(conn, clientID)

	SendWelcome(conn)
	return clientID
}

// HandleConnectionClose 处理连接关闭（同步执行：对局中的玩家转为等待重连，其余清理记录）
func HandleConnectionClose(clientID string) {
	logic.PlayerDisconnected(clientID, "connection_close")
}

// UpdateClientActivity 更新客户端活动时间
func UpdateClientActivity(clientID string) {
	service.GetConnectionManager().UpdateActivity(clientID)
}
