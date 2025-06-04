package types

import (
	"GoServer/tcpgameserver/models"
	"fmt"
	"sync"
)

// PlayerInfo 房间内玩家信息
type PlayerInfo struct {
	Username       string        `json:"username"`        // 玩家用户名
	HandCards      []models.Card `json:"hand_cards"`      // 手牌列表
	MaxHealth      int           `json:"max_health"`      // 总血量
	CurrentHealth  int           `json:"current_health"`  // 当前血量
	DamageDealt    float64       `json:"damage_dealt"`    // 造成的伤害
	DamageReceived float64       `json:"damage_received"` // 承受的伤害
	IsReady        bool          `json:"is_ready"`        // 是否准备就绪
}

// RoomInfo 游戏房间信息
type RoomInfo struct {
	// 房间基本信息
	RoomID     string `json:"room_id"`     // 房间ID
	RoomName   string `json:"room_name"`   // 房间名称
	MaxPlayers int    `json:"max_players"` // 最大玩家数量
	Status     string `json:"status"`      // 房间状态：waiting, ready, playing, finished

	// 玩家信息
	Players map[string]*PlayerInfo `json:"players"` // 玩家列表，key为username

	// 共享卡牌池
	Level1CardPool []models.Card `json:"level1_card_pool"` // 1级共享卡牌池
	Level2CardPool []models.Card `json:"level2_card_pool"` // 2级共享卡牌池
	Level3CardPool []models.Card `json:"level3_card_pool"` // 3级共享卡牌池

	// 游戏设置
	InitialHealth int `json:"initial_health"` // 初始血量
	MaxHandCards  int `json:"max_hand_cards"` // 最大手牌数量

	// 内部使用
	mutex sync.RWMutex `json:"-"` // 读写锁
}

// NewRoomInfo 创建新的房间信息
func NewRoomInfo(roomID, roomName string, maxPlayers int) *RoomInfo {
	return &RoomInfo{
		RoomID:         roomID,
		RoomName:       roomName,
		MaxPlayers:     maxPlayers,
		Status:         "waiting",
		Players:        make(map[string]*PlayerInfo),
		Level1CardPool: make([]models.Card, 0),
		Level2CardPool: make([]models.Card, 0),
		Level3CardPool: make([]models.Card, 0),
		InitialHealth:  100, // 默认初始血量
		MaxHandCards:   7,   // 默认最大手牌数量
	}
}

// InitializeCardPools 初始化房间卡牌池（从传入的卡牌池复制）
func (r *RoomInfo) InitializeCardPools(level1Cards, level2Cards, level3Cards []models.Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 复制传入的卡牌池
	r.Level1CardPool = make([]models.Card, len(level1Cards))
	copy(r.Level1CardPool, level1Cards)

	r.Level2CardPool = make([]models.Card, len(level2Cards))
	copy(r.Level2CardPool, level2Cards)

	r.Level3CardPool = make([]models.Card, len(level3Cards))
	copy(r.Level3CardPool, level3Cards)

	return nil
}

// AddPlayer 添加玩家到房间
func (r *RoomInfo) AddPlayer(username string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查房间是否已满
	if len(r.Players) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	// 检查玩家是否已存在
	if _, exists := r.Players[username]; exists {
		return fmt.Errorf("player %s already in room", username)
	}

	// 创建玩家信息
	player := &PlayerInfo{
		Username:      username,
		HandCards:     make([]models.Card, 0),
		MaxHealth:     r.InitialHealth,
		CurrentHealth: r.InitialHealth,
		IsReady:       false,
	}

	r.Players[username] = player
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

// GetPlayerHandCards 获取指定玩家的手牌列表
func (r *RoomInfo) GetPlayerHandCards(username string) ([]models.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, exists := r.Players[username]
	if !exists {
		return nil, fmt.Errorf("player %s not found in room", username)
	}

	// 返回手牌副本
	handCards := make([]models.Card, len(player.HandCards))
	copy(handCards, player.HandCards)
	return handCards, nil
}

// GetPlayerCurrentHealth 获取指定玩家的当前血量
func (r *RoomInfo) GetPlayerCurrentHealth(username string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, exists := r.Players[username]
	if !exists {
		return 0, fmt.Errorf("player %s not found in room", username)
	}

	return player.CurrentHealth, nil
}

// GetPlayerInfo 获取指定玩家的完整信息
func (r *RoomInfo) GetPlayerInfo(username string) (*PlayerInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, exists := r.Players[username]
	if !exists {
		return nil, fmt.Errorf("player %s not found in room", username)
	}

	// 返回玩家信息副本
	playerCopy := &PlayerInfo{
		Username:      player.Username,
		HandCards:     make([]models.Card, len(player.HandCards)),
		MaxHealth:     player.MaxHealth,
		CurrentHealth: player.CurrentHealth,
		IsReady:       player.IsReady,
	}
	copy(playerCopy.HandCards, player.HandCards)

	return playerCopy, nil
}

// SetPlayerHealth 设置玩家血量
func (r *RoomInfo) SetPlayerHealth(username string, health int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	// 确保血量不超过最大值
	if health > player.MaxHealth {
		health = player.MaxHealth
	}
	if health < 0 {
		health = 0
	}

	player.CurrentHealth = health
	return nil
}

// AddCardToPlayer 给玩家添加手牌
func (r *RoomInfo) AddCardToPlayer(username string, card models.Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	// 检查手牌数量限制
	if len(player.HandCards) >= r.MaxHandCards {
		return fmt.Errorf("player %s hand is full", username)
	}

	player.HandCards = append(player.HandCards, card)
	return nil
}

// RemoveCardFromPlayer 从玩家手牌中移除卡牌
func (r *RoomInfo) RemoveCardFromPlayer(username string, cardIndex int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	if cardIndex < 0 || cardIndex >= len(player.HandCards) {
		return fmt.Errorf("invalid card index %d", cardIndex)
	}

	// 移除指定索引的卡牌
	player.HandCards = append(player.HandCards[:cardIndex], player.HandCards[cardIndex+1:]...)
	return nil
}

// SetPlayerReady 设置玩家准备状态
func (r *RoomInfo) SetPlayerReady(username string, ready bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	player.IsReady = ready
	return nil
}

// GetRoomStats 获取房间统计信息
func (r *RoomInfo) GetRoomStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	readyCount := 0
	for _, player := range r.Players {
		if player.IsReady {
			readyCount++
		}
	}

	return map[string]interface{}{
		"room_id":         r.RoomID,
		"room_name":       r.RoomName,
		"status":          r.Status,
		"current_players": len(r.Players),
		"max_players":     r.MaxPlayers,
		"ready_players":   readyCount,
		"level1_cards":    len(r.Level1CardPool),
		"level2_cards":    len(r.Level2CardPool),
		"level3_cards":    len(r.Level3CardPool),
	}
}

// GetSharedCardPool 获取指定等级的共享卡牌池
func (r *RoomInfo) GetSharedCardPool(level int) ([]models.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var cardPool []models.Card
	switch level {
	case 1:
		cardPool = make([]models.Card, len(r.Level1CardPool))
		copy(cardPool, r.Level1CardPool)
	case 2:
		cardPool = make([]models.Card, len(r.Level2CardPool))
		copy(cardPool, r.Level2CardPool)
	case 3:
		cardPool = make([]models.Card, len(r.Level3CardPool))
		copy(cardPool, r.Level3CardPool)
	default:
		return nil, fmt.Errorf("invalid card level: %d", level)
	}

	return cardPool, nil
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

// IsAllPlayersReady 检查是否所有玩家都准备就绪
func (r *RoomInfo) IsAllPlayersReady() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.Players) == 0 {
		return false
	}

	for _, player := range r.Players {
		if !player.IsReady {
			return false
		}
	}
	return true
}
