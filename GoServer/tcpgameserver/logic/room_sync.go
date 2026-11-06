package logic

import (
	"fmt"
	"sync"

	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// 房间操作串行化
// ------------------------------------------------------------
// 同一房间的所有会改变或读取对局进度的操作——真人操作、阶段计时器、机器人决策、
// 开局、结束、重连推送——都通过 withRoom 执行，同一时刻只有一个在运行。
// 这样"校验 → 结算 → 推进阶段 → 广播快照"是一个不可分割的整体，
// 不会出现两次操作交错导致的半结算，广播的快照也一定是某次操作完成后的状态。
//
// 约定：
//   - withRoom 不可重入：fn 内不得再调用 withRoom（同一房间会死锁）。
//   - fn 内发送消息是安全的：发送只是入队（见 protocol/conn_writer.go），不会阻塞。

var roomLocks sync.Map // roomID -> *sync.Mutex

// withRoom 持有房间操作锁执行 fn。房间不存在（或已结束被删除）时返回错误且不执行 fn。
func withRoom(roomID string, fn func(room *types.RoomInfo) error) error {
	mu, _ := roomLocks.LoadOrStore(roomID, &sync.Mutex{})
	m := mu.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()

	// 必须在持锁后再取房间：等待期间房间可能已结束并被删除
	room, err := service.GetRoomManager().GetRoom(roomID)
	if err != nil {
		return fmt.Errorf("room %s: %w", roomID, err)
	}
	return fn(room)
}

// forgetRoomLock 房间删除后释放其锁（仍在等待的调用拿到锁后会发现房间已不存在）
func forgetRoomLock(roomID string) {
	roomLocks.Delete(roomID)
}
