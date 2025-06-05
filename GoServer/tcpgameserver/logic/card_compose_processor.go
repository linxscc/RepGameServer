package logic

import (
	"fmt"
	"log"

	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// CardComposeProcessor 卡牌合成处理器
type CardComposeProcessor struct {
	Name            string
	cardPoolManager *cards.CardPoolManager
}

// NewCardComposeProcessor 创建新的卡牌合成处理器
func NewCardComposeProcessor() *CardComposeProcessor {
	return &CardComposeProcessor{
		Name:            "CardComposeProcessor",
		cardPoolManager: cards.GetCardPoolManager(),
	}
}

// CardComposeData 卡牌合成事件数据
type CardComposeData struct {
	RoomID   string        `json:"room_id"`
	Player   string        `json:"player"`
	Cards    []models.Card `json:"cards"`
	ClientID string        `json:"client_id"`
}

// ComposeResult 合成结果
type ComposeResult struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	ComposedCards []models.Card `json:"composed_cards"`
	RemovedCards  []models.Card `json:"removed_cards"`
	NewCards      []models.Card `json:"new_cards"`
}

// ProcessCardCompose 处理卡牌合成逻辑
func (ccp *CardComposeProcessor) ProcessCardCompose(data *CardComposeData) error {
	log.Printf("CardComposeProcessor: Processing card compose for player %s in room %s", data.Player, data.RoomID)

	// 获取房间管理器
	roomManager := service.GetRoomManager()

	// 获取房间信息
	room, err := roomManager.GetRoom(data.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room %s: %v", data.RoomID, err)
	}

	// 验证房间状态
	if room.Status != "playing" {
		return fmt.Errorf("room %s is not in playing state: %s", data.RoomID, room.Status)
	}

	// 获取玩家信息
	playerInfo, err := room.GetPlayerInfo(data.Player)
	if err != nil {
		return fmt.Errorf("failed to get player info for %s: %v", data.Player, err)
	}

	// 验证卡牌数量（必须是3的倍数）
	if len(data.Cards)%3 != 0 {
		return fmt.Errorf("invalid card count for composition: %d (must be multiple of 3)", len(data.Cards))
	}

	// 构建手牌UID映射用于快速查找和验证
	handCardMap := make(map[int64]models.Card)
	for _, handCard := range playerInfo.HandCards {
		handCardMap[handCard.UID] = handCard
	}

	// 验证所有要合成的卡牌都在玩家手牌中
	var validatedCards []models.Card
	for _, cardToCompose := range data.Cards {
		if handCard, exists := handCardMap[cardToCompose.UID]; exists {
			// 验证卡牌详细信息匹配
			if handCard.Name != cardToCompose.Name || handCard.ID != cardToCompose.ID {
				return fmt.Errorf("card information mismatch for UID %d: expected %s (ID: %d), got %s (ID: %d)",
					cardToCompose.UID, handCard.Name, handCard.ID, cardToCompose.Name, cardToCompose.ID)
			}
			validatedCards = append(validatedCards, cardToCompose)
		} else {
			return fmt.Errorf("card UID %d not found in player %s's hand", cardToCompose.UID, data.Player)
		}
	}

	// 按照卡牌名称分组
	cardGroups := make(map[string][]models.Card)
	for _, card := range validatedCards {
		cardGroups[card.Name] = append(cardGroups[card.Name], card)
	}

	// 验证合成条件并执行合成
	composeResult := ccp.performComposition(cardGroups)
	if !composeResult.Success {
		return fmt.Errorf("composition failed: %s", composeResult.Message)
	}

	// 从玩家手牌中移除已合成的卡牌
	var cardUIDs []int64
	for _, card := range composeResult.RemovedCards {
		cardUIDs = append(cardUIDs, card.UID)
	}

	err = room.RemoveCardsFromPlayerByUID(data.Player, cardUIDs)
	if err != nil {
		return fmt.Errorf("failed to remove composed cards from player %s: %v", data.Player, err)
	}
	// 将新合成的卡牌添加到玩家手牌
	for _, newCard := range composeResult.NewCards {
		err = room.AddCardToPlayer(data.Player, newCard)
		if err != nil {
			return fmt.Errorf("failed to add new card %s to player: %v", newCard.Name, err)
		}
	}

	// 发布游戏状态更新事件
	ccp.publishComposeResult(room, data.Player, &composeResult)

	log.Printf("CardComposeProcessor: Successfully composed %d new cards for player %s, removed %d cards",
		len(composeResult.NewCards), data.Player, len(composeResult.RemovedCards))
	return nil
}

// performComposition 执行卡牌合成逻辑
func (ccp *CardComposeProcessor) performComposition(cardGroups map[string][]models.Card) ComposeResult {
	var removedCards []models.Card
	var newCards []models.Card
	composedCount := 0

	for cardName, cards := range cardGroups {
		// 检查是否有足够的卡牌合成（每3张合成1张）
		groupSize := len(cards)
		canCompose := groupSize / 3

		if canCompose == 0 {
			continue
		}

		// 获取第一张卡牌作为模板
		templateCard := cards[0]

		// 检查TargetName是否为空（不可合成卡牌）
		if templateCard.TargetName == nil || *templateCard.TargetName == "" {
			log.Printf("Card %s cannot be composed (TargetName is empty)", cardName)
			continue
		} // 获取目标卡牌信息
		targetCard, err := ccp.cardPoolManager.GetCardByName(*templateCard.TargetName)
		if err != nil {
			log.Printf("Failed to get target card %s for composition: %v", *templateCard.TargetName, err)
			continue
		}
		// 合成指定数量的新卡牌
		for i := 0; i < canCompose; i++ {
			// 选择3张要移除的卡牌
			startIdx := i * 3
			endIdx := startIdx + 3
			cardsToRemove := cards[startIdx:endIdx]
			removedCards = append(removedCards, cardsToRemove...)

			// 创建新的高级卡牌
			newCard := models.NewCard(
				targetCard.ID,
				targetCard.Name,
				targetCard.Damage,
				targetCard.TargetName,
				targetCard.Level,
			)
			newCards = append(newCards, newCard)
			composedCount++

			log.Printf("Composed new card: %s (Level %d) from 3x %s",
				newCard.Name, newCard.Level, cardName)
		}
	}

	if composedCount == 0 {
		return ComposeResult{
			Success: false,
			Message: "No valid compositions found",
		}
	}

	return ComposeResult{
		Success:       true,
		Message:       fmt.Sprintf("Successfully composed %d new cards", composedCount),
		ComposedCards: append(removedCards, newCards...),
		RemovedCards:  removedCards,
		NewCards:      newCards,
	}
}

// publishComposeResult 发布合成结果事件
func (ccp *CardComposeProcessor) publishComposeResult(room *types.RoomInfo, player string, result *ComposeResult) {
	// 构建移除的卡牌信息
	var removedCardsInfo []map[string]interface{}
	for _, card := range result.RemovedCards {
		cardInfo := map[string]interface{}{
			"uid":    card.UID,
			"id":     card.ID,
			"name":   card.Name,
			"damage": card.Damage,
			"level":  card.Level,
		}
		removedCardsInfo = append(removedCardsInfo, cardInfo)
	}

	// 构建新卡牌信息
	var newCardsInfo []map[string]interface{}
	for _, card := range result.NewCards {
		cardInfo := map[string]interface{}{
			"uid":    card.UID,
			"id":     card.ID,
			"name":   card.Name,
			"damage": card.Damage,
			"level":  card.Level,
		}
		newCardsInfo = append(newCardsInfo, cardInfo)
	}

	// 获取玩家状态信息
	var playersState []map[string]interface{}
	for _, p := range room.Players {
		playerState := map[string]interface{}{
			"username":   p.Username,
			"health":     p.CurrentHealth,
			"hand_count": len(p.HandCards),
			"round":      p.Round,
		}
		playersState = append(playersState, playerState)
	}

	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "card_compose_processor", map[string]interface{}{
		"room_id":        room.RoomID,
		"player":         player,
		"action":         "card_compose",
		"success":        result.Success,
		"message":        result.Message,
		"removed_cards":  removedCardsInfo,
		"new_cards":      newCardsInfo,
		"composed_count": len(result.NewCards),
		"removed_count":  len(result.RemovedCards),
		"players":        playersState,
	})
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
}
