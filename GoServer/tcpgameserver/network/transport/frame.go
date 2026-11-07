// Package protocol 定义客户端与服务器之间的 TCP 消息帧格式。
//
// 每条消息 = 4 字节大端序长度 + 对应长度的 protobuf 编码内容。
package transport

import (
	"GoServer/tcpgameserver/data/config"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// MaxFrameSize 单条消息的最大字节数（network.yaml 的 max_frame_bytes）。
func MaxFrameSize() int {
	return config.Game().Network.Connection.MaxFrameBytes
}

const headerSize = 4

// EncodeFrame 给 payload 加上长度头。
func EncodeFrame(payload []byte) ([]byte, error) {
	if len(payload) > MaxFrameSize() {
		return nil, fmt.Errorf("protocol: frame too large (%d bytes)", len(payload))
	}
	buf := make([]byte, headerSize+len(payload))
	binary.BigEndian.PutUint32(buf, uint32(len(payload)))
	copy(buf[headerSize:], payload)
	return buf, nil
}

// WriteFrame 把 payload 加上长度头后发送（并发安全）。
// 已 Attach 的连接只是入队（见 conn_writer.go），不会阻塞调用方。
func WriteFrame(conn net.Conn, payload []byte) error {
	if conn == nil {
		return fmt.Errorf("protocol: nil connection")
	}
	frame, err := EncodeFrame(payload)
	if err != nil {
		return err
	}
	return sendFrame(conn, frame)
}

// ReadFrame 从 reader 读取一条完整消息。
func ReadFrame(r *bufio.Reader) ([]byte, error) {
	var header [headerSize]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(header[:])
	if int64(size) > int64(MaxFrameSize()) {
		return nil, fmt.Errorf("protocol: frame too large (%d bytes)", size)
	}

	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
