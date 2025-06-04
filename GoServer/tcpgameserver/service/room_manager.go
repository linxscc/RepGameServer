package service

import (
	"GoServer/tcpgameserver/types"
	"fmt"
	"log"
	"sync"
	"time"
)

// RoomManager 房间管理器
type RoomManager struct {
	rooms map[string]*types.RoomInfo // 房间列表，key为房间ID
	mutex sync.RWMutex               // 读写锁
}

var (
	roomManager *RoomManager
	roomOnce    sync.Once
)

// GetRoomManager 获取房间管理器单例
func GetRoomManager() *RoomManager {
	roomOnce.Do(func() {
		roomManager = &RoomManager{
			rooms: make(map[string]*types.RoomInfo),
		}
	})
	return roomManager
}

// CreateRoom 创建新房间
func (rm *RoomManager) CreateRoom(roomName string, maxPlayers int) (*types.RoomInfo, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 生成房间ID（使用时间戳）
	roomID := fmt.Sprintf("room_%d", time.Now().UnixNano())
	// 创建房间
	room := types.NewRoomInfo(roomID, roomName, maxPlayers)

	// 注意：卡牌池初始化将在逻辑层处理，避免循环依赖

	// 添加到房间列表
	rm.rooms[roomID] = room

	log.Printf("Created room: %s (%s) with max players: %d", roomID, roomName, maxPlayers)
	return room, nil
}

// GetRoom 获取指定房间
func (rm *RoomManager) GetRoom(roomID string) (*types.RoomInfo, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	return room, nil
}

// RemoveRoom 移除房间
func (rm *RoomManager) RemoveRoom(roomID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.rooms[roomID]; !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	delete(rm.rooms, roomID)
	log.Printf("Removed room: %s", roomID)
	return nil
}

// JoinRoom 玩家加入房间
func (rm *RoomManager) JoinRoom(roomID, username string) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return err
	}

	return room.AddPlayer(username)
}

// LeaveRoom 玩家离开房间
func (rm *RoomManager) LeaveRoom(roomID, username string) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return err
	}

	return room.RemovePlayer(username)
}

// FindRoomByPlayer 根据玩家名查找房间
func (rm *RoomManager) FindRoomByPlayer(username string) (*types.RoomInfo, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	for _, room := range rm.rooms {
		if _, err := room.GetPlayerInfo(username); err == nil {
			return room, nil
		}
	}

	return nil, fmt.Errorf("player %s not found in any room", username)
}

// GetAvailableRooms 获取可加入的房间列表
func (rm *RoomManager) GetAvailableRooms() []*types.RoomInfo {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var availableRooms []*types.RoomInfo
	for _, room := range rm.rooms {
		if !room.IsRoomFull() && room.Status == "waiting" {
			availableRooms = append(availableRooms, room)
		}
	}

	return availableRooms
}

// GetAllRooms 获取所有房间
func (rm *RoomManager) GetAllRooms() map[string]*types.RoomInfo {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	rooms := make(map[string]*types.RoomInfo)
	for id, room := range rm.rooms {
		rooms[id] = room
	}

	return rooms
}

// GetRoomStats 获取房间管理器统计信息
func (rm *RoomManager) GetRoomStats() map[string]interface{} {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	totalPlayers := 0
	waitingRooms := 0
	playingRooms := 0

	for _, room := range rm.rooms {
		stats := room.GetRoomStats()
		totalPlayers += stats["current_players"].(int)

		switch room.Status {
		case "waiting":
			waitingRooms++
		case "playing":
			playingRooms++
		}
	}

	return map[string]interface{}{
		"total_rooms":   len(rm.rooms),
		"waiting_rooms": waitingRooms,
		"playing_rooms": playingRooms,
		"total_players": totalPlayers,
	}
}
