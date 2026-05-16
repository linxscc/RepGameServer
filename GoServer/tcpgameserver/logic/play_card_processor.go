package logic

import (
	"fmt"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// PlayCardProcessor 出牌逻辑处理器
type PlayCardProcessor struct {
	Name           string
	bondCalculator *BondCalculator
}

// NewPlayCardProcessor 创建新的出牌逻辑处理器
func NewPlayCardProcessor() *PlayCardProcessor {
	return &PlayCardProcessor{
		Name:           "PlayCardProcessor",
		bondCalculator: NewBondCalculator(),
	}
}

// PlayCardData 出牌事件数据
type PlayCardData struct {
	RoomID      string        `json:"room_id"`
	Player      string        `json:"player"`
	CardsToPlay []models.Card `json:"cards_to_play"` // 要出的所有卡牌
	TargetType  string        `json:"target_type"`
}

// ProcessPlayCard 处理出牌逻辑
func (p *PlayCardProcessor) ProcessPlayCard(eventData *events.EventData) {


	// 获取玩家名称
	player, _ := eventData.GetString("player")
	// 获取房间ID
	roomID, _ := eventData.GetString("room_id")
	// 获取玩家发送的自身卡牌数据
	selfCardsData, _ := eventData.GetData("self_cards")

	// 转换为卡牌切片
	receivedSelfCards, _ := selfCardsData.([]models.Card)

	// 构建出牌数据（所有验证交给ProcessPlayCard处理）
	data := &PlayCardData{
		RoomID:      roomID,
		Player:      player,
		CardsToPlay: receivedSelfCards,
		TargetType:  "opponent",
	}

	// 步骤1: 验证出牌信息是否正确
	room, validatedCards, err := p.validatePlayCardRequest(data)
	if err != nil {
		return
	}

	// 步骤2: 计算羁绊伤害加成，得到伤害结果和触发羁绊
	bondResult := p.bondCalculator.CalculateBondDamage(validatedCards)

	// 步骤3: 为房间内玩家更新信息（血量、收到伤害、造成伤害等）并为出牌方抽取新卡牌
	gameEnded, err := p.updateRoomPlayersInfo(room, data.Player, bondResult.TotalDamage, data.TargetType, &bondResult, validatedCards)
	if err != nil {
		return
	}

	// 步骤4: 发送游戏状态更新事件（仅在游戏未结束时）
	if !gameEnded {
		p.publishGameStateUpdateWithBonds(room)
	}
}

// validatePlayCardRequest 验证出牌请求信息
func (p *PlayCardProcessor) validatePlayCardRequest(data *PlayCardData) (*types.RoomInfo, []models.Card, error) {
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

	// 验证玩家回合状态
	if playerInfo.Round != "current" {
		return nil, nil, fmt.Errorf("it's not player %s's turn (round status: %s)", data.Player, playerInfo.Round)
	}

	// 验证是否有卡牌要出
	if len(data.CardsToPlay) == 0 {
		return nil, nil, fmt.Errorf("no cards to play for player %s", data.Player)
	}
	// 构建手牌UID映射用于快速查找和验证
	handCardMap := make(map[string]models.Card)
	for _, handCard := range playerInfo.HandCards {
		handCardMap[handCard.UID] = handCard
	}

	// 验证所有要出的卡牌都在玩家手牌中
	var validatedCards []models.Card
	for _, cardToPlay := range data.CardsToPlay {
		if handCard, exists := handCardMap[cardToPlay.UID]; exists {
			// 验证卡牌详细信息匹配
			if handCard.Name != cardToPlay.Name || handCard.ID != cardToPlay.ID {
				return nil, nil, fmt.Errorf("card information mismatch for UID %s: expected %s (ID: %d), got %s (ID: %d)",
					cardToPlay.UID, handCard.Name, handCard.ID, cardToPlay.Name, cardToPlay.ID)
			}
			validatedCards = append(validatedCards, cardToPlay)
		} else {
			return nil, nil, fmt.Errorf("card UID %s not found in player %s's hand", cardToPlay.UID, data.Player)
		}
	}

	// 检查是否有重复的卡牌UID
	uidSet := make(map[string]bool)
	for _, card := range validatedCards {
		if uidSet[card.UID] {
			return nil, nil, fmt.Errorf("duplicate card UID %s in play request", card.UID)
		}
		uidSet[card.UID] = true
	}

	return room, validatedCards, nil
}

// updateRoomPlayersInfo 更新房间内玩家信息（血量、伤害统计、羁绊、移除卡牌、切换回合、抽取新卡牌等）
// 返回值：(gameEnded bool, error) - gameEnded表示游戏是否结束
func (p *PlayCardProcessor) updateRoomPlayersInfo(room *types.RoomInfo, playerName string, totalDamage float64, targetType string, bondResult *BondCalculationResult, playedCards []models.Card) (bool, error) { // 1. 根据目标类型执行伤害效果
	err := p.executeCardEffectWithBondDamage(room, playerName, totalDamage, targetType, bondResult)
	if err != nil {
		return false, fmt.Errorf("failed to execute card effect: %v", err)
	}

	// 2. 更新玩家战斗统计数据
	err = p.updatePlayerBattleStats(room, playerName, totalDamage, targetType, bondResult)
	if err != nil {
		return false, fmt.Errorf("failed to update battle stats: %v", err)
	}
	// 3. 从手牌中移除出的卡牌
	var cardUIDs []string
	for _, card := range playedCards {
		cardUIDs = append(cardUIDs, card.UID)
	}
	err = room.RemoveCardsFromPlayerByUID(playerName, cardUIDs)
	if err != nil {
		return false, fmt.Errorf("failed to remove cards from player %s: %v", playerName, err)
	}

	// 4. 切换到下一个玩家的回合
	err = p.switchToNextPlayer(room, playerName)
	if err != nil {
		return false, fmt.Errorf("failed to switch to next player: %v", err)
	} // 5. 检查游戏是否结束
	gameEnded := p.checkGameEnd(room)
	if gameEnded {
		// 游戏已结束，不再执行后续逻辑
		return true, nil
	}

	// 6. 为出牌方新增三张卡牌（如果游戏仍在进行）
	if room.Status == "playing" {
		err = p.drawCardsForPlayer(room, playerName, 3)
		if err != nil {
		}
	}

	return false, nil
}

// executeCardEffectWithBondDamage 使用羁绊计算后的伤害执行效果
func (p *PlayCardProcessor) executeCardEffectWithBondDamage(room *types.RoomInfo, playerName string, totalDamage float64, targetType string, bondResult *BondCalculationResult) error {
	switch targetType {
	case "opponent":
		return p.applyDamageToOpponent(room, playerName, totalDamage)
	case "self":
		return p.applySelfHeal(room, playerName, totalDamage)
	case "all":
		return p.applyAOEDamage(room, playerName, totalDamage)
	default:
		// 默认对对手造成伤害
		return p.applyDamageToOpponent(room, playerName, totalDamage)
	}
}

// applyDamageToOpponent 对对手造成伤害
func (p *PlayCardProcessor) applyDamageToOpponent(room *types.RoomInfo, playerName string, damage float64) error {
	// 获取对手名称
	var opponentName string
	for _, player := range room.Players {
		if player.Username != playerName {
			opponentName = player.Username
			break
		}
	}

	if opponentName == "" {
		return fmt.Errorf("opponent not found for player %s", playerName)
	}

	// 获取对手当前血量
	currentHealth, err := room.GetPlayerCurrentHealth(opponentName)
	if err != nil {
		return fmt.Errorf("failed to get opponent health: %v", err)
	}

	// 计算伤害后的血量
	newHealth := currentHealth - damage
	if newHealth < 0 {
		newHealth = 0
	}

	// 设置对手新血量
	err = room.SetPlayerHealth(opponentName, newHealth)
	if err != nil {
		return fmt.Errorf("failed to set opponent health: %v", err)
	}

	return nil
}

// applySelfHeal 对自己应用治疗效果
func (p *PlayCardProcessor) applySelfHeal(room *types.RoomInfo, playerName string, healAmount float64) error {
	// 获取玩家当前血量
	currentHealth, err := room.GetPlayerCurrentHealth(playerName)
	if err != nil {
		return fmt.Errorf("failed to get player health: %v", err)
	}

	// 获取玩家信息以获取最大血量
	playerInfo, err := room.GetPlayerInfo(playerName)
	if err != nil {
		return fmt.Errorf("failed to get player info: %v", err)
	}

	// 计算治疗后的血量
	newHealth := currentHealth + healAmount
	if newHealth > playerInfo.MaxHealth {
		newHealth = playerInfo.MaxHealth
	}

	// 设置新血量
	err = room.SetPlayerHealth(playerName, newHealth)
	if err != nil {
		return fmt.Errorf("failed to set player health: %v", err)
	}
	return nil
}

// applyAOEDamage 对所有玩家应用AOE伤害
func (p *PlayCardProcessor) applyAOEDamage(room *types.RoomInfo, playerName string, damage float64) error {
	// 对所有玩家造成伤害（AOE效果）
	for _, player := range room.Players {
		if player.Username != playerName { // 通常AOE不影响施法者
			// 获取玩家当前血量
			currentHealth, err := room.GetPlayerCurrentHealth(player.Username)
			if err != nil {
				continue
			}

			// 计算伤害后的血量
			newHealth := currentHealth - damage
			if newHealth < 0 {
				newHealth = 0
			}

			// 设置新血量
			err = room.SetPlayerHealth(player.Username, newHealth)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// switchToNextPlayer 切换到下一个玩家
func (p *PlayCardProcessor) switchToNextPlayer(room *types.RoomInfo, currentPlayer string) error {
	// 获取所有玩家
	players := room.Players
	if len(players) != 2 {
		return fmt.Errorf("invalid number of players: %d", len(players))
	}

	// 找到下一个玩家
	var nextPlayer string
	for _, player := range players {
		if player.Username != currentPlayer {
			nextPlayer = player.Username
			break
		}
	}

	if nextPlayer == "" {
		return fmt.Errorf("next player not found")
	}
	roomManager := service.GetRoomManager()
	roomManager.SetPlayerRound(room.RoomID, currentPlayer, "waiting")
	roomManager.SetPlayerRound(room.RoomID, nextPlayer, "current")

	return nil
}

// checkGameEnd 检查游戏是否结束，返回true表示游戏已结束
func (p *PlayCardProcessor) checkGameEnd(room *types.RoomInfo) bool {
	for _, player := range room.Players {
		if player.CurrentHealth <= 0 {

			// 发布游戏结束事件
			gameEndData := events.NewEventData(events.EventGameEnd, "play_card_processor", map[string]interface{}{})
			gameEndData.SetRoom(room.RoomID)
			events.Publish(events.EventGameEnd, gameEndData)

			return true // 游戏已结束
		}
	}
	return false // 游戏继续
}

// publishGameStateUpdateWithBonds 发布包含羁绊信息的游戏状态更新事件
func (p *PlayCardProcessor) publishGameStateUpdateWithBonds(room *types.RoomInfo) {
	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "play_card_processor", map[string]interface{}{})
	stateUpdateData.SetRoom(room.RoomID)
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
}

// drawCardsForPlayer 为玩家从1级卡牌池抽取指定数量的卡牌
func (p *PlayCardProcessor) drawCardsForPlayer(room *types.RoomInfo, playerName string, count int) error {
	// 检查玩家当前手牌数量
	playerInfo, err := room.GetPlayerInfo(playerName)
	if err != nil {
		return fmt.Errorf("failed to get player info: %v", err)
	}

	// 检查是否有足够空间添加卡牌
	if len(playerInfo.HandCards)+count > room.MaxHandCards {
		availableSlots := room.MaxHandCards - len(playerInfo.HandCards)
		if availableSlots <= 0 {
			return nil // 不返回错误，只是无法抽卡
		}
		count = availableSlots // 只抽取可用槽位数量的卡牌
	}

	// 从1级卡牌池抽取卡牌
	drawnCards, err := room.DrawRandomCardsFromLevel1Pool(count)
	if err != nil {
		return fmt.Errorf("failed to draw %d cards from level 1 pool: %v", count, err)
	}

	// 将抽取的卡牌添加到玩家手牌
	successCount := 0
	for _, card := range drawnCards {
		err = room.AddCardToPlayer(playerName, card)
		if err != nil {
			// 如果添加失败，将剩余未添加的卡牌放回卡牌池
			for i := successCount; i < len(drawnCards); i++ {
				room.Level1CardPool = append(room.Level1CardPool, drawnCards[i])
			}
			return fmt.Errorf("failed to add card %s to player %s after %d successful additions: %v",
				card.Name, playerName, successCount, err)
		}
		successCount++
	}

	return nil
}

// updatePlayerBattleStats 更新双方玩家的战斗数据
func (p *PlayCardProcessor) updatePlayerBattleStats(room *types.RoomInfo, attackerName string, totalDamage float64, targetType string, bondResult *BondCalculationResult) error {
	// 将触发的羁绊转换为BondModel切片
	triggeredBondModels := make([]models.BondModel, 0, len(bondResult.TriggeredBonds))
	for _, triggeredBond := range bondResult.TriggeredBonds {
		triggeredBondModels = append(triggeredBondModels, *triggeredBond.Bond)
	}

	switch targetType {
	case "opponent":
		// 为其他玩家设置承受伤害
		for _, player := range room.Players {
			DamageInfo := models.DamageInfo{
				DamageSource:   attackerName,
				DamageTarget:   player.Username,
				DamageType:     "Attacked",
				DamageValue:    totalDamage,
				TriggeredBonds: triggeredBondModels,
			}
			room.SetPlayerDamage(player.Username, DamageInfo)
		}
	case "self":
		// 治疗情况
		for _, player := range room.Players {
			DamageInfo := models.DamageInfo{
				DamageSource:   attackerName,
				DamageTarget:   attackerName,
				DamageType:     "Recover",
				DamageValue:    totalDamage,
				TriggeredBonds: triggeredBondModels,
			}
			room.SetPlayerDamage(player.Username, DamageInfo)
		}

	case "all":
		// 为所有其他玩家设置承受伤害
		for _, player := range room.Players {
			// 更新被攻击方数据
			DamageInfo := models.DamageInfo{
				DamageSource:   attackerName,
				DamageTarget:   player.Username,
				DamageType:     "AOE",
				DamageValue:    totalDamage,
				TriggeredBonds: triggeredBondModels,
			}
			err := room.SetPlayerDamage(player.Username, DamageInfo)
			if err != nil {
				return fmt.Errorf("failed to set attacker bonds: %v", err)
			}
		}

	default:
		// 默认按对手处理
		return p.updatePlayerBattleStats(room, attackerName, totalDamage, "opponent", bondResult)
	}


	return nil
}
