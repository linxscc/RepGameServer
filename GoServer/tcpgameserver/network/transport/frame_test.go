package transport

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRoundTripMultipleFramesInOneStream(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	msgs := []string{`{"a":1}`, `{"text":"中文 {不是} 括号\"}"}`, `{}`}
	go func() {
		for _, m := range msgs {
			if err := WriteFrame(server, []byte(m)); err != nil {
				t.Errorf("write: %v", err)
			}
		}
	}()

	r := bufio.NewReader(client)
	for _, want := range msgs {
		got, err := ReadFrame(r)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if string(got) != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestReadFrameAcrossSplitReads(t *testing.T) {
	payload := []byte(`{"msg":"拆包测试"}`)
	var stream bytes.Buffer
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	stream.Write(hdr[:])
	stream.Write(payload)

	// 每次只给 1 字节，模拟 TCP 拆包
	r := bufio.NewReaderSize(&oneByteReader{data: stream.Bytes()}, 16)
	got, err := ReadFrame(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("got %q", got)
	}
}

func TestReadFrameRejectsOversize(t *testing.T) {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(MaxFrameSize()+1))
	_, err := ReadFrame(bufio.NewReader(bytes.NewReader(hdr[:])))
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected too large error, got %v", err)
	}
}

func TestConcurrentWritesDoNotInterleave(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	const n = 50
	payload := []byte(strings.Repeat("x", 2000))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = WriteFrame(server, payload)
		}()
	}

	r := bufio.NewReader(client)
	for i := 0; i < n; i++ {
		got, err := ReadFrame(r)
		if err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("frame %d corrupted", i)
		}
	}
	wg.Wait()
}

type oneByteReader struct {
	data []byte
	pos  int
}

func (o *oneByteReader) Read(p []byte) (int, error) {
	if o.pos >= len(o.data) {
		return 0, bytes.ErrTooLarge
	}
	p[0] = o.data[o.pos]
	o.pos++
	return 1, nil
}

func TestAttachedWriterDeliversInOrderAndCloseAfterFlush(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	Attach(server)
	defer Detach(server)

	for i := 0; i < 5; i++ {
		if err := WriteFrame(server, []byte{byte(i)}); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}
	CloseAfterFlush(server)

	r := bufio.NewReader(client)
	for i := 0; i < 5; i++ {
		got, err := ReadFrame(r)
		if err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		if len(got) != 1 || got[0] != byte(i) {
			t.Fatalf("frame %d out of order: %v", i, got)
		}
	}
	if _, err := ReadFrame(r); err == nil {
		t.Fatal("expected connection closed after flush")
	}
}

func TestSlowReaderDoesNotBlockSender(t *testing.T) {
	server, client := net.Pipe() // 对端从不读取
	defer client.Close()
	Attach(server)
	defer Detach(server)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			_ = WriteFrame(server, []byte("x"))
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WriteFrame blocked on a reader that never reads")
	}
}
