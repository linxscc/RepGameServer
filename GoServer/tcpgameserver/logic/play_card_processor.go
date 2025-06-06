package logic

import (
	"fmt"
	"log"

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
func (p *PlayCardProcessor) ProcessPlayCard(data *PlayCardData) error {
	log.Printf("PlayCardProcessor: Processing play card for player %s in room %s", data.Player, data.RoomID)

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

	// 验证玩家回合状态
	if playerInfo.Round != "first" {
		return fmt.Errorf("it's not player %s's turn (round status: %s)", data.Player, playerInfo.Round)
	}
	// 验证是否有卡牌要出
	if len(data.CardsToPlay) == 0 {
		return fmt.Errorf("no cards to play for player %s", data.Player)
	}

	// 构建手牌UID映射用于快速查找和验证
	handCardMap := make(map[int64]models.Card) // UID -> 卡牌映射
	for _, handCard := range playerInfo.HandCards {
		handCardMap[handCard.UID] = handCard
	}

	// 验证所有要出的卡牌都在玩家手牌中
	var validatedCards []models.Card
	var cardUIDs []int64

	for _, cardToPlay := range data.CardsToPlay {
		if handCard, exists := handCardMap[cardToPlay.UID]; exists {
			// 验证卡牌详细信息匹配
			if handCard.Name != cardToPlay.Name || handCard.ID != cardToPlay.ID {
				return fmt.Errorf("card information mismatch for UID %d: expected %s (ID: %d), got %s (ID: %d)",
					cardToPlay.UID, handCard.Name, handCard.ID, cardToPlay.Name, cardToPlay.ID)
			}

			validatedCards = append(validatedCards, cardToPlay)
			cardUIDs = append(cardUIDs, cardToPlay.UID)
		} else {
			return fmt.Errorf("card UID %d not found in player %s's hand", cardToPlay.UID, data.Player)
		}
	}

	// 检查是否有重复的卡牌UID
	uidSet := make(map[int64]bool)
	for _, card := range validatedCards {
		if uidSet[card.UID] {
			return fmt.Errorf("duplicate card UID %d in play request", card.UID)
		}
		uidSet[card.UID] = true
	}
	// 使用羁绊计算器计算总伤害
	bondResult := p.bondCalculator.CalculateBondDamage(validatedCards)
	totalDamage := bondResult.TotalDamage

	log.Printf("Calculated bond damage for player %s: total=%.1f, bonds=%d",
		data.Player, totalDamage, len(bondResult.TriggeredBonds))

	// 记录触发的羁绊信息
	for _, bond := range bondResult.TriggeredBonds {
		log.Printf("Triggered bond: %s (Level %d) with damage %.1f",
			bond.Bond.Name, bond.Bond.Level, bond.BondDamage)
	}

	// 根据目标类型执行伤害效果
	err = p.executeCardEffectWithBondDamage(room, data.Player, totalDamage, data.TargetType, &bondResult)
	if err != nil {
		return fmt.Errorf("failed to execute card effect with bond damage: %v", err)
	}

	// 从手牌中移除所有出的卡牌（基于UID）
	err = room.RemoveCardsFromPlayerByUID(data.Player, cardUIDs)
	if err != nil {
		return fmt.Errorf("failed to remove cards from player %s: %v", data.Player, err)
	}

	// 切换到下一个玩家的回合
	err = p.switchToNextPlayer(room, data.Player)
	if err != nil {
		return fmt.Errorf("failed to switch to next player: %v", err)
	}

	// 检查游戏是否结束
	p.checkGameEnd(room)
	// 发布游戏状态更新事件
	p.publishGameStateUpdateWithBonds(room, data.Player, validatedCards, &bondResult)
	log.Printf("PlayCardProcessor: Successfully processed play card for player %s, played %d cards, total damage: %.1f, triggered %d bonds",
		data.Player, len(validatedCards), bondResult.TotalDamage, len(bondResult.TriggeredBonds))
	return nil
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

// executeCardEffect 执行卡牌效果
func (p *PlayCardProcessor) executeCardEffect(room *types.RoomInfo, playerName string, playedCard *models.Card, targetType string) error {
	switch targetType {
	case "opponent":
		return p.applyDamageToOpponent(room, playerName, playedCard.Damage)
	case "self":
		return p.applySelfEffect(room, playerName, playedCard)
	case "all":
		return p.applyEffectToAll(room, playerName, playedCard)
	default:
		// 默认对对手造成伤害
		return p.applyDamageToOpponent(room, playerName, playedCard.Damage)
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

	log.Printf("Applied %.1f damage to opponent %s (health: %d -> %d)", damage, opponentName, currentHealth, newHealth)
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

	log.Printf("Healed player %s for %.1f points (health: %d -> %d)", playerName, healAmount, currentHealth, newHealth)
	return nil
}

// applySelfEffect 对自己应用效果
func (p *PlayCardProcessor) applySelfEffect(room *types.RoomInfo, playerName string, playedCard *models.Card) error {
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
	newHealth := currentHealth + playedCard.Damage
	if newHealth > playerInfo.MaxHealth {
		newHealth = playerInfo.MaxHealth
	}

	// 设置新血量
	err = room.SetPlayerHealth(playerName, newHealth)
	if err != nil {
		return fmt.Errorf("failed to set player health: %v", err)
	}

	log.Printf("Healed player %s for %.1f points (health: %d -> %d)", playerName, playedCard.Damage, currentHealth, newHealth)
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
				log.Printf("Failed to get player %s health: %v", player.Username, err)
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
				log.Printf("Failed to set player %s health: %v", player.Username, err)
				continue
			}

			log.Printf("Applied %.1f AOE damage to player %s (health: %d -> %d)", damage, player.Username, currentHealth, newHealth)
		}
	}

	return nil
}

// applyEffectToAll 对所有玩家应用效果
func (p *PlayCardProcessor) applyEffectToAll(room *types.RoomInfo, playerName string, playedCard *models.Card) error {
	// 对所有玩家造成伤害（AOE效果）
	for _, player := range room.Players {
		if player.Username != playerName { // 通常AOE不影响施法者
			// 获取玩家当前血量
			currentHealth, err := room.GetPlayerCurrentHealth(player.Username)
			if err != nil {
				log.Printf("Failed to get player %s health: %v", player.Username, err)
				continue
			}

			// 计算伤害后的血量
			newHealth := currentHealth - playedCard.Damage
			if newHealth < 0 {
				newHealth = 0
			}

			// 设置新血量
			err = room.SetPlayerHealth(player.Username, newHealth)
			if err != nil {
				log.Printf("Failed to set player %s health: %v", player.Username, err)
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

	// 更新回合状态 - 直接修改玩家的Round字段
	room.Players[currentPlayer].Round = "waiting"
	room.Players[nextPlayer].Round = "current"

	log.Printf("Switched turn from %s to %s in room %s", currentPlayer, nextPlayer, room.RoomID)
	return nil
}

// checkGameEnd 检查游戏是否结束
func (p *PlayCardProcessor) checkGameEnd(room *types.RoomInfo) {
	for _, player := range room.Players {
		if player.CurrentHealth <= 0 {
			// 游戏结束
			room.Status = "finished"

			// 确定获胜者
			var winner string
			for _, p := range room.Players {
				if p.CurrentHealth > 0 {
					winner = p.Username
					break
				}
			}

			// 发布游戏结束事件
			gameEndData := events.NewEventData(events.EventGameEnd, "play_card_processor", map[string]interface{}{
				"room_id": room.RoomID,
				"winner":  winner,
				"loser":   player.Username,
			})
			events.Publish(events.EventGameEnd, gameEndData)

			log.Printf("Game ended in room %s, winner: %s", room.RoomID, winner)
			return
		}
	}
}

// publishGameStateUpdateWithBonds 发布包含羁绊信息的游戏状态更新事件
func (p *PlayCardProcessor) publishGameStateUpdateWithBonds(room *types.RoomInfo, player string, playedCards []models.Card, bondResult *BondCalculationResult) {
	// 构建出牌卡牌信息
	var playedCardsInfo []map[string]interface{}
	for _, card := range playedCards {
		cardInfo := map[string]interface{}{
			"uid":    card.UID,
			"id":     card.ID,
			"name":   card.Name,
			"damage": card.Damage,
		}
		playedCardsInfo = append(playedCardsInfo, cardInfo)
	}

	// 构建触发的羁绊信息
	var triggeredBondsInfo []map[string]interface{}
	for _, bond := range bondResult.TriggeredBonds {
		bondInfo := map[string]interface{}{
			"bond_id":     bond.Bond.ID,
			"bond_name":   bond.Bond.Name,
			"bond_level":  bond.Bond.Level,
			"bond_damage": bond.BondDamage,
			"used_cards":  len(bond.UsedCards),
		}
		triggeredBondsInfo = append(triggeredBondsInfo, bondInfo)
	}

	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "play_card_processor", map[string]interface{}{
		"room_id":         room.RoomID,
		"player":          player,
		"played_cards":    playedCardsInfo,
		"card_count":      len(playedCards),
		"total_damage":    bondResult.TotalDamage,
		"triggered_bonds": triggeredBondsInfo,
		"unused_cards":    len(bondResult.UnusedCards),
		"used_cards":      len(bondResult.UsedCards),
		"players":         p.buildPlayersState(room),
	})
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
}

// publishGameStateUpdate 发布游戏状态更新事件
func (p *PlayCardProcessor) publishGameStateUpdate(room *types.RoomInfo, player string, playedCards []models.Card) {
	// 构建出牌卡牌信息
	var playedCardsInfo []map[string]interface{}
	for _, card := range playedCards {
		cardInfo := map[string]interface{}{
			"uid":    card.UID,
			"id":     card.ID,
			"name":   card.Name,
			"damage": card.Damage,
		}
		playedCardsInfo = append(playedCardsInfo, cardInfo)
	}

	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "play_card_processor", map[string]interface{}{
		"room_id":      room.RoomID,
		"player":       player,
		"played_cards": playedCardsInfo,
		"card_count":   len(playedCards),
		"players":      p.buildPlayersState(room),
	})
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
}

// buildPlayersState 构建玩家状态信息
func (p *PlayCardProcessor) buildPlayersState(room *types.RoomInfo) []map[string]interface{} {
	var playersState []map[string]interface{}

	for _, player := range room.Players {
		playerState := map[string]interface{}{
			"username":   player.Username,
			"health":     player.CurrentHealth,
			"hand_count": len(player.HandCards),
			"round":      player.Round,
		}
		playersState = append(playersState, playerState)
	}

	return playersState
}
