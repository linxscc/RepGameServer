package types

import (
	"net"
	"sync"
	"time"
)

// PlayerStatus 玩家状态枚举
type PlayerStatus string

const (
	StatusConnected        PlayerStatus = "connected"         // 已连接
	StatusLoggedIn         PlayerStatus = "logged_in"         // 已登录
	StatusReady            PlayerStatus = "ready"             // 准备就绪
	StatusInGame           PlayerStatus = "in_game"           // 游戏中
	StatusWaitingReconnect PlayerStatus = "waiting_reconnect" // 等待重连
	StatusDisconnected     PlayerStatus = "disconnected"      // 已断开连接
)

// ClientInfo 客户端连接信息
type ClientInfo struct {
	// 连接信息
	Conn         net.Conn  `json:"-"`             // 网络连接（不序列化）
	ClientID     string    `json:"client_id"`     // 客户端唯一ID
	RemoteAddr   string    `json:"remote_addr"`   // 客户端地址
	ConnectedAt  time.Time `json:"connected_at"`  // 连接时间
	LastActivity time.Time `json:"last_activity"` // 最后活动时间
	// 用户信息
	Username   string `json:"username,omitempty"` // 用户名（登录后绑定）
	IsLoggedIn bool   `json:"is_logged_in"`       // 是否已登录

	// 状态信息
	Status     PlayerStatus `json:"status"`                 // 玩家状态
	GameRoomID string       `json:"game_room_id,omitempty"` // 所在游戏房间ID

	// 扩展信息
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 额外的元数据

	// 内部使用
	mutex sync.RWMutex `json:"-"` // 读写锁
}

// NewClientInfo 创建新的客户端信息
func NewClientInfo(conn net.Conn, clientID string) *ClientInfo {
	return &ClientInfo{
		Conn:         conn,
		ClientID:     clientID,
		RemoteAddr:   conn.RemoteAddr().String(),
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		Status:       StatusConnected,
		IsLoggedIn:   false,
		Metadata:     make(map[string]interface{}),
	}
}

// UpdateActivity 更新最后活动时间
func (c *ClientInfo) UpdateActivity() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.LastActivity = time.Now()
}

// BindUser 绑定用户信息
func (c *ClientInfo) BindUser(username string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Username = username
	c.IsLoggedIn = true
	c.Status = StatusLoggedIn
}

// UnbindUser 解绑用户信息
func (c *ClientInfo) UnbindUser() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Username = ""
	c.IsLoggedIn = false
	c.Status = StatusConnected
}

// SetStatus 设置玩家状态
func (c *ClientInfo) SetStatus(status PlayerStatus) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Status = status
}

// GetStatus 获取玩家状态
func (c *ClientInfo) GetStatus() PlayerStatus {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Status
}

// SetGameRoom 设置游戏房间
func (c *ClientInfo) SetGameRoom(roomID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.GameRoomID = roomID
}

// GetGameRoom 获取游戏房间ID
func (c *ClientInfo) GetGameRoom() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.GameRoomID
}

// SetMetadata 设置元数据
func (c *ClientInfo) SetMetadata(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Metadata[key] = value
}

// GetMetadata 获取元数据
func (c *ClientInfo) GetMetadata(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.Metadata[key]
	return value, exists
}

// IsActive 检查连接是否活跃（基于最后活动时间）
func (c *ClientInfo) IsActive(timeout time.Duration) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.LastActivity) < timeout
}

// GetConnectionDuration 获取连接持续时间
func (c *ClientInfo) GetConnectionDuration() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.ConnectedAt)
}
