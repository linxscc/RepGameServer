package logic

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
)

// ============================================================
// 机器人对局
// ------------------------------------------------------------
// 1. 玩家开始匹配后调用 ScheduleBotMatch；若 bot.yaml 的
//    match_wait_seconds 内没有匹配到真人对手，就为其安排机器人。
// 2. 机器人只存在于房间中（没有网络连接）。轮到机器人时，回合调度器
//    （room_timer_processor）在 think_delay_ms 后调用 PlayBotTurn。
// 3. 机器人出牌走与真人完全相同的出牌流程（EventCardPlay），
//    由服务器统一校验和结算，不存在特殊规则。
// ============================================================

var botSeq int64

// NewBotName 生成本次运行内唯一的机器人名称（前缀见 bot.yaml，真人不能注册该前缀）
func NewBotName() string {
	n := atomic.AddInt64(&botSeq, 1)
	return fmt.Sprintf("%s%04d", config.Game().Bot.NamePrefix, 1000+n%9000)
}

// IsBot 判断玩家是否为机器人
func IsBot(username string) bool {
	return strings.HasPrefix(username, config.Game().Bot.NamePrefix)
}

// botMatchTimers 等待机器人匹配的玩家，clientID -> 计时器
var (
	botMatchMu     sync.Mutex
	botMatchTimers = make(map[string]*time.Timer)
)

// ScheduleBotMatch 玩家开始匹配时调用：等待一段时间后若仍未匹配到真人，则安排机器人对局
func ScheduleBotMatch(clientID string) {
	bot := config.Game().Bot
	if !bot.Enabled {
		return
	}

	botMatchMu.Lock()
	defer botMatchMu.Unlock()

	if old, ok := botMatchTimers[clientID]; ok {
		old.Stop()
	}
	botMatchTimers[clientID] = time.AfterFunc(bot.MatchWait(), func() {
		botMatchMu.Lock()
		delete(botMatchTimers, clientID)
		botMatchMu.Unlock()

		// StartBotMatch 内部会再次确认玩家仍在等待匹配
		if err := NewGameStartProcessor().StartBotMatch(clientID); err != nil {
			log.Printf("start bot match for %s: %v", clientID, err)
		}
	})
}

// PlayBotTurn 机器人出牌：选择伤害最高的出法，通过正常出牌流程提交
func PlayBotTurn(roomID, botName string) {
	room, err := service.GetRoomManager().GetRoom(roomID)
	if err != nil || room.Status != "playing" {
		return
	}
	player, err := room.GetPlayerInfo(botName)
	if err != nil || player.Round != "current" {
		return
	}

	hand, err := room.GetPlayerHandCards(botName)
	if err != nil || len(hand) == 0 {
		return
	}

	cards := ChooseBotPlay(hand, config.Game().Bot.MaxCardsPerPlay, NewBondCalculator().CalculateBondDamage)
	events.Publish(events.EventCardPlay, events.NewEventData(events.EventCardPlay, "bot_player", map[string]interface{}{
		"player":     botName,
		"self_cards": cards,
		"room_id":    roomID,
	}))
}

// ChooseBotPlay 从手牌中选出要打出的卡牌。
// 候选出法：每张单卡、每组同名卡牌、伤害最高的若干张；用羁绊计算器评估总伤害，
// 取伤害最高者，伤害相同则选用牌更少的。maxCards 限制一次出牌数量。
func ChooseBotPlay(hand []models.Card, maxCards int, score func([]models.Card) BondCalculationResult) []models.Card {
	if maxCards < 1 {
		maxCards = 1
	}

	var candidates [][]models.Card

	// 单张
	for _, c := range hand {
		candidates = append(candidates, []models.Card{c})
	}

	// 同名组合（最多 maxCards 张）
	byName := make(map[string][]models.Card)
	var names []string
	for _, c := range hand {
		if _, ok := byName[c.Name]; !ok {
			names = append(names, c.Name)
		}
		byName[c.Name] = append(byName[c.Name], c)
	}
	for _, name := range names {
		if group := byName[name]; len(group) > 1 {
			candidates = append(candidates, limit(group, maxCards))
		}
	}

	// 伤害最高的 maxCards 张（可能凑出跨名称的羁绊）
	sorted := append([]models.Card(nil), hand...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Damage > sorted[j].Damage })
	candidates = append(candidates, limit(sorted, maxCards))

	best := candidates[0]
	bestDamage := score(best).TotalDamage
	for _, cand := range candidates[1:] {
		d := score(cand).TotalDamage
		if d > bestDamage || (d == bestDamage && len(cand) < len(best)) {
			best, bestDamage = cand, d
		}
	}
	return best
}

func limit(cards []models.Card, n int) []models.Card {
	if len(cards) > n {
		cards = cards[:n]
	}
	return append([]models.Card(nil), cards...)
}
