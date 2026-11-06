package logic

import (
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/rules"
)

// chess_views.go converts rules state into protocol messages. It only reads state;
// every number the client shows is computed here on the server.

func matchConfigMessage(roomID string, m *rules.Match) *pb.ServerMessage {
	cfg := m.Config()
	r := cfg.Rules
	msg := &pb.MatchConfig{
		RoomId:       roomID,
		MaxStar:      int32(r.Merge.MaxStar),
		BenchSize:    int32(r.Board.BenchSize),
		RefreshCost:  int32(r.Economy.RefreshCost),
		XpCost:       int32(r.Economy.XPCost),
		MaxLevel:     int32(cfg.MaxLevel()),
		StartHp:      int32(r.Player.StartHP),
		BondTierMult: r.Battle.BondTierMult,
	}
	for _, l := range cfg.Lines {
		msg.Lines = append(msg.Lines, &pb.LineInfo{
			Key: l.Key, Name: l.Name, StarNames: []string{l.Card}, Cost: int32(l.Cost), Points: int32(l.Points),
			TradeKey: l.Trade, OriginKey: l.Origin, LegendDesc: l.Legend.Desc,
		})
	}
	bondInfo := func(b rules.BondDef, isTrade bool) *pb.BondInfo {
		bi := &pb.BondInfo{Key: b.Key, Name: b.Name, Desc: b.Desc}
		for _, t := range b.Tiers {
			bi.Tiers = append(bi.Tiers, &pb.TierInfo{At: int32(t.At), Desc: t.Desc})
		}
		for _, l := range cfg.Lines {
			if (isTrade && l.Trade == b.Key) || (!isTrade && l.Origin == b.Key) {
				bi.LineKeys = append(bi.LineKeys, l.Key)
			}
		}
		return bi
	}
	for _, b := range cfg.Trades {
		msg.Trades = append(msg.Trades, bondInfo(b, true))
	}
	for _, b := range cfg.Origins {
		msg.Origins = append(msg.Origins, bondInfo(b, false))
	}
	for i, rl := range cfg.Relics {
		msg.Relics = append(msg.Relics, &pb.RelicInfo{
			Key: rl.Key, Name: rl.Name, Rarity: int32(cfg.RarityOf(i)), Desc: rl.Desc,
			NeedsTradeChoice: m.RelicNeedsChoice(i),
		})
	}
	for _, rr := range r.Rarities {
		msg.Rarities = append(msg.Rarities, &pb.RarityInfo{Key: rr.Key, Name: rr.Name, Price: int32(rr.Price)})
	}
	for _, f := range cfg.Families {
		msg.Families = append(msg.Families, &pb.FamilyInfo{Key: f.Key, Name: f.Name, Desc: f.Desc})
	}
	return &pb.ServerMessage{Body: &pb.ServerMessage_MatchConfig{MatchConfig: msg}}
}

func chessCards(m *rules.Match, cards []rules.Card) []*pb.ChessCard {
	out := make([]*pb.ChessCard, 0, len(cards))
	for _, c := range cards {
		out = append(out, &pb.ChessCard{
			Uid: c.UID, LineKey: m.Config().Lines[c.Line].Key, Star: int32(c.Star), Name: m.CardName(c),
			Points: int32(m.Points(c)), Bonus: int32(c.Bonus), SellPrice: int32(m.SellPrice(c)),
		})
	}
	return out
}

func relicSlots(m *rules.Match, relics []rules.RelicInst) []*pb.RelicSlot {
	out := make([]*pb.RelicSlot, 0, len(relics))
	for _, r := range relics {
		out = append(out, &pb.RelicSlot{
			RelicKey: m.Config().Relics[r.Relic].Key, Value: r.Value, TradeChoice: int32(r.Choice),
			SellPrice: int32(m.Config().RelicPrice(r.Relic) / 2),
		})
	}
	return out
}

func bondCounts(m *rules.Match, bc rules.BondCounts) (trades, origins []*pb.BondCount) {
	cfg := m.Config()
	for i, s := range bc.Trade {
		trades = append(trades, &pb.BondCount{Key: cfg.Trades[i].Key, Count: int32(s.Count), Tier: int32(s.Tier)})
	}
	for i, s := range bc.Origin {
		origins = append(origins, &pb.BondCount{Key: cfg.Origins[i].Key, Count: int32(s.Count), Tier: int32(s.Tier)})
	}
	return
}

func scoreBreakdown(r rules.ScoreResult) *pb.ScoreBreakdown {
	out := &pb.ScoreBreakdown{Points: r.Points, Mult: r.Mult, Score: r.Score}
	for _, s := range r.Steps {
		kind := pb.ScoreStepKind_SCORE_STEP_KIND_POINTS
		switch s.Kind {
		case rules.StepAddMult:
			kind = pb.ScoreStepKind_SCORE_STEP_KIND_ADD_MULT
		case rules.StepTimesMult:
			kind = pb.ScoreStepKind_SCORE_STEP_KIND_TIMES_MULT
		}
		out.Steps = append(out.Steps, &pb.ScoreStep{
			Phase: int32(s.Phase), Source: s.Source, CardUid: s.CardUID, RelicSlot: int32(s.Relic),
			Kind: kind, Value: s.Value, Points: s.Points, Mult: s.Mult,
		})
	}
	return out
}

func selfView(m *rules.Match, p *rules.Player) *pb.SelfView {
	cfg := m.Config()
	v := &pb.SelfView{
		Hp: int32(p.HP), Gold: int32(p.Gold), Level: int32(p.Level), Xp: int32(p.XP), XpToNext: int32(m.XPToNext(p)),
		Slots: int32(m.Slots(p)), RelicSlots: int32(m.RelicSlots(p)),
		Board: chessCards(m, p.Board), Bench: chessCards(m, p.Bench), Relics: relicSlots(m, p.Relics),
		ShopLocked: p.Locked, FreeRefreshes: int32(m.FreeRefreshesLeft(p)), Interest: int32(m.Interest(p)),
		Family: int32(p.Family), WinStreak: int32(p.WinStreak), LossStreak: int32(p.LossStreak), Ready: p.Ready,
		Preview: scoreBreakdown(m.Score(p)),
	}
	for _, l := range p.Shop {
		if l < 0 {
			v.Shop = append(v.Shop, &pb.ShopSlot{})
			continue
		}
		v.Shop = append(v.Shop, &pb.ShopSlot{LineKey: cfg.Lines[l].Key, Cost: int32(cfg.Lines[l].Cost), Owned: int32(m.CountCopies(p, l, 1))})
	}
	if p.ShopRelic >= 0 {
		v.ShopRelicKey = cfg.Relics[p.ShopRelic].Key
	}
	v.TradeBonds, v.OriginBonds = bondCounts(m, m.Bonds(p))
	return v
}

func opponentView(cr *chessRoom, seat int) *pb.OpponentView {
	m := cr.match
	o := m.P[1-seat]
	return &pb.OpponentView{
		Username: o.ID, Hp: int32(o.HP), Level: int32(o.Level), Gold: int32(o.Gold),
		CardCount: int32(len(o.Board) + len(o.Bench)), Relics: relicSlots(m, o.Relics), Family: int32(o.Family),
		LastBoard: chessCards(m, o.LastBoard), LastScore: cr.lastScore[1-seat], Ready: o.Ready,
		Connected: cr.isPresent(1 - seat),
	}
}

func matchStateMessage(cr *chessRoom, seat int, replyTo uint32) *pb.ServerMessage {
	m := cr.match
	p := m.P[seat]
	round := m.Round
	if cr.phase == pb.MatchPhase_MATCH_PHASE_SETTLE {
		round = cr.settled
	}
	st := &pb.MatchState{
		RoomId: cr.roomID, Phase: cr.phase, Round: int32(round), StageDamage: int32(m.Config().StageDamage(round)),
		Self: selfView(m, p), Opponent: opponentView(cr, seat),
	}
	if !cr.deadline.IsZero() {
		st.PhaseDeadlineMs = cr.deadline.UnixMilli()
	}
	if cr.phase == pb.MatchPhase_MATCH_PHASE_FAMILY {
		for _, f := range p.FamilyOffer {
			st.FamilyOffer = append(st.FamilyOffer, int32(f))
		}
	}
	if cr.phase == pb.MatchPhase_MATCH_PHASE_VAULT {
		for _, id := range p.VaultOffer {
			st.VaultOffer = append(st.VaultOffer, m.Config().Relics[id].Key)
		}
	}
	return &pb.ServerMessage{ReplyTo: replyTo, Body: &pb.ServerMessage_MatchState{MatchState: st}}
}

func sideResult(cr *chessRoom, res rules.RoundResult, seat int) *pb.SideResult {
	m := cr.match
	p := m.P[seat]
	return &pb.SideResult{
		Username: p.ID, Board: chessCards(m, p.LastBoard), Relics: relicSlots(m, p.Relics),
		Score: scoreBreakdown(res.Scores[seat]), DamageTaken: int32(res.Damage[seat]), HpAfter: int32(p.HP),
	}
}

func roundResultMessage(cr *chessRoom, res rules.RoundResult, seat int) *pb.ServerMessage {
	winner := int32(0)
	switch res.Winner {
	case seat:
		winner = 1
	case 1 - seat:
		winner = -1
	}
	rr := &pb.RoundResult{
		RoomId: cr.roomID, Round: int32(res.Round), Self: sideResult(cr, res, seat), Opponent: sideResult(cr, res, 1-seat),
		Winner: winner, MatchOver: res.Over, VaultNext: res.Vault, NextPhaseAtMs: cr.deadline.UnixMilli(),
	}
	return &pb.ServerMessage{Body: &pb.ServerMessage_RoundResult{RoundResult: rr}}
}

func matchOverMessage(cr *chessRoom, seat int) *pb.ServerMessage {
	m := cr.match
	result := int32(0)
	switch m.Winner {
	case seat:
		result = 1
	case 1 - seat:
		result = -1
	}
	return &pb.ServerMessage{Body: &pb.ServerMessage_MatchOver{MatchOver: &pb.MatchOver{
		RoomId: cr.roomID, Result: result, Rounds: int32(m.Round), Hp: int32(m.P[seat].HP), OpponentHp: int32(m.P[1-seat].HP),
		Surrendered: m.Surrendered,
	}}}
}
