package protocol

import (
	"fmt"
	"net"

	"google.golang.org/protobuf/proto"

	"GoServer/tcpgameserver/pb"
)

// ProtocolVersion 随不兼容的协议变更递增，通过 Welcome 告知客户端。
const ProtocolVersion uint32 = 1

// Send 序列化一条服务器消息并按帧写入连接。
func Send(conn net.Conn, msg *pb.ServerMessage) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("protocol: marshal %T: %w", msg.GetBody(), err)
	}
	return WriteFrame(conn, payload)
}

// DecodeClientMessage 把一帧数据解析为客户端消息。
func DecodeClientMessage(payload []byte) (*pb.ClientMessage, error) {
	msg := &pb.ClientMessage{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		return nil, fmt.Errorf("protocol: unmarshal client message: %w", err)
	}
	return msg, nil
}

// NewError 构造一条错误回复。
func NewError(replyTo uint32, code pb.ErrorCode, text string) *pb.ServerMessage {
	return &pb.ServerMessage{
		ReplyTo: replyTo,
		Body:    &pb.ServerMessage_Error{Error: &pb.Error{Code: code, Message: text}},
	}
}
