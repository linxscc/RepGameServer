package logic

import (
	"bufio"
	"net"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// 对局中断线 → 保留座位 → 新连接恢复，收到完整对局状态
func TestDisconnectThenResumeGame(t *testing.T) {
	cm := service.GetConnectionManager()
	room := newTestRoom(t, nil, NewBotName(), "carol",
		map[string][]models.Card{"carol": {card("k1", "Fire", 1)}})

	oldServer, oldClient := net.Pipe()
	defer oldClient.Close()
	cm.AddConnection(oldServer, "c-old")
	if err := cm.BindUser("c-old", "carol"); err != nil {
		t.Fatal(err)
	}
	_ = cm.SetPlayerStatus("c-old", types.StatusInGame)
	_ = cm.SetPlayerGameRoom("c-old", room.RoomID)

	PlayerDisconnected("c-old", "test")
	if !CanResumeGame("carol") {
		t.Fatal("in-game player must keep a seat after disconnecting")
	}

	newServer, newClient := net.Pipe()
	defer newClient.Close()
	cm.AddConnection(newServer, "c-new")

	got := make(chan *pb.ServerMessage, 1)
	go func() {
		payload, err := protocol.ReadFrame(bufio.NewReader(newClient))
		if err != nil {
			close(got)
			return
		}
		msg := &pb.ServerMessage{}
		_ = proto.Unmarshal(payload, msg)
		got <- msg
	}()

	if !ResumeGame("c-new", "carol") {
		t.Fatal("ResumeGame should succeed while the room is playing")
	}
	select {
	case msg := <-got:
		state := msg.GetGameState()
		if state == nil || state.GetReason() != pb.GameStateReason_GAME_STATE_REASON_RECONNECTED || len(state.GetSelfCards()) != 1 {
			t.Fatalf("expected RECONNECTED state with the hand, got %v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no state pushed to the new connection")
	}

	info, ok := cm.GetConnectionByUsername("carol")
	if !ok || info.ClientID != "c-new" || info.GetStatus() != types.StatusInGame {
		t.Fatalf("carol should now be bound to c-new and in game, got %+v", info)
	}
	if _, ok := cm.GetConnectionByClientID("c-old"); ok {
		t.Fatal("the old seat record should be replaced")
	}
	cancelReconnectGrace("carol")
}

// 不在对局中的玩家断线后记录直接清理
func TestDisconnectOutsideGameRemovesRecord(t *testing.T) {
	cm := service.GetConnectionManager()
	server, client := net.Pipe()
	defer client.Close()
	cm.AddConnection(server, "c-lobby")
	_ = cm.BindUser("c-lobby", "dave")

	PlayerDisconnected("c-lobby", "test")
	if _, ok := cm.GetConnectionByUsername("dave"); ok {
		t.Fatal("lobby player record should be removed on disconnect")
	}
}
