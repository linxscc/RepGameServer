package logic

import (
	"errors"
	"fmt"
	"log"

	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/types"
)

// ComposeRequest 一次合成：每 group_size 张同名卡合成 1 张目标卡（game_rules.yaml）
type ComposeRequest struct {
	RoomID   string
	Player   string
	CardUIDs []string
	ClientID string
	Seq      uint32
}

// errNothingComposed 校验通过但一张也没合成出来（例如目标卡的共享卡池已空）
var errNothingComposed = errors.New("nothing composed")

// ComposeCards 处理一次合成并广播结果；失败时回复发起请求的客户端。
func ComposeCards(req ComposeRequest) {
	err := withRoom(req.RoomID, func(room *types.RoomInfo) error {
		return defaultComposeProcessor.compose(room, req)
	})
	if err != nil {
		log.Printf("reject compose from %s in room %s: %v", req.Player, req.RoomID, err)
		notifyClientError(req.ClientID, req.Seq, pb.ErrorCode_ERROR_CODE_INVALID_CARDS, config.Msg(config.MsgCardsInvalid))
	}
}

// CardComposeProcessor 合成结算。调用方必须持有房间操作锁（见 room_sync.go）。
type CardComposeProcessor struct {
	cardPoolManager *cards.CardPoolManager
}

var defaultComposeProcessor = NewCardComposeProcessor()

// NewCardComposeProcessor 创建新的卡牌合成处理器
func NewCardComposeProcessor() *CardComposeProcessor {
	return &CardComposeProcessor{cardPoolManager: cards.GetCardPoolManager()}
}

func (ccp *CardComposeProcessor) compose(room *types.RoomInfo, req ComposeRequest) error {
	groups, err := ccp.validate(room, req)
	if err != nil {
		return err
	}

	removed, added := ccp.drawTargets(room, groups)
	if len(added) == 0 {
		return errNothingComposed
	}

	if err := room.RemoveCardsFromPlayerByUID(req.Player, uidsOf(removed)); err != nil {
		room.ReturnCardsToPool(added)
		return err
	}
	for _, c := range added {
		// 每 group_size 张换 1 张，手牌只会变少，不会超出上限
		if err := room.AddCardToPlayer(req.Player, c); err != nil {
			return fmt.Errorf("add composed card %s: %w", c.Name, err)
		}
	}

	broadcastRoomState(room, pb.GameStateReason_GAME_STATE_REASON_CARDS_COMPOSED)
	return nil
}

// validate 校验房间、卡牌归属与数量，返回按名称分组的服务器手牌
func (ccp *CardComposeProcessor) validate(room *types.RoomInfo, req ComposeRequest) (map[string][]models.Card, error) {
	if status := room.GetStatus(); status != types.RoomStatusPlaying {
		return nil, fmt.Errorf("room is not playing: %s", status)
	}
	player, err := room.GetPlayerInfo(req.Player)
	if err != nil {
		return nil, err
	}
	need := config.Game().Rules.Compose.GroupSize
	if len(req.CardUIDs) == 0 || len(req.CardUIDs)%need != 0 {
		return nil, fmt.Errorf("invalid card count for composition: %d (must be multiple of %d)", len(req.CardUIDs), need)
	}

	hand := make(map[string]models.Card, len(player.HandCards))
	for _, c := range player.HandCards {
		hand[c.UID] = c
	}
	seen := make(map[string]bool, len(req.CardUIDs))
	groups := make(map[string][]models.Card)
	for _, uid := range req.CardUIDs {
		if seen[uid] {
			return nil, fmt.Errorf("duplicate card UID %s in compose request", uid)
		}
		seen[uid] = true
		c, ok := hand[uid]
		if !ok {
			return nil, fmt.Errorf("card UID %s not found in player %s's hand", uid, req.Player)
		}
		groups[c.Name] = append(groups[c.Name], c)
	}
	return groups, nil
}

// drawTargets 为每组凑满的同名卡从共享卡池抽取目标卡。
// 只有成功抽到目标卡的那一组才计入 removed，抽不到时这组牌原样留在手里。
func (ccp *CardComposeProcessor) drawTargets(room *types.RoomInfo, groups map[string][]models.Card) (removed, added []models.Card) {
	need := config.Game().Rules.Compose.GroupSize
	for _, group := range groups {
		template := group[0]
		if template.TargetName == nil || *template.TargetName == "" {
			continue // 不可合成的卡
		}
		target, err := ccp.cardPoolManager.GetCardByName(*template.TargetName)
		if err != nil {
			continue
		}
		for i := 0; i+need <= len(group); i += need {
			drawn, err := room.DrawCardByNameFromPool(*template.TargetName, target.Level)
			if err != nil {
				break // 目标卡池已空
			}
			removed = append(removed, group[i:i+need]...)
			added = append(added, *drawn)
		}
	}
	return removed, added
}

func uidsOf(cards []models.Card) []string {
	uids := make([]string, 0, len(cards))
	for _, c := range cards {
		uids = append(uids, c.UID)
	}
	return uids
}
