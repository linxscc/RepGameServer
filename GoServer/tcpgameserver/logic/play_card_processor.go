package logic

import (
	"errors"
	"fmt"
	"log"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/types"
)

// PlayRequest 一次出牌（真人出牌、超时自动出牌、机器人出牌共用）
type PlayRequest struct {
	RoomID   string
	Player   string
	CardUIDs []string // 客户端只提供 UID，牌面数据一律取服务器手牌中的值
	// Forced 回合超时由服务器代为出牌；CardUIDs 为空时打出手牌中的第一张
	Forced bool
	// 发起请求的连接与请求序号，用于回复错误；服务器内部发起时为空
	ClientID string
	Seq      uint32
}

// errNotYourTurn 出牌时不是该玩家的回合
var errNotYourTurn = errors.New("not your turn")

// PlayCards 处理一次出牌：校验、结算、切换回合、广播，并安排下一回合。
func PlayCards(req PlayRequest) {
	err := withRoom(req.RoomID, func(room *types.RoomInfo) error {
		return defaultPlayCardProcessor.play(room, req)
	})
	if err == nil {
		return
	}

	log.Printf("reject play from %s in room %s: %v", req.Player, req.RoomID, err)
	if errors.Is(err, errNotYourTurn) {
		notifyClientError(req.ClientID, req.Seq, pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE, config.Msg(config.MsgNotYourTurn))
	} else {
		notifyClientError(req.ClientID, req.Seq, pb.ErrorCode_ERROR_CODE_INVALID_CARDS, config.Msg(config.MsgCardsInvalid))
	}
}

// PlayCardProcessor 出牌结算。调用方必须持有房间操作锁（见 room_sync.go）。
type PlayCardProcessor struct {
	bondCalculator *BondCalculator
}

var defaultPlayCardProcessor = NewPlayCardProcessor()

// NewPlayCardProcessor 创建新的出牌逻辑处理器
func NewPlayCardProcessor() *PlayCardProcessor {
	return &PlayCardProcessor{bondCalculator: NewBondCalculator()}
}

func (p *PlayCardProcessor) play(room *types.RoomInfo, req PlayRequest) error {
	cards, err := p.validate(room, req)
	if err != nil {
		return err
	}

	bondResult := p.bondCalculator.CalculateBondDamage(cards)
	if err := p.settle(room, req.Player, cards, &bondResult); err != nil {
		// 校验已保证结算各步骤成功；走到这里说明房间数据异常
		return fmt.Errorf("settle play: %w", err)
	}

	if room.AnyPlayerDefeated() {
		finishGame(room)
		return nil
	}

	reason := pb.GameStateReason_GAME_STATE_REASON_CARDS_PLAYED
	if req.Forced {
		reason = pb.GameStateReason_GAME_STATE_REASON_FORCED_PLAY
	}
	broadcastRoomState(room, reason)
	return GlobalRoomTimerProcessor.StartRoomTimer(room.RoomID)
}

// validate 校验房间状态、回合与卡牌，返回服务器手牌中对应的卡牌
func (p *PlayCardProcessor) validate(room *types.RoomInfo, req PlayRequest) ([]models.Card, error) {
	if status := room.GetStatus(); status != types.RoomStatusPlaying {
		return nil, fmt.Errorf("room is not playing: %s", status)
	}

	player, err := room.GetPlayerInfo(req.Player)
	if err != nil {
		return nil, err
	}
	if player.Round != types.RoundCurrent {
		return nil, fmt.Errorf("%w: player %s (round status: %s)", errNotYourTurn, req.Player, player.Round)
	}

	uids := req.CardUIDs
	if req.Forced && len(uids) == 0 && len(player.HandCards) > 0 {
		uids = []string{player.HandCards[0].UID}
	}
	if len(uids) == 0 {
		return nil, fmt.Errorf("no cards to play for player %s", req.Player)
	}

	hand := make(map[string]models.Card, len(player.HandCards))
	for _, c := range player.HandCards {
		hand[c.UID] = c
	}
	seen := make(map[string]bool, len(uids))
	cards := make([]models.Card, 0, len(uids))
	for _, uid := range uids {
		if seen[uid] {
			return nil, fmt.Errorf("duplicate card UID %s in play request", uid)
		}
		seen[uid] = true
		c, ok := hand[uid]
		if !ok {
			return nil, fmt.Errorf("card UID %s not found in player %s's hand", uid, req.Player)
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// settle 结算：移除出的牌 → 造成伤害 → 记录伤害信息 → 切换回合 → 补牌
func (p *PlayCardProcessor) settle(room *types.RoomInfo, playerName string, cards []models.Card, bondResult *BondCalculationResult) error {
	uids := make([]string, 0, len(cards))
	for _, c := range cards {
		uids = append(uids, c.UID)
	}
	if err := room.RemoveCardsFromPlayerByUID(playerName, uids); err != nil {
		return err
	}

	opponent := opponentOf(room, playerName)
	if opponent == "" {
		return fmt.Errorf("opponent not found for player %s", playerName)
	}
	health, err := room.GetPlayerCurrentHealth(opponent)
	if err != nil {
		return err
	}
	if err := room.SetPlayerHealth(opponent, health-bondResult.TotalDamage); err != nil {
		return err
	}

	p.recordDamage(room, playerName, opponent, bondResult)

	if err := room.SetPlayerRound(playerName, types.RoundWaiting); err != nil {
		return err
	}
	if err := room.SetPlayerRound(opponent, types.RoundCurrent); err != nil {
		return err
	}

	if !room.AnyPlayerDefeated() {
		p.drawCards(room, playerName, config.Game().Rules.Player.DrawAfterPlay)
	}
	return nil
}

// recordDamage 让房间内每个玩家都收到这次伤害的信息（用于客户端展示），广播后清空
func (p *PlayCardProcessor) recordDamage(room *types.RoomInfo, attacker, target string, bondResult *BondCalculationResult) {
	bonds := make([]models.BondModel, 0, len(bondResult.TriggeredBonds))
	for _, tb := range bondResult.TriggeredBonds {
		bonds = append(bonds, *tb.Bond)
	}
	info := models.DamageInfo{
		DamageSource:   attacker,
		DamageTarget:   target,
		DamageType:     "Attacked",
		DamageValue:    bondResult.TotalDamage,
		TriggeredBonds: bonds,
	}
	for _, name := range room.PlayerNames() {
		_ = room.SetPlayerDamage(name, info)
	}
}

// drawCards 出牌方补牌，不超过手牌上限；卡池不足时少抽或不抽
func (p *PlayCardProcessor) drawCards(room *types.RoomInfo, playerName string, count int) {
	hand, err := room.GetPlayerHandCards(playerName)
	if err != nil {
		return
	}
	if free := room.MaxHandCards - len(hand); count > free {
		count = free
	}
	if count <= 0 {
		return
	}

	drawn, err := room.DrawRandomCardsFromLevel1Pool(count)
	if err != nil {
		log.Printf("room %s: draw %d cards for %s: %v", room.RoomID, count, playerName, err)
		return
	}
	for i, c := range drawn {
		if err := room.AddCardToPlayer(playerName, c); err != nil {
			room.ReturnCardsToPool(drawn[i:])
			return
		}
	}
}

// opponentOf 双人房间中 playerName 的对手
func opponentOf(room *types.RoomInfo, playerName string) string {
	for _, name := range room.PlayerNames() {
		if name != playerName {
			return name
		}
	}
	return ""
}
