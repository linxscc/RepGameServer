package cards

import (
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"fmt"
	"log"
	"sync"
)

// CardPoolManager 卡牌池管理器
type CardPoolManager struct {
	ALLCards    []models.Card // 卡牌模板池 - 每种卡牌类型只存储一个模板实例
	level1Cards []models.Card // 1级卡牌池
	level2Cards []models.Card // 2级卡牌池
	level3Cards []models.Card // 3级卡牌池
	mutex       sync.RWMutex  // 读写锁
}

var (
	cardPool *CardPoolManager
	once     sync.Once
)

// GetCardPoolManager 获取卡牌池管理器单例
func GetCardPoolManager() *CardPoolManager {
	once.Do(func() {
		cardPool = &CardPoolManager{
			ALLCards:    make([]models.Card, 0),
			level1Cards: make([]models.Card, 0),
			level2Cards: make([]models.Card, 0),
			level3Cards: make([]models.Card, 0),
		}
	})
	return cardPool
}

// InitCardPool 初始化卡牌池
func InitCardPool() error {
	manager := GetCardPoolManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// 从数据库获取所有卡牌配置
	cardDecks, err := service.GetAllCardDeck()
	if err != nil {
		return fmt.Errorf("failed to load card decks from database: %v", err)
	} // 清空现有卡牌池
	manager.ALLCards = make([]models.Card, 0)
	manager.level1Cards = make([]models.Card, 0)
	manager.level2Cards = make([]models.Card, 0)
	manager.level3Cards = make([]models.Card, 0)

	totalCards := 0

	// 根据CardDeck配置创建卡牌实例
	for _, deck := range cardDecks {
		// 创建卡牌模板（只创建一个实例作为模板）
		cardTemplate := models.NewCard(deck.ID, deck.Name, deck.Damage, deck.TargetName, deck.Level)

		// 将卡牌模板添加到总卡牌模板池（每种卡牌类型只存储一个模板）
		manager.ALLCards = append(manager.ALLCards, cardTemplate)

		// 根据cards_num创建对应数量的卡牌实例到各等级池中
		for i := 0; i < deck.CardsNum; i++ {
			card := models.NewCard(deck.ID, deck.Name, deck.Damage, deck.TargetName, deck.Level)

			// 根据等级分配到不同的卡牌池
			switch deck.Level {
			case 1:
				manager.level1Cards = append(manager.level1Cards, card)
			case 2:
				manager.level2Cards = append(manager.level2Cards, card)
			case 3:
				manager.level3Cards = append(manager.level3Cards, card)
			default:
				log.Printf("Warning: Unknown card level %d for card %s", deck.Level, deck.Name)
				continue
			}

			totalCards++
		}
	}

	log.Printf("Initialized card pool with %d total cards across all levels", totalCards)

	return nil
}

// GetLevel1Cards 获取1级卡牌池（副本）
func GetLevel1Cards() []models.Card {
	manager := GetCardPoolManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	cards := make([]models.Card, len(manager.level1Cards))
	copy(cards, manager.level1Cards)
	return cards
}

// GetLevel2Cards 获取2级卡牌池（副本）
func GetLevel2Cards() []models.Card {
	manager := GetCardPoolManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	cards := make([]models.Card, len(manager.level2Cards))
	copy(cards, manager.level2Cards)
	return cards
}

// GetLevel3Cards 获取3级卡牌池（副本）
func GetLevel3Cards() []models.Card {
	manager := GetCardPoolManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	cards := make([]models.Card, len(manager.level3Cards))
	copy(cards, manager.level3Cards)
	return cards
}

// GetCardPoolStats 获取卡牌池统计信息
func GetCardPoolStats() map[string]interface{} {
	manager := GetCardPoolManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	totalCount := len(manager.level1Cards) + len(manager.level2Cards) + len(manager.level3Cards)

	return map[string]interface{}{
		"level1_count": len(manager.level1Cards),
		"level2_count": len(manager.level2Cards),
		"level3_count": len(manager.level3Cards),
		"total_count":  totalCount,
	}
}

// ReloadCardPool 重新加载卡牌池
func ReloadCardPool() error {
	return InitCardPool()
}

// GetAllCards 获取所有卡牌模板（副本）- 每种卡牌类型只有一个模板实例
func GetAllCards() []models.Card {
	manager := GetCardPoolManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	cards := make([]models.Card, len(manager.ALLCards))
	copy(cards, manager.ALLCards)
	return cards
}

// GetCardByName 根据卡牌名称获取卡牌模板（从卡牌模板池中查找匹配的）
func (cpm *CardPoolManager) GetCardByName(cardName string) (*models.Card, error) {
	cpm.mutex.RLock()
	defer cpm.mutex.RUnlock()
	// 在卡牌模板池中查找匹配的卡牌模板
	for _, card := range cpm.ALLCards {
		if card.Name == cardName { // 返回卡牌模板的副本
			foundCard := models.Card{
				UID:        card.UID,
				ID:         card.ID,
				Name:       card.Name,
				Damage:     card.Damage,
				TargetName: card.TargetName,
				Level:      card.Level,
			}
			return &foundCard, nil
		}
	}

	return nil, fmt.Errorf("card with name '%s' not found in card pool", cardName)
}
