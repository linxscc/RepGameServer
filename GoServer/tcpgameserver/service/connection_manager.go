package service

import (
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/protocol"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"GoServer/tcpgameserver/types"
)

// ConnectionManager 连接管理器
type ConnectionManager struct {
	// 连接映射 - 使用读写锁保护
	connections map[string]*types.ClientInfo // clientID -> ClientInfo
	userConns   map[string]string            // username -> clientID
	addrConns   map[string]string            // remoteAddr -> clientID

	// 读写锁
	mutex sync.RWMutex

	// 配置
	heartbeatTimeout time.Duration // 心跳超时时间
	cleanupInterval  time.Duration // 清理间隔

	// 停止信号
	stopChan chan struct{}
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections:      make(map[string]*types.ClientInfo),
		userConns:        make(map[string]string),
		addrConns:        make(map[string]string),
		heartbeatTimeout: config.Game().Network.StaleConnection(),
		cleanupInterval:  config.Game().Network.CleanupInterval(),
		stopChan:         make(chan struct{}),
	}
}

// Start 启动连接管理器（开始清理任务）
func (cm *ConnectionManager) Start() {
	go cm.cleanupInactiveConnections()
}

// Stop 停止连接管理器
func (cm *ConnectionManager) Stop() {
	close(cm.stopChan)
}

// AddConnection 添加新连接
func (cm *ConnectionManager) AddConnection(conn net.Conn, clientID string) *types.ClientInfo {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	clientInfo := types.NewClientInfo(conn, clientID)
	remoteAddr := conn.RemoteAddr().String()

	// 检查是否已存在相同地址的连接
	if existingClientID, exists := cm.addrConns[remoteAddr]; exists {
		cm.removeConnectionUnsafe(existingClientID)
	}

	// 添加新连接
	cm.connections[clientID] = clientInfo
	cm.addrConns[remoteAddr] = clientID

	return clientInfo
}

// RemoveConnection 移除连接
func (cm *ConnectionManager) RemoveConnection(clientID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.removeConnectionUnsafe(clientID)
}

// removeConnectionUnsafe 内部使用的移除连接方法（不加锁）
func (cm *ConnectionManager) removeConnectionUnsafe(clientID string) bool {
	clientInfo, exists := cm.connections[clientID]
	if !exists {
		return false
	}

	// 清理所有映射
	delete(cm.connections, clientID)
	delete(cm.addrConns, clientInfo.RemoteAddr)
	if clientInfo.Username != "" {
		delete(cm.userConns, clientInfo.Username)
	}

	// 关闭连接（先发完已排队的消息，例如"账号在别处登录"的通知）
	if conn := clientInfo.GetConn(); conn != nil {
		protocol.CloseAfterFlush(conn)
	}

	return true
}

// MarkWaitingReconnect 对局中断线：保留玩家记录（用户名映射不变）以便重连恢复，
// 但解除与已断开网络连接的关联。记录不存在时返回 false。
func (cm *ConnectionManager) MarkWaitingReconnect(clientID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	clientInfo, exists := cm.connections[clientID]
	if !exists {
		return false
	}
	if cm.addrConns[clientInfo.RemoteAddr] == clientID {
		delete(cm.addrConns, clientInfo.RemoteAddr)
	}
	clientInfo.DetachConn()
	return true
}

// BindUser 绑定用户到连接
func (cm *ConnectionManager) BindUser(clientID, username string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	clientInfo, exists := cm.connections[clientID]
	if !exists {
		return fmt.Errorf("connection not found for clientID: %s", clientID)
	}

	// 检查用户名是否已被其他连接使用
	if existingClientID, exists := cm.userConns[username]; exists && existingClientID != clientID {
		// 踢出旧连接
		cm.removeConnectionUnsafe(existingClientID)
	}

	// 绑定用户
	clientInfo.BindUser(username)
	cm.userConns[username] = clientID

	return nil
}

// UnbindUser 解绑用户
func (cm *ConnectionManager) UnbindUser(clientID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	clientInfo, exists := cm.connections[clientID]
	if !exists {
		return fmt.Errorf("connection not found for clientID: %s", clientID)
	}

	if clientInfo.Username != "" {
		delete(cm.userConns, clientInfo.Username)
	}

	clientInfo.UnbindUser()
	return nil
}

// GetConnectionByClientID 根据客户端ID获取连接信息
func (cm *ConnectionManager) GetConnectionByClientID(clientID string) (*types.ClientInfo, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clientInfo, exists := cm.connections[clientID]
	return clientInfo, exists
}

// GetConnectionByUsername 根据用户名获取连接信息
func (cm *ConnectionManager) GetConnectionByUsername(username string) (*types.ClientInfo, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clientID, exists := cm.userConns[username]
	if !exists {
		return nil, false
	}

	clientInfo, exists := cm.connections[clientID]
	return clientInfo, exists
}

// GetConnectionByAddr 根据地址获取连接信息
func (cm *ConnectionManager) GetConnectionByAddr(addr string) (*types.ClientInfo, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clientID, exists := cm.addrConns[addr]
	if !exists {
		return nil, false
	}

	clientInfo, exists := cm.connections[clientID]
	return clientInfo, exists
}

// UpdateActivity 更新连接活动时间
func (cm *ConnectionManager) UpdateActivity(clientID string) {
	cm.mutex.RLock()
	clientInfo, exists := cm.connections[clientID]
	cm.mutex.RUnlock()

	if exists {
		clientInfo.UpdateActivity()
	}
}

// SetPlayerStatus 设置玩家状态
func (cm *ConnectionManager) SetPlayerStatus(clientID string, status types.PlayerStatus) error {
	cm.mutex.RLock()
	clientInfo, exists := cm.connections[clientID]
	cm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found for clientID: %s", clientID)
	}

	clientInfo.SetStatus(status)
	return nil
}

// SetPlayerGameRoom 设置玩家游戏房间
func (cm *ConnectionManager) SetPlayerGameRoom(clientID, roomID string) error {
	cm.mutex.RLock()
	clientInfo, exists := cm.connections[clientID]
	cm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found for clientID: %s", clientID)
	}

	clientInfo.SetGameRoom(roomID)
	return nil
}

// GetAllConnections 获取所有连接信息
func (cm *ConnectionManager) GetAllConnections() map[string]*types.ClientInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[string]*types.ClientInfo)
	for clientID, clientInfo := range cm.connections {
		result[clientID] = clientInfo
	}
	return result
}

// GetConnectionsByStatus 根据状态获取连接列表
func (cm *ConnectionManager) GetConnectionsByStatus(status types.PlayerStatus) []*types.ClientInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var result []*types.ClientInfo
	for _, clientInfo := range cm.connections {
		if clientInfo.GetStatus() == status {
			result = append(result, clientInfo)
		}
	}
	return result
}

// GetConnectionStats 获取连接统计信息
func (cm *ConnectionManager) GetConnectionStats() map[string]int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	stats := map[string]int{
		"total":        len(cm.connections),
		"logged_in":    0,
		"connected":    0,
		"ready":        0,
		"in_game":      0,
		"disconnected": 0,
	}

	for _, clientInfo := range cm.connections {
		switch clientInfo.GetStatus() {
		case types.StatusLoggedIn:
			stats["logged_in"]++
		case types.StatusConnected:
			stats["connected"]++
		case types.StatusReady:
			stats["ready"]++
		case types.StatusInGame:
			stats["in_game"]++
		case types.StatusDisconnected:
			stats["disconnected"]++
		}
	}

	return stats
}

// SendToUser 向特定用户发送消息
func (cm *ConnectionManager) SendToUser(username string, data []byte) error {
	clientInfo, exists := cm.GetConnectionByUsername(username)
	if !exists {
		return fmt.Errorf("user %s not found or not connected", username)
	}

	conn := clientInfo.GetConn()
	if conn == nil {
		return fmt.Errorf("connection is nil for user %s", username)
	}

	err := protocol.WriteFrame(conn, data)
	if err != nil {
		// 连接出错，移除该连接
		cm.RemoveConnection(clientInfo.ClientID)
		return err
	}

	// 更新活动时间
	clientInfo.UpdateActivity()
	return nil
}

// SendToClient 向特定客户端发送消息
func (cm *ConnectionManager) SendToClient(clientID string, data []byte) error {
	clientInfo, exists := cm.GetConnectionByClientID(clientID)
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	conn := clientInfo.GetConn()
	if conn == nil {
		return fmt.Errorf("connection is nil for client %s", clientID)
	}

	err := protocol.WriteFrame(conn, data)
	if err != nil {
		// 连接出错，移除该连接
		cm.RemoveConnection(clientID)
		return err
	}

	// 更新活动时间
	clientInfo.UpdateActivity()
	return nil
}

// Broadcast 广播消息给所有连接
func (cm *ConnectionManager) Broadcast(data []byte) {
	cm.mutex.RLock()
	connections := make([]*types.ClientInfo, 0, len(cm.connections))
	for _, clientInfo := range cm.connections {
		connections = append(connections, clientInfo)
	}
	cm.mutex.RUnlock()

	for _, clientInfo := range connections {
		if conn := clientInfo.GetConn(); conn != nil {
			err := protocol.WriteFrame(conn, data)
			if err != nil {
				// 连接出错，移除该连接
				cm.RemoveConnection(clientInfo.ClientID)
			} else {
				clientInfo.UpdateActivity()
			}
		}
	}
}

// cleanupInactiveConnections 清理不活跃的连接
func (cm *ConnectionManager) cleanupInactiveConnections() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.performCleanup()
		case <-cm.stopChan:
			return
		}
	}
}

// performCleanup 执行清理操作
func (cm *ConnectionManager) performCleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var toRemove []string
	for clientID, clientInfo := range cm.connections {
		// 等待重连的玩家由重连宽限计时器负责清理
		if clientInfo.GetStatus() == types.StatusWaitingReconnect {
			continue
		}
		if !clientInfo.IsActive(cm.heartbeatTimeout) {
			toRemove = append(toRemove, clientID)
		}
	}

	for _, clientID := range toRemove {
		cm.removeConnectionUnsafe(clientID)
	}

	if len(toRemove) > 0 {
		log.Printf("connection cleanup removed %d stale records", len(toRemove))
	}
}

// 全局连接管理器实例
var globalConnectionManager *ConnectionManager

// InitConnectionManager 初始化全局连接管理器
func InitConnectionManager() {
	globalConnectionManager = NewConnectionManager()
	globalConnectionManager.Start()
	log.Println("Connection manager initialized and started")
}

// GetConnectionManager 获取全局连接管理器
func GetConnectionManager() *ConnectionManager {
	if globalConnectionManager == nil {
		InitConnectionManager()
	}
	return globalConnectionManager
}

// StopConnectionManager 停止全局连接管理器
func StopConnectionManager() {
	if globalConnectionManager != nil {
		globalConnectionManager.Stop()
	}
}
