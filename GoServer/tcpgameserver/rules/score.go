package rules

import "math"

type StepKind int

const (
	StepPoints    StepKind = iota // 加点
	StepAddMult                   // 加倍率
	StepTimesMult                 // 乘倍率
)

// Step is one visible change during scoring, in order. The client replays these.
type Step struct {
	Phase   int    // 1..5，对应文档「计分」的步骤
	Source  string // 卡名 / 羁绊名 / 法宝名
	CardUID string // 逐卡步骤的卡；否则为空
	Relic   int    // 法宝栏位置；否则 -1
	Kind    StepKind
	Value   float64 // 本步的增量或乘数
	Points  float64 // 本步之后的点数
	Mult    float64 // 本步之后的倍率
}

// ScoreResult is the full scoring of a board.
type ScoreResult struct {
	Points float64
	Mult   float64
	Score  float64
	Steps  []Step
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }

type scorer struct {
	r ScoreResult
}

func (s *scorer) add(phase int, src, uid string, relic int, kind StepKind, v float64) {
	if v == 0 && kind != StepTimesMult {
		return
	}
	switch kind {
	case StepPoints:
		s.r.Points += v
	case StepAddMult:
		s.r.Mult += v
	case StepTimesMult:
		if v == 1 {
			return
		}
		s.r.Mult = round2(s.r.Mult * v)
	}
	s.r.Steps = append(s.r.Steps, Step{Phase: phase, Source: src, CardUID: uid, Relic: relic, Kind: kind, Value: v, Points: s.r.Points, Mult: s.r.Mult})
}

// Score scores the player's current board/bench/relics.
func (m *Match) Score(p *Player) ScoreResult { return m.ScoreWith(p, p.Board, p.Bench, p.Relics) }

// ScoreWith scores a hypothetical layout (used by previews and bots).
func (m *Match) ScoreWith(p *Player, board, bench []Card, relics []RelicInst) ScoreResult {
	s := &scorer{r: ScoreResult{Mult: 1}}
	if len(board) == 0 {
		s.r.Mult = 1
		return s.r
	}
	cfg := m.cfg
	eff := m.effective(relics)
	bc := m.bondCounts(board, relics)

	// extra triggers on the leftmost relic slots
	times := make([]int, len(eff))
	for i := range times {
		times[i] = 1
	}
	extra := int(m.modifier(bc, "retrigger_left"))
	if v, ok := m.legendSum(board, "retrigger_left", "count"); ok {
		extra += int(v)
	}
	if len(times) > 0 {
		times[0] += extra
	}
	reps := func(r effRelic) int {
		if m.behaviour(r.Relic).times {
			return 1
		}
		return times[r.slot]
	}
	relicName := func(r effRelic) string { return cfg.Relics[r.Relic].Name }
	param := func(r effRelic, k string) float64 { return cfg.Relics[r.Relic].Params[k] }

	// 1. per card, left to right
	last := len(board) - 1
	scoreCard := func(idx int, c Card) {
		own := float64(m.Points(c))
		s.add(1, m.CardName(c), c.UID, -1, StepPoints, own)
		for _, r := range eff {
			if r.empty || m.behaviour(r.Relic).kind != kindPerCard {
				continue
			}
			for t := 0; t < reps(r); t++ {
				switch m.relicKey(r) {
				case "heirloom":
					if c.Star == 2 {
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case "new_wave":
					if c.Star == 1 {
						s.add(1, relicName(r), c.UID, r.slot, StepPoints, param(r, "points"))
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case "rearguard":
					if idx == last {
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case "vanguard":
					if idx == 0 {
						s.add(1, relicName(r), c.UID, r.slot, StepPoints, (param(r, "factor")-1)*own)
					}
				}
			}
		}
	}
	for idx, c := range board {
		scoreCard(idx, c)
		for _, r := range eff {
			if !r.empty && m.relicKey(r) == "secret_art" && r.Choice == cfg.TradeOf(c.Line) {
				for t := 0; t < reps(r)*int(param(r, "repeats")); t++ {
					scoreCard(idx, c)
				}
			}
		}
	}

	// 2. bench relics
	for _, r := range eff {
		if r.empty || m.behaviour(r.Relic).kind != kindBench {
			continue
		}
		for t := 0; t < reps(r); t++ {
			switch m.relicKey(r) {
			case "hidden":
				n := 0
				for _, c := range bench {
					if c.Star >= 2 {
						n++
					}
				}
				s.add(2, relicName(r), "", r.slot, StepAddMult, param(r, "mult")*float64(n))
			case "storehouse":
				s.add(2, relicName(r), "", r.slot, StepPoints, param(r, "points")*float64(len(bench)))
			}
		}
	}

	// 3. bonds: points → generic tier bonus + add_mult (+ sect token) → times_mult
	for i, st := range bc.Origin {
		if st.Tier >= 0 {
			s.add(3, cfg.Origins[i].Name, "", -1, StepPoints, cfg.Origins[i].Tiers[st.Tier].Values["points_per_card"]*float64(len(board)))
		}
	}
	for i, st := range bc.Trade {
		if st.Tier >= 0 {
			s.add(3, cfg.Trades[i].Name, "", -1, StepPoints, cfg.Trades[i].Tiers[st.Tier].Values["points_per_card"]*float64(len(board)))
		}
	}
	s.add(3, "羁绊", "", -1, StepAddMult, cfg.Rules.Battle.BondTierMult*float64(bc.TierCount()))
	for i, st := range bc.Trade {
		if st.Tier >= 0 {
			s.add(3, cfg.Trades[i].Name, "", -1, StepAddMult, cfg.Trades[i].Tiers[st.Tier].Values["add_mult"])
		}
	}
	for i, st := range bc.Origin {
		if st.Tier >= 0 {
			s.add(3, cfg.Origins[i].Name, "", -1, StepAddMult, cfg.Origins[i].Tiers[st.Tier].Values["add_mult"])
		}
	}
	for _, r := range eff {
		if !r.empty && m.relicKey(r) == "sect_token" && r.Choice >= 0 && bc.Trade[r.Choice].Tier >= 0 {
			s.add(3, relicName(r), "", r.slot, StepAddMult, param(r, "mult"))
		}
	}
	for _, tm := range m.timesModifiers(bc) {
		s.add(3, tm.name, "", -1, StepTimesMult, tm.factor)
	}

	// 4. 3★ skills, left to right
	distinct := map[int]bool{}
	nMax := 0
	for _, c := range board {
		distinct[c.Line] = true
		if c.Star == cfg.Rules.Merge.MaxStar {
			nMax++
		}
	}
	for _, ls := range m.legendsOnWithLine(board) {
		name := cfg.Lines[ls.line].Card
		pp := ls.skill.Params
		switch ls.skill.Key {
		case "mult_per_high_star":
			n := 0
			for _, x := range board {
				if x.Star >= 2 {
					n++
				}
			}
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(n))
		case "mult_per_relic":
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(len(relics)))
		case "mult_per_level":
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(p.Level))
		case "mult_per_bond":
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(bc.ActiveBonds()))
		case "mult_per_line":
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(len(distinct)))
		case "times_mult":
			s.add(4, name, "", -1, StepTimesMult, pp["factor"])
		}
	}

	// 5. whole relics, left to right
	empty := m.Slots(p) - len(board)
	if empty < 0 {
		empty = 0
	}
	for _, r := range eff {
		if r.empty || m.behaviour(r.Relic).kind != kindWhole {
			continue
		}
		for t := 0; t < reps(r); t++ {
			name := relicName(r)
			switch m.relicKey(r) {
			case "furnace", "accountant", "grudge":
				s.add(5, name, "", r.slot, StepAddMult, r.Value)
			case "master_seal":
				for k := 0; k < nMax; k++ {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case "fellowship":
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(bc.ActiveBonds()))
			case "eclectic":
				if len(distinct) >= int(param(r, "min_lines")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case "lone_wolf":
				if len(board) <= int(param(r, "max_cards")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case "lean":
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(empty))
			case "multitude":
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(len(board)))
			case "piggy":
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(p.Gold/int(param(r, "per_gold"))))
			case "momentum":
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(p.WinStreak))
			case "last_stand":
				if float64(p.HP) < param(r, "hp_below") {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case "tailwind":
				s.add(5, name, "", r.slot, StepTimesMult, 1+param(r, "per_level")*float64(p.Level))
			case "literary_pen":
				if bc.Trade[cfg.tradeIdx["letter"]].Tier >= 0 {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			}
		}
	}

	s.r.Score = math.Floor(s.r.Points * s.r.Mult)
	if s.r.Score > cfg.Rules.Battle.ScoreCap {
		s.r.Score = cfg.Rules.Battle.ScoreCap
	}
	if s.r.Score < 0 {
		s.r.Score = 0
	}
	return s.r
}
