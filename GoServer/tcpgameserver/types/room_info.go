package types

import (
	"fmt"
	"sort"
	"sync"
)

// 房间状态
const (
	RoomStatusWaiting  = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

// PlayerInfo 房间内玩家信息（对局数据在 rules.Match 中，见 logic/chess_match.go）
type PlayerInfo struct {
	Username string `json:"username"` // 玩家用户名
}

// RoomInfo 游戏房间信息
type RoomInfo struct {
	// 房间基本信息
	RoomID     string `json:"room_id"`     // 房间ID
	RoomName   string `json:"room_name"`   // 房间名称
	MaxPlayers int    `json:"max_players"` // 最大玩家数量
	Status     string `json:"status"`      // 房间状态：waiting, playing, finished

	// 玩家信息
	Players map[string]*PlayerInfo `json:"players"` // 玩家列表，key为username

	// 内部使用
	mutex sync.RWMutex `json:"-"` // 读写锁
}

// NewRoomInfo 创建新的房间信息
func NewRoomInfo(roomID, roomName string, maxPlayers int) *RoomInfo {
	return &RoomInfo{
		RoomID:     roomID,
		RoomName:   roomName,
		MaxPlayers: maxPlayers,
		Status:     RoomStatusWaiting,
		Players:    make(map[string]*PlayerInfo),
	}
}

// AddPlayer 添加玩家到房间
func (r *RoomInfo) AddPlayer(username string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.Players) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}
	if _, exists := r.Players[username]; exists {
		return fmt.Errorf("player %s already in room", username)
	}

	r.Players[username] = &PlayerInfo{Username: username}
	return nil
}

// RemovePlayer 从房间移除玩家
func (r *RoomInfo) RemovePlayer(username string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.Players[username]; !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	delete(r.Players, username)
	return nil
}

// GetRoomStats 获取房间统计信息
func (r *RoomInfo) GetRoomStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return map[string]interface{}{
		"room_id":         r.RoomID,
		"room_name":       r.RoomName,
		"status":          r.Status,
		"current_players": len(r.Players),
		"max_players":     r.MaxPlayers,
	}
}

// UpdateRoomStatus 更新房间状态
func (r *RoomInfo) UpdateRoomStatus(status string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Status = status
}

// IsRoomFull 检查房间是否已满
func (r *RoomInfo) IsRoomFull() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.Players) >= r.MaxPlayers
}

// ------------------------------------------------------------
// 以下方法供业务层读写房间而不直接访问内部字段，全部持有房间锁。
// ------------------------------------------------------------

// GetStatus 房间状态（见 RoomStatus* 常量）
func (r *RoomInfo) GetStatus() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.Status
}

// PlayerNames 房间内所有玩家，按名称排序（结果稳定）
func (r *RoomInfo) PlayerNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	names := make([]string, 0, len(r.Players))
	for name := range r.Players {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasPlayer 玩家是否在房间中
func (r *RoomInfo) HasPlayer(username string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	_, ok := r.Players[username]
	return ok
}

// ResetForGameEnd 对局结束：标记房间已结束
func (r *RoomInfo) ResetForGameEnd() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Status = RoomStatusFinished
}
