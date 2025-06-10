package types

import (
	"GoServer/tcpgameserver/models"
	"fmt"
	"math/rand"
	"sync"
)

// PlayerInfo 房间内玩家信息
type PlayerInfo struct {
	Username      string                       `json:"username"`       // 玩家用户名
	HandCards     []models.Card                `json:"hand_cards"`     // 手牌列表
	MaxHealth     float64                      `json:"max_health"`     // 总血量
	CurrentHealth float64                      `json:"current_health"` // 当前血量
	IsReady       bool                         `json:"is_ready"`       // 是否准备就绪
	Round         string                       `json:"round"`          // 是否是当前回合玩家
	OtherPlayers  []models.OtherPlayerGameInfo `json:"OtherPlayer"`    // 其他玩家信息
	DamageInfo    []models.DamageInfo          `json:"DamageInfo"`     // 伤害信息列表
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
	InitialHealth float64 `json:"initial_health"` // 初始血量
	MaxHandCards  int     `json:"max_hand_cards"` // 最大手牌数量

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
		InitialHealth:  10, // 默认初始血量
		MaxHandCards:   10, // 默认最大手牌数量
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
		IsReady:       true,
		Round:         "waiting",
		OtherPlayers:  make([]models.OtherPlayerGameInfo, 0),
		DamageInfo:    make([]models.DamageInfo, 0),
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
func (r *RoomInfo) GetPlayerCurrentHealth(username string) (float64, error) {
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
		HandCards:     player.HandCards,
		MaxHealth:     player.MaxHealth,
		CurrentHealth: player.CurrentHealth,
		Round:         player.Round,
		IsReady:       player.IsReady,
	}
	copy(playerCopy.HandCards, player.HandCards)

	return playerCopy, nil
}

// SetPlayerHealth 设置玩家血量
func (r *RoomInfo) SetPlayerHealth(username string, health float64) error {
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

// SetOpponentPlayerDamage 设置其他玩家触发的伤害信息
func (r *RoomInfo) SetPlayerDamage(username string, DamageInfo models.DamageInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	player.DamageInfo = append(player.DamageInfo, DamageInfo)
	return nil
}

// SetOpponentPlayerDamage 设置其他玩家触发的伤害信息
func (r *RoomInfo) CleanPlayerDamage(username string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	player.DamageInfo = []models.DamageInfo{} // 清空伤害信息
	return nil
}

// GetPlayerBonds 获取玩家触发的伤害信息
func (r *RoomInfo) GetPlayerBonds(username string) (damageInfo []models.DamageInfo, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return nil, fmt.Errorf("player %s not found in room", username)
	}

	return player.DamageInfo, nil
}

// SetPlayerRound 设置玩家回合状态
func (r *RoomInfo) SetPlayerRound(username string, round string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	// 设置玩家的回合状态
	player.Round = round
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

// RemoveCardsFromPlayerByUID 从玩家手牌中移除指定UID的卡牌
func (r *RoomInfo) RemoveCardsFromPlayerByUID(username string, cardUIDs []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	player, exists := r.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	// 创建UID映射用于快速查找
	uidMap := make(map[string]bool)
	for _, uid := range cardUIDs {
		uidMap[uid] = true
	}

	// 过滤出不在移除列表中的卡牌
	var remainingCards []models.Card
	removedCount := 0

	for _, card := range player.HandCards {
		if uidMap[card.UID] {
			removedCount++
		} else {
			remainingCards = append(remainingCards, card)
		}
	}

	// 检查是否所有要移除的卡牌都找到了
	if removedCount != len(cardUIDs) {
		return fmt.Errorf("could not find all cards to remove: expected %d, found %d", len(cardUIDs), removedCount)
	}

	// 更新玩家手牌
	player.HandCards = remainingCards
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

// DrawRandomCardFromLevel1Pool 从一级卡牌池随机抽取一张卡牌
func (r *RoomInfo) DrawRandomCardFromLevel1Pool() (*models.Card, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.Level1CardPool) == 0 {
		return nil, fmt.Errorf("level 1 card pool is empty")
	}

	// 随机选择一个索引
	randomIndex := rand.Intn(len(r.Level1CardPool))

	// 获取选中的卡牌
	selectedCard := r.Level1CardPool[randomIndex]

	// 从卡牌池中移除该卡牌
	r.Level1CardPool = append(r.Level1CardPool[:randomIndex], r.Level1CardPool[randomIndex+1:]...)

	return &selectedCard, nil
}

// DrawRandomCardsFromLevel1Pool 从一级卡牌池随机抽取多张卡牌
func (r *RoomInfo) DrawRandomCardsFromLevel1Pool(count int) ([]models.Card, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	if len(r.Level1CardPool) < count {
		return nil, fmt.Errorf("not enough cards in level 1 pool: requested %d, available %d", count, len(r.Level1CardPool))
	}

	var drawnCards []models.Card

	// 抽取指定数量的卡牌
	for i := 0; i < count; i++ {
		if len(r.Level1CardPool) == 0 {
			break
		}

		// 随机选择一个索引
		randomIndex := rand.Intn(len(r.Level1CardPool))

		// 获取选中的卡牌
		selectedCard := r.Level1CardPool[randomIndex]
		drawnCards = append(drawnCards, selectedCard)

		// 从卡牌池中移除该卡牌
		r.Level1CardPool = append(r.Level1CardPool[:randomIndex], r.Level1CardPool[randomIndex+1:]...)
	}

	return drawnCards, nil
}

// DrawRandomCardsFromLevel2Pool 从二级卡牌池随机抽取多张卡牌
func (r *RoomInfo) DrawRandomCardsFromLevel2Pool(count int) ([]models.Card, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	if len(r.Level2CardPool) < count {
		return nil, fmt.Errorf("not enough cards in level 2 pool: requested %d, available %d", count, len(r.Level2CardPool))
	}

	var drawnCards []models.Card

	// 抽取指定数量的卡牌
	for i := 0; i < count; i++ {
		if len(r.Level2CardPool) == 0 {
			break
		}

		// 随机选择一个索引
		randomIndex := rand.Intn(len(r.Level2CardPool))

		// 获取选中的卡牌
		selectedCard := r.Level2CardPool[randomIndex]
		drawnCards = append(drawnCards, selectedCard)

		// 从卡牌池中移除该卡牌
		r.Level2CardPool = append(r.Level2CardPool[:randomIndex], r.Level2CardPool[randomIndex+1:]...)
	}

	return drawnCards, nil
}

// DrawRandomCardsFromLevel3Pool 从三级卡牌池随机抽取多张卡牌
func (r *RoomInfo) DrawRandomCardsFromLevel3Pool(count int) ([]models.Card, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	if len(r.Level3CardPool) < count {
		return nil, fmt.Errorf("not enough cards in level 3 pool: requested %d, available %d", count, len(r.Level3CardPool))
	}

	var drawnCards []models.Card

	// 抽取指定数量的卡牌
	for i := 0; i < count; i++ {
		if len(r.Level3CardPool) == 0 {
			break
		}

		// 随机选择一个索引
		randomIndex := rand.Intn(len(r.Level3CardPool))

		// 获取选中的卡牌
		selectedCard := r.Level3CardPool[randomIndex]
		drawnCards = append(drawnCards, selectedCard)

		// 从卡牌池中移除该卡牌
		r.Level3CardPool = append(r.Level3CardPool[:randomIndex], r.Level3CardPool[randomIndex+1:]...)
	}

	return drawnCards, nil
}

// DrawCardByNameFromPool 根据卡牌名称从指定等级的卡牌池中抽取一张卡牌
func (r *RoomInfo) DrawCardByNameFromPool(cardName string, level int) (*models.Card, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var cardPool *[]models.Card
	var poolName string

	switch level {
	case 1:
		cardPool = &r.Level1CardPool
		poolName = "level 1"
	case 2:
		cardPool = &r.Level2CardPool
		poolName = "level 2"
	case 3:
		cardPool = &r.Level3CardPool
		poolName = "level 3"
	default:
		return nil, fmt.Errorf("invalid card level: %d", level)
	}

	// 在指定等级的卡牌池中查找匹配的卡牌
	for i, card := range *cardPool {
		if card.Name == cardName {
			// 找到匹配的卡牌，从池中移除并返回
			selectedCard := card
			*cardPool = append((*cardPool)[:i], (*cardPool)[i+1:]...)
			return &selectedCard, nil
		}
	}

	return nil, fmt.Errorf("card '%s' not found in %s pool", cardName, poolName)
}
