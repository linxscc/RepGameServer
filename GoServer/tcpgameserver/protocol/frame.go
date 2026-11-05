// Package protocol 定义客户端与服务器之间的 TCP 消息帧格式。
//
// 每条消息 = 4 字节大端序长度 + 对应长度的 UTF-8 JSON 内容。
package protocol

import (
	"GoServer/tcpgameserver/config"
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

// MaxFrameSize 单条消息的最大字节数（network.yaml 的 max_frame_bytes）。
func MaxFrameSize() int {
	return config.Game().Network.Connection.MaxFrameBytes
}

const headerSize = 4

// 每个连接一把写锁，避免多个 goroutine 并发写同一连接时帧内容交错。
var writeLocks sync.Map // map[net.Conn]*sync.Mutex

func lockFor(conn net.Conn) *sync.Mutex {
	mu, _ := writeLocks.LoadOrStore(conn, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// WriteFrame 把 payload 加上长度头后完整写入连接（并发安全）。
func WriteFrame(conn net.Conn, payload []byte) error {
	if conn == nil {
		return fmt.Errorf("protocol: nil connection")
	}
	if len(payload) > MaxFrameSize() {
		return fmt.Errorf("protocol: frame too large (%d bytes)", len(payload))
	}

	buf := make([]byte, headerSize+len(payload))
	binary.BigEndian.PutUint32(buf, uint32(len(payload)))
	copy(buf[headerSize:], payload)

	mu := lockFor(conn)
	mu.Lock()
	defer mu.Unlock()

	_, err := conn.Write(buf)
	return err
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

// Forget 在连接关闭后释放其写锁。
func Forget(conn net.Conn) {
	writeLocks.Delete(conn)
}
