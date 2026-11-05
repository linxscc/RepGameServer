package protocol

import (
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/pb"
)

// 领域模型 -> 协议消息 的转换。领域模型不感知协议，转换只在这里发生。

func ToCard(c models.Card) *pb.Card {
	return &pb.Card{
		Uid:        c.UID,
		Id:         int32(c.ID),
		Name:       c.Name,
		Damage:     c.Damage,
		TargetName: c.TargetName,
		Level:      int32(c.Level),
	}
}

func ToBond(b *models.BondModel) *pb.Bond {
	if b == nil {
		return nil
	}
	return &pb.Bond{
		Id:          int32(b.ID),
		Name:        b.Name,
		Level:       int32(b.Level),
		CardNames:   append([]string(nil), b.CardNames...),
		Damage:      b.Damage,
		Description: b.Description,
		Skill:       b.Skill,
	}
}

func ToBondList(bonds []*models.BondModel) *pb.BondList {
	out := &pb.BondList{Bonds: make([]*pb.Bond, 0, len(bonds))}
	for _, b := range bonds {
		if pbBond := ToBond(b); pbBond != nil {
			out.Bonds = append(out.Bonds, pbBond)
		}
	}
	return out
}

func ToGameState(info *models.PlayerGameInfo, reason pb.GameStateReason) *pb.GameState {
	if info == nil {
		return &pb.GameState{Reason: reason}
	}

	state := &pb.GameState{
		Reason:   reason,
		RoomId:   info.RoomId,
		Username: info.Username,
		Round:    info.Round,
		Health:   info.Health,
	}
	for _, c := range info.SelfCards {
		state.SelfCards = append(state.SelfCards, ToCard(c))
	}
	for _, o := range info.OtherPlayers {
		state.OtherPlayers = append(state.OtherPlayers, &pb.OtherPlayer{
			Username:   o.Username,
			Round:      o.Round,
			Health:     o.Health,
			CardsCount: int32(o.CardsCount),
		})
	}
	for _, d := range info.DamageInfo {
		dmg := &pb.DamageInfo{
			Source: d.DamageSource,
			Target: d.DamageTarget,
			Type:   d.DamageType,
			Value:  d.DamageValue,
		}
		for i := range d.TriggeredBonds {
			dmg.TriggeredBonds = append(dmg.TriggeredBonds, ToBond(&d.TriggeredBonds[i]))
		}
		state.DamageInfo = append(state.DamageInfo, dmg)
	}
	return state
}

// GameStateMessage 构造一条推送给玩家的对局状态消息。
func GameStateMessage(info *models.PlayerGameInfo, reason pb.GameStateReason) *pb.ServerMessage {
	return &pb.ServerMessage{Body: &pb.ServerMessage_GameState{GameState: ToGameState(info, reason)}}
}

// CardsFromUIDs 把客户端提交的 UID 列表转成仅含 UID 的卡牌，
// 由业务层在服务器手牌中查找真实数据。
func CardsFromUIDs(uids []string) []models.Card {
	cards := make([]models.Card, 0, len(uids))
	for _, uid := range uids {
		cards = append(cards, models.Card{UID: uid})
	}
	return cards
}
