package protocol

import (
	"errors"
	"net"
	"sync"
	"time"

	"GoServer/tcpgameserver/config"
)

// 每个连接一个发送队列 + 一个专用写协程：
//   - 业务代码发送消息只是入队，永远不会被慢客户端阻塞；
//   - 每次写都有超时，对端长期不读数据时连接会被关闭；
//   - 同一连接的消息按入队顺序发出，不会交错。
//
// 连接建立后调用 Attach，连接结束时调用 Detach。未 Attach 的连接（如单元测试中的 net.Pipe）
// 退化为直接写。

// ErrConnClosed 连接已关闭或已被 Detach，消息未发送。
var ErrConnClosed = errors.New("protocol: connection closed")

// ErrSendQueueFull 发送队列已满：对端处理不过来，连接已被关闭。
var ErrSendQueueFull = errors.New("protocol: send queue full")

type connWriter struct {
	conn    net.Conn
	queue   chan []byte // nil 元素表示"写完前面的消息后关闭连接"
	done    chan struct{}
	stop    sync.Once
	timeout time.Duration
}

var writers sync.Map // net.Conn -> *connWriter

// Attach 为连接启动发送队列。重复调用无效果。
func Attach(conn net.Conn) {
	cfg := config.Game().Network
	w := &connWriter{
		conn:    conn,
		queue:   make(chan []byte, cfg.Connection.SendQueueSize),
		done:    make(chan struct{}),
		timeout: cfg.WriteTimeout(),
	}
	if _, loaded := writers.LoadOrStore(conn, w); loaded {
		return
	}
	go w.run()
}

// Detach 停止连接的发送队列（未发出的消息被丢弃）。连接关闭后调用。
func Detach(conn net.Conn) {
	if v, ok := writers.LoadAndDelete(conn); ok {
		v.(*connWriter).shutdown()
	}
}

// CloseAfterFlush 发完队列中已有的消息后关闭连接，用于"先通知再断开"（如被顶号）。
// 队列已满或连接未 Attach 时立即关闭。
func CloseAfterFlush(conn net.Conn) {
	if conn == nil {
		return
	}
	w := writerFor(conn)
	if w == nil || w.enqueue(nil) != nil {
		_ = conn.Close()
	}
}

func writerFor(conn net.Conn) *connWriter {
	if v, ok := writers.Load(conn); ok {
		return v.(*connWriter)
	}
	return nil
}

// sendFrame 发送一条已编码的帧：已 Attach 的连接入队，否则直接写。
func sendFrame(conn net.Conn, frame []byte) error {
	if w := writerFor(conn); w != nil {
		return w.enqueue(frame)
	}
	_, err := conn.Write(frame) // net.Conn 允许并发 Write，单次调用写完整帧
	return err
}

func (w *connWriter) enqueue(frame []byte) error {
	select {
	case <-w.done:
		return ErrConnClosed
	default:
	}
	select {
	case w.queue <- frame:
		return nil
	case <-w.done:
		return ErrConnClosed
	default:
		_ = w.conn.Close() // 读循环随之退出，走正常的断线流程
		w.shutdown()
		return ErrSendQueueFull
	}
}

func (w *connWriter) run() {
	for {
		select {
		case <-w.done:
			return
		case frame := <-w.queue:
			if frame == nil {
				_ = w.conn.Close()
				w.shutdown()
				return
			}
			_ = w.conn.SetWriteDeadline(time.Now().Add(w.timeout))
			if _, err := w.conn.Write(frame); err != nil {
				_ = w.conn.Close()
				w.shutdown()
				return
			}
		}
	}
}

func (w *connWriter) shutdown() {
	w.stop.Do(func() { close(w.done) })
}
