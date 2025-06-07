package logic

import (
	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/models"
	"sort"
)

// BondCalculationResult 羁绊计算结果
type BondCalculationResult struct {
	TotalDamage    float64         `json:"total_damage"`    // 总伤害值
	TriggeredBonds []TriggeredBond `json:"triggered_bonds"` // 触发的羁绊列表
	UnusedCards    []models.Card   `json:"unused_cards"`    // 未参与羁绊的卡牌
	UsedCards      []models.Card   `json:"used_cards"`      // 参与羁绊的卡牌
}

// TriggeredBond 触发的羁绊信息
type TriggeredBond struct {
	Bond       *models.BondModel `json:"bond"`        // 羁绊信息
	UsedCards  []models.Card     `json:"used_cards"`  // 该羁绊使用的卡牌
	BondDamage float64           `json:"bond_damage"` // 羁绊伤害值
}

// BondCalculator 羁绊计算器
type BondCalculator struct {
	bondPoolManager *cards.BondPoolManager
}

// NewBondCalculator 创建新的羁绊计算器
func NewBondCalculator() *BondCalculator {
	return &BondCalculator{
		bondPoolManager: cards.GetBondPoolManager(),
	}
}

// CalculateBondDamage 计算羁绊伤害
// cards: 玩家出的手牌列表
// 返回: 计算结果，包含总伤害值和触发的羁绊信息
func (bc *BondCalculator) CalculateBondDamage(cards []models.Card) BondCalculationResult {
	if len(cards) == 0 {
		return BondCalculationResult{
			TotalDamage:    0,
			TriggeredBonds: []TriggeredBond{},
			UnusedCards:    []models.Card{},
			UsedCards:      []models.Card{},
		}
	}

	// 获取所有可用的羁绊
	allBonds := bc.bondPoolManager.GetAllBonds()

	// 找出所有可能的羁绊组合
	possibleBonds := bc.findPossibleBonds(cards, allBonds)

	// 如果没有可能的羁绊，直接计算基础伤害
	if len(possibleBonds) == 0 {
		totalDamage := bc.calculateBaseDamage(cards)
		return BondCalculationResult{
			TotalDamage:    totalDamage,
			TriggeredBonds: []TriggeredBond{},
			UnusedCards:    cards,
			UsedCards:      []models.Card{},
		}
	}

	// 计算最优羁绊组合
	bestCombination := bc.findOptimalBondCombination(cards, possibleBonds)

	return bestCombination
}

// findPossibleBonds 找出所有可能触发的羁绊
func (bc *BondCalculator) findPossibleBonds(cards []models.Card, allBonds map[int]*models.BondModel) []PossibleBond {
	var possibleBonds []PossibleBond

	// 创建卡牌名称映射，用于快速查找
	cardNameCount := make(map[string][]models.Card)
	for _, card := range cards {
		cardNameCount[card.Name] = append(cardNameCount[card.Name], card)
	}

	// 检查每个羁绊是否可以触发
	for _, bond := range allBonds {
		usedCards := bc.checkBondRequirement(bond, cardNameCount)
		if len(usedCards) > 0 {
			possibleBonds = append(possibleBonds, PossibleBond{
				Bond:      bond,
				UsedCards: usedCards,
				Damage:    bond.Damage,
			})
		}
	}

	// 按伤害值降序排序，优先选择高伤害羁绊
	sort.Slice(possibleBonds, func(i, j int) bool {
		return possibleBonds[i].Damage > possibleBonds[j].Damage
	})

	return possibleBonds
}

// PossibleBond 可能触发的羁绊
type PossibleBond struct {
	Bond      *models.BondModel `json:"bond"`
	UsedCards []models.Card     `json:"used_cards"`
	Damage    float64           `json:"damage"`
}

// checkBondRequirement 检查是否满足羁绊要求
func (bc *BondCalculator) checkBondRequirement(bond *models.BondModel, cardNameCount map[string][]models.Card) []models.Card {
	var usedCards []models.Card
	tempCardCount := make(map[string][]models.Card)

	// 复制卡牌映射
	for name, cards := range cardNameCount {
		tempCardCount[name] = make([]models.Card, len(cards))
		copy(tempCardCount[name], cards)
	}

	// 检查羁绊所需的每张卡牌
	for _, requiredCardName := range bond.CardNames {
		availableCards, exists := tempCardCount[requiredCardName]
		if !exists || len(availableCards) == 0 {
			// 缺少必需的卡牌，无法触发羁绊
			return []models.Card{}
		}

		// 使用一张该名称的卡牌
		usedCard := availableCards[0]
		usedCards = append(usedCards, usedCard)
		tempCardCount[requiredCardName] = availableCards[1:]
	}

	return usedCards
}

// findOptimalBondCombination 找出最优的羁绊组合
func (bc *BondCalculator) findOptimalBondCombination(cards []models.Card, possibleBonds []PossibleBond) BondCalculationResult {
	bestResult := BondCalculationResult{
		TotalDamage:    bc.calculateBaseDamage(cards),
		TriggeredBonds: []TriggeredBond{},
		UnusedCards:    cards,
		UsedCards:      []models.Card{},
	}
	// 使用贪心算法，优先选择高伤害羁绊
	usedCardUIDs := make(map[string]bool)
	var triggeredBonds []TriggeredBond
	var allUsedCards []models.Card
	totalBondDamage := 0.0

	for _, possibleBond := range possibleBonds {
		// 检查此羁绊的卡牌是否已被使用
		canUse := true
		for _, card := range possibleBond.UsedCards {
			if usedCardUIDs[card.UID] {
				canUse = false
				break
			}
		}

		if canUse {
			// 标记这些卡牌为已使用
			for _, card := range possibleBond.UsedCards {
				usedCardUIDs[card.UID] = true
				allUsedCards = append(allUsedCards, card)
			}

			// 添加触发的羁绊
			triggeredBonds = append(triggeredBonds, TriggeredBond{
				Bond:       possibleBond.Bond,
				UsedCards:  possibleBond.UsedCards,
				BondDamage: possibleBond.Damage,
			})

			totalBondDamage += possibleBond.Damage
		}
	}

	// 计算未使用卡牌的基础伤害
	var unusedCards []models.Card
	unusedBaseDamage := 0.0

	for _, card := range cards {
		if !usedCardUIDs[card.UID] {
			unusedCards = append(unusedCards, card)
			unusedBaseDamage += card.Damage
		}
	}

	// 计算总伤害
	totalDamage := totalBondDamage + unusedBaseDamage

	// 如果羁绊组合的总伤害更高，则使用羁绊组合
	if totalDamage > bestResult.TotalDamage {
		bestResult = BondCalculationResult{
			TotalDamage:    totalDamage,
			TriggeredBonds: triggeredBonds,
			UnusedCards:    unusedCards,
			UsedCards:      allUsedCards,
		}
	}

	return bestResult
}

// calculateBaseDamage 计算基础伤害（所有卡牌伤害之和）
func (bc *BondCalculator) calculateBaseDamage(cards []models.Card) float64 {
	totalDamage := 0.0
	for _, card := range cards {
		totalDamage += card.Damage
	}
	return totalDamage
}

// GetBondsByCards 根据卡牌获取相关羁绊（辅助方法）
func (bc *BondCalculator) GetBondsByCards(cards []models.Card) []*models.BondModel {
	var relatedBonds []*models.BondModel
	cardNames := make(map[string]bool)

	// 收集所有卡牌名称
	for _, card := range cards {
		cardNames[card.Name] = true
	}

	// 找出包含这些卡牌的羁绊
	allBonds := bc.bondPoolManager.GetAllBonds()
	for _, bond := range allBonds {
		hasRelatedCard := false
		for _, cardName := range bond.CardNames {
			if cardNames[cardName] {
				hasRelatedCard = true
				break
			}
		}
		if hasRelatedCard {
			relatedBonds = append(relatedBonds, bond)
		}
	}

	return relatedBonds
}

// ValidateBondRequirements 验证羁绊要求（调试用）
func (bc *BondCalculator) ValidateBondRequirements(cards []models.Card, bondID int) (bool, []string) {
	bond, exists := bc.bondPoolManager.GetBondByID(bondID)
	if !exists {
		return false, []string{"羁绊不存在"}
	}

	cardNameCount := make(map[string]int)
	for _, card := range cards {
		cardNameCount[card.Name]++
	}

	var missingCards []string
	for _, requiredCard := range bond.CardNames {
		if cardNameCount[requiredCard] == 0 {
			missingCards = append(missingCards, requiredCard)
		} else {
			cardNameCount[requiredCard]--
		}
	}

	return len(missingCards) == 0, missingCards
}
