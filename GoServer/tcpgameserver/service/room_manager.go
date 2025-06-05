package service

import (
	"GoServer/tcpgameserver/models"
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

// GetPlayerGameInfo 获取玩家游戏信息
func (rm *RoomManager) GetPlayerGameInfo(username string) (*models.PlayerGameInfo, error) {
	room, err := rm.FindRoomByPlayer(username)
	if err != nil {
		return nil, err
	}

	// 检查房间是否在游戏状态
	if room.Status != "playing" {
		return nil, nil // 不在游戏中
	}

	// 获取房间内所有玩家
	allPlayers := make([]string, 0, len(room.Players))
	for username := range room.Players {
		allPlayers = append(allPlayers, username)
	}

	// 创建游戏信息
	return rm.createPlayerGameInfo(room, username, allPlayers), nil
}

// createPlayerGameInfo 创建玩家游戏信息
func (rm *RoomManager) createPlayerGameInfo(room *types.RoomInfo, username string, allPlayers []string) *models.PlayerGameInfo {
	roomPlayer := room.Players[username]

	// 获取对方玩家的卡牌列表
	otherCards := make([]models.Card, 0)
	for _, playerName := range allPlayers {
		if playerName != username {
			if otherPlayer, exists := room.Players[playerName]; exists {
				otherCards = append(otherCards, otherPlayer.HandCards...)
			}
		}
	}

	return &models.PlayerGameInfo{
		RoomId:         room.RoomID,
		Username:       username,
		Round:          roomPlayer.Round,
		Health:         float64(roomPlayer.CurrentHealth),
		DamageDealt:    roomPlayer.DamageDealt,
		DamageReceived: roomPlayer.DamageReceived,
		TriggeredBonds: make([]models.BondModel, 0),
		SelfCards:      roomPlayer.HandCards,
		OtherCards:     otherCards,
	}
}

// DrawCardFromLevel1Pool 从指定房间的一级卡牌池中抽取一张卡牌
func (rm *RoomManager) DrawCardFromLevel1Pool(roomID string) (*models.Card, error) {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	return room.DrawRandomCardFromLevel1Pool()
}

// DrawCardsFromLevel1Pool 从指定房间的一级卡牌池中抽取多张卡牌
func (rm *RoomManager) DrawCardsFromLevel1Pool(roomID string, count int) ([]models.Card, error) {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	return room.DrawRandomCardsFromLevel1Pool(count)
}

// DrawCardForPlayer 为指定玩家从一级卡牌池抽取一张卡牌并添加到手牌
func (rm *RoomManager) DrawCardForPlayer(roomID, username string) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 从一级卡牌池抽取卡牌
	card, err := room.DrawRandomCardFromLevel1Pool()
	if err != nil {
		return fmt.Errorf("failed to draw card from level 1 pool: %v", err)
	}

	// 将卡牌添加到玩家手牌
	err = room.AddCardToPlayer(username, *card)
	if err != nil {
		// 如果添加失败，需要将卡牌放回卡牌池
		room.Level1CardPool = append(room.Level1CardPool, *card)
		return fmt.Errorf("failed to add card to player %s: %v", username, err)
	}

	log.Printf("Drew card %s (UID: %d) for player %s in room %s", card.Name, card.UID, username, roomID)
	return nil
}

// DrawCardsForPlayer 为指定玩家从一级卡牌池抽取多张卡牌并添加到手牌
func (rm *RoomManager) DrawCardsForPlayer(roomID, username string, count int) error {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 检查玩家当前手牌数量
	playerInfo, err := room.GetPlayerInfo(username)
	if err != nil {
		return err
	}

	// 检查是否有足够空间添加卡牌
	if len(playerInfo.HandCards)+count > room.MaxHandCards {
		return fmt.Errorf("player %s cannot hold %d more cards (current: %d, max: %d)",
			username, count, len(playerInfo.HandCards), room.MaxHandCards)
	}

	// 从一级卡牌池抽取卡牌
	cards, err := room.DrawRandomCardsFromLevel1Pool(count)
	if err != nil {
		return fmt.Errorf("failed to draw %d cards from level 1 pool: %v", count, err)
	}

	// 将卡牌添加到玩家手牌
	successCount := 0
	for _, card := range cards {
		err = room.AddCardToPlayer(username, card)
		if err != nil {
			// 如果添加失败，将已抽取但未添加的卡牌放回卡牌池
			for i := successCount; i < len(cards); i++ {
				room.Level1CardPool = append(room.Level1CardPool, cards[i])
			}
			return fmt.Errorf("failed to add card %s to player %s after %d successful additions: %v",
				card.Name, username, successCount, err)
		}
		successCount++
	}

	log.Printf("Drew %d cards for player %s in room %s", count, username, roomID)
	return nil
}

// GetLevel1CardPoolSize 获取指定房间一级卡牌池的大小
func (rm *RoomManager) GetLevel1CardPoolSize(roomID string) (int, error) {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		return 0, err
	}

	return room.GetLevel1CardPoolSize(), nil
}
