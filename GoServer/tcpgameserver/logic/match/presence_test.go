package match

import (
	"net"
	"testing"

	"GoServer/tcpgameserver/service"
)

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
