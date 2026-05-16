package logic

import (
	"fmt"

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
func (ccp *CardComposeProcessor) ProcessCardCompose(eventData *events.EventData) {


	// 获取玩家名称
	player, _ := eventData.GetString("player")
	// 获取房间ID
	roomID, _ := eventData.GetString("room_id")
	// 获取客户端ID
	clientID, _ := eventData.GetString("client_id")
	// 获取卡牌数据
	cardsData, _ := eventData.GetData("cards")
	// 转换为卡牌切片
	cards, ok := cardsData.([]models.Card)
	if !ok {
		return
	}

		// 构建合成数据
	data := &CardComposeData{
		RoomID:   roomID,
		Player:   player,
		Cards:    cards,
		ClientID: clientID,
	}


	// 步骤1: 验证卡牌信息
	room, validatedCardGroups, err := ccp.validateComposeRequest(data)
	if err != nil {
		return
	}

	// 步骤2: 进行合成
	composeResult := ccp.performComposition(room, validatedCardGroups)
	if !composeResult.Success {
		return
	}

	// 步骤3: 更新房间内玩家信息
	err = ccp.updatePlayerInfo(room, data.Player, &composeResult)
	if err != nil {
		return
	}
	// 步骤4: 发布游戏状态更新事件
	ccp.publishComposeResult(room)
}

// performComposition 执行卡牌合成逻辑
func (ccp *CardComposeProcessor) performComposition(room *types.RoomInfo, cardGroups map[string][]models.Card) ComposeResult {
	var removedCards []models.Card
	var newCards []models.Card
	composedCount := 0

	for _, cards := range cardGroups {
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
			continue
		}

		// 获取目标卡牌信息
		targetCard, err := ccp.cardPoolManager.GetCardByName(*templateCard.TargetName)
		if err != nil {
			continue
		}
		// 合成指定数量的新卡牌
		for i := 0; i < canCompose; i++ {
			// 选择3张要移除的卡牌
			startIdx := i * 3
			endIdx := startIdx + 3
			cardsToRemove := cards[startIdx:endIdx]
			removedCards = append(removedCards, cardsToRemove...)

			// 从对应等级的卡牌池中抽取目标卡牌
			drawnCard, err := room.DrawCardByNameFromPool(*templateCard.TargetName, targetCard.Level)
			if err != nil {

				// 如果抽取失败，将已移除的卡牌放回原组
				// 这里简化处理，可以根据需要实现更复杂的回滚逻辑
				continue
			}

			newCards = append(newCards, *drawnCard)
			composedCount++

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
func (ccp *CardComposeProcessor) publishComposeResult(room *types.RoomInfo) {

	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "card_compose_processor", map[string]interface{}{})
	stateUpdateData.SetRoom(room.RoomID)
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
}

// validateComposeRequest 验证合成请求信息
func (ccp *CardComposeProcessor) validateComposeRequest(data *CardComposeData) (*types.RoomInfo, map[string][]models.Card, error) {
	// 获取房间管理器
	roomManager := service.GetRoomManager()

	// 获取房间信息
	room, err := roomManager.GetRoom(data.RoomID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get room %s: %v", data.RoomID, err)
	}

	// 验证房间状态
	if room.Status != "playing" {
		return nil, nil, fmt.Errorf("room %s is not in playing state: %s", data.RoomID, room.Status)
	}

	// 获取玩家信息
	playerInfo, err := room.GetPlayerInfo(data.Player)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get player info for %s: %v", data.Player, err)
	}

	// 验证卡牌数量（必须是3的倍数）
	if len(data.Cards)%3 != 0 {
		return nil, nil, fmt.Errorf("invalid card count for composition: %d (must be multiple of 3)", len(data.Cards))
	}

	// 构建手牌UID映射用于快速查找和验证
	handCardMap := make(map[string]models.Card)
	for _, handCard := range playerInfo.HandCards {
		handCardMap[handCard.UID] = handCard
	}

	// 验证所有要合成的卡牌都在玩家手牌中
	var validatedCards []models.Card
	for _, cardToCompose := range data.Cards {
		if handCard, exists := handCardMap[cardToCompose.UID]; exists {
			// 验证卡牌详细信息匹配
			if handCard.Name != cardToCompose.Name || handCard.ID != cardToCompose.ID {
				return nil, nil, fmt.Errorf("card information mismatch for UID %s: expected %s (ID: %d), got %s (ID: %d)",
					cardToCompose.UID, handCard.Name, handCard.ID, cardToCompose.Name, cardToCompose.ID)
			}
			validatedCards = append(validatedCards, cardToCompose)
		} else {
			return nil, nil, fmt.Errorf("card UID %s not found in player %s's hand", cardToCompose.UID, data.Player)
		}
	}

	// 按照卡牌名称分组
	cardGroups := make(map[string][]models.Card)
	for _, card := range validatedCards {
		cardGroups[card.Name] = append(cardGroups[card.Name], card)
	}

	return room, cardGroups, nil
}

// updatePlayerInfo 更新玩家信息（移除旧卡牌，添加新卡牌）
func (ccp *CardComposeProcessor) updatePlayerInfo(room *types.RoomInfo, playerName string, result *ComposeResult) error {
	// 1. 从玩家手牌中移除已合成的卡牌
	var cardUIDs []string
	for _, card := range result.RemovedCards {
		cardUIDs = append(cardUIDs, card.UID)
	}

	err := room.RemoveCardsFromPlayerByUID(playerName, cardUIDs)
	if err != nil {
		return fmt.Errorf("failed to remove composed cards from player %s: %v", playerName, err)
	}

	// 2. 将新合成的卡牌添加到玩家手牌
	for _, newCard := range result.NewCards {
		err = room.AddCardToPlayer(playerName, newCard)
		if err != nil {
			// 如果添加失败，需要考虑回滚已移除的卡牌
			return fmt.Errorf("failed to add new card %s to player %s: %v", newCard.Name, playerName, err)
		}
	}

		return nil
}
