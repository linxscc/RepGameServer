package logic

import (
	"log"
	"net"

	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/service"
)

// 业务层向客户端发消息的统一出口。

func sendTo(conn net.Conn, msg *pb.ServerMessage) error {
	if conn == nil {
		return nil
	}
	err := protocol.Send(conn, msg)
	if err != nil {
		log.Printf("send %T to %s failed: %v", msg.GetBody(), conn.RemoteAddr(), err)
	}
	return err
}

func sendToClient(clientID string, msg *pb.ServerMessage) {
	if info, ok := service.GetConnectionManager().GetConnectionByClientID(clientID); ok && info != nil {
		_ = sendTo(info.GetConn(), msg)
	}
}

// notifyClientError 业务校验失败时回复发起请求的客户端。
func notifyClientError(clientID string, replyTo uint32, code pb.ErrorCode, text string) {
	sendToClient(clientID, protocol.NewError(replyTo, code, text))
}

// notifyRoomPeers 通知同房间除 exceptUser 以外的已登录玩家。
func notifyRoomPeers(roomID, exceptUser string, msg *pb.ServerMessage) {
	for _, info := range service.GetConnectionManager().GetAllConnections() {
		if info.GetGameRoom() == roomID && info.Username != "" && info.Username != exceptUser {
			_ = sendTo(info.GetConn(), msg)
		}
	}
}

func presenceMessage(username, roomID string, status pb.PresenceStatus, reason string) *pb.ServerMessage {
	return &pb.ServerMessage{Body: &pb.ServerMessage_PlayerPresence{PlayerPresence: &pb.PlayerPresence{
		Username: username,
		RoomId:   roomID,
		Status:   status,
		Reason:   reason,
	}}}
}

// requestMeta 从事件数据中取出发起请求的客户端和请求序号。
func requestMeta(get func(string) (interface{}, bool)) (clientID string, seq uint32) {
	if v, ok := get("client_id"); ok {
		clientID, _ = v.(string)
	}
	if v, ok := get("seq"); ok {
		seq, _ = v.(uint32)
	}
	return
}
