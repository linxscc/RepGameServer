package rules

import "math"

type StepKind int

const (
	StepPoints    StepKind = iota // 加点
	StepAddMult                   // 加倍率
	StepTimesMult                 // 乘倍率
)

// StepOrigin says whose effect produced a step.
type StepOrigin int

const (
	OriginOwn      StepOrigin = 0 // 自己的卡 / 羁绊 / 法宝
	OriginFoeRelic StepOrigin = 1 // 对手的法宝（Relic = 对手法宝栏位置）
	OriginOmen     StepOrigin = 2 // 天象（Relic = -1，Source = 天象名）
)

// Step is one visible change during scoring, in order. The client replays these.
type Step struct {
	Phase   int    // 1..5，对应文档「计分」的步骤
	Source  string // 卡名 / 羁绊名 / 法宝名
	CardUID string // 逐卡步骤的卡；否则为空
	Relic   int    // 法宝栏位置；否则 -1（Origin 为对手法宝时是对手的法宝栏位置）
	Origin  StepOrigin
	Kind    StepKind
	Value   float64 // 本步的增量或乘数
	Points  float64 // 本步之后的点数
	Mult    float64 // 本步之后的倍率
	// 触发次数：本步是第 Trigger 次触发，共 Triggers 次（不重复触发时都为 1）。
	// 法宝自身重复触发时记法宝的次数，否则记所在卡的次数。
	Trigger  int
	Triggers int
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
	r                     ScoreResult
	cardTrig, cardTrigs   int // 当前逐卡结算是这张卡的第几次触发 / 共几次
	relicTrig, relicTrigs int // 当前法宝是第几次触发 / 共几次
	origin                StepOrigin
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
	trig, trigs := 1, 1
	if s.relicTrigs > 1 {
		trig, trigs = s.relicTrig, s.relicTrigs
	} else if s.cardTrigs > 1 {
		trig, trigs = s.cardTrig, s.cardTrigs
	}
	s.r.Steps = append(s.r.Steps, Step{Phase: phase, Source: src, CardUID: uid, Relic: relic, Origin: s.origin, Kind: kind, Value: v, Points: s.r.Points, Mult: s.r.Mult,
		Trigger: trig, Triggers: trigs})
}

// relicLoop runs one relic's effect n times, tagging the steps with the trigger count.
func (s *scorer) relicLoop(n int, f func()) {
	for t := 0; t < n; t++ {
		s.relicTrig, s.relicTrigs = t+1, n
		f()
	}
	s.relicTrig, s.relicTrigs = 0, 0
}

// Foe is the opponent side a score is computed against (counter relics).
type Foe struct {
	Board  []Card
	Relics []RelicInst
}

// Score scores the player's current board/bench/relics at settlement, against the
// opponent's current (locked) board and relics.
func (m *Match) Score(p *Player) ScoreResult {
	o := m.Opponent(p)
	return m.ScoreVs(p, p.Board, p.Bench, p.Relics, Foe{o.Board, o.Relics}, nil)
}

// Preview scores the player's current layout for the prep-phase view. The opponent is
// seen as their board at the last settlement plus their current relics, so the
// board they are arranging right now never leaks.
func (m *Match) Preview(p *Player) ScoreResult { return m.ScoreWith(p, p.Board, p.Bench, p.Relics) }

// ScoreWith scores a hypothetical layout (used by previews and bots) against the
// opponent's last settled board and current relics.
func (m *Match) ScoreWith(p *Player, board, bench []Card, relics []RelicInst) ScoreResult {
	o := m.Opponent(p)
	return m.ScoreVs(p, board, bench, relics, Foe{o.LastBoard, o.Relics}, nil)
}

// ScoreVs scores a layout against the given foe. extraOmens are omen indexes to treat
// as active in addition to the match's (hypothetical scoring); they are not stored.
func (m *Match) ScoreVs(p *Player, board, bench []Card, relics []RelicInst, foe Foe, extraOmens []int) ScoreResult {
	s := &scorer{r: ScoreResult{Mult: 1}}
	if len(board) == 0 {
		s.r.Mult = 1
		return s.r
	}
	cfg := m.cfg
	eff := m.effective(relics)
	foeEff := m.effective(foe.Relics)
	// active omens in order of appearance, then extras; 失和 counts extras too
	omens := m.omenList(extraOmens)
	bc := m.bondCounts(board, relics, int(m.omenSum(OmenDiscord, "at", extraOmens)))

	// 天地异变: the suits with the most cards on this board (at least 2; ties all count)
	var boardSuits [4]int
	for _, c := range board {
		boardSuits[m.SuitOf(c)]++
	}
	topSuit := 2
	for _, n := range boardSuits {
		if n > topSuit {
			topSuit = n
		}
	}

	// extra triggers on the leftmost relic slots
	times := make([]int, len(eff))
	for i := range times {
		times[i] = 1
	}
	extra := 0
	if v, ok := m.legendSum(board, LegendRetriggerLeft, "count"); ok {
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

	// 1. per card, left to right. Each settlement (including 家传绝学 repeats) runs, in order:
	//   a. the card's own points
	//   b. own per-card factor relics (kindFactor: 神龛)
	//   c. foe per-card counters (kindFoeCard: 照妖镜、破神符), foe relic order, once
	//   d. active omen factors (and extraOmens), in order of appearance
	//   e. foe 斩首 when idx == 0
	//   f. other per-card relics (kindPerCard), in relic-bar order
	// cardPts is the card's current points; a factor step adds (f−1)×cardPts and then
	// scales cardPts by f, so later steps (先锋旗) see the factored value.
	last := len(board) - 1
	scoreCard := func(idx int, c Card) {
		cardPts := float64(m.Points(c))
		s.add(1, m.CardName(c), c.UID, -1, StepPoints, cardPts)
		factor := func(src string, relic int, f float64) {
			s.add(1, src, c.UID, relic, StepPoints, (f-1)*cardPts)
			cardPts *= f
		}

		// b. own per-card factor relics
		for _, r := range eff {
			if r.empty || m.behaviour(r.Relic).kind != kindFactor {
				continue
			}
			s.relicLoop(reps(r), func() {
				switch m.relicKey(r) {
				case RelicShrine:
					if m.IsGod(c) {
						factor(relicName(r), r.slot, param(r, "factor"))
					}
				}
			})
		}

		// c. foe per-card counters, in the foe's relic-bar order; once per settlement
		// (the foe's retriggers do not apply). Copies (铜镜/拓本) act as usual.
		s.origin = OriginFoeRelic
		for _, r := range foeEff {
			if r.empty || m.behaviour(r.Relic).kind != kindFoeCard {
				continue
			}
			switch m.relicKey(r) {
			case RelicDemonMirror:
				if !m.IsGod(c) {
					factor(relicName(r), r.slot, param(r, "factor"))
				}
			case RelicGodBreaker:
				if m.IsGod(c) {
					factor(relicName(r), r.slot, param(r, "factor"))
				}
			}
		}
		s.origin = OriginOwn

		// d. active omen factors, in order of appearance; once per settlement
		s.origin = OriginOmen
		for _, oi := range omens {
			o := cfg.Omens[oi]
			hit := false
			switch o.Key {
			case OmenSolarEclipse:
				hit = m.IsGod(c)
			case OmenGhostRout:
				hit = !m.IsGod(c)
			case OmenRedMoon:
				hit = c.Color == Red
			case OmenInkRain:
				hit = c.Color == Black
			case OmenUpheaval:
				hit = boardSuits[m.SuitOf(c)] >= topSuit
			}
			if hit {
				factor(o.Name, -1, o.Params["factor"])
			}
		}
		s.origin = OriginOwn

		// e. foe 斩首 on the leftmost card
		if idx == 0 {
			s.origin = OriginFoeRelic
			for _, r := range foeEff {
				if !r.empty && m.relicKey(r) == RelicBehead {
					s.add(1, relicName(r), c.UID, r.slot, StepPoints, -cardPts)
					cardPts = 0
				}
			}
			s.origin = OriginOwn
		}

		// f. other per-card relics
		for _, r := range eff {
			if r.empty || m.behaviour(r.Relic).kind != kindPerCard {
				continue
			}
			s.relicLoop(reps(r), func() {
				switch m.relicKey(r) {
				case RelicHeirloom:
					if c.Star == StarTwo {
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case RelicNewWave:
					if c.Star == StarOne {
						s.add(1, relicName(r), c.UID, r.slot, StepPoints, param(r, "points"))
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case RelicRearguard:
					if idx == last {
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case RelicVanguard:
					if idx == 0 {
						s.add(1, relicName(r), c.UID, r.slot, StepPoints, (param(r, "factor")-1)*cardPts)
					}
				case RelicCinnabar:
					if c.Color == Red {
						s.add(1, relicName(r), c.UID, r.slot, StepAddMult, param(r, "mult"))
					}
				case RelicPineInk:
					if c.Color == Black {
						s.add(1, relicName(r), c.UID, r.slot, StepPoints, param(r, "points"))
					}
				}
			})
		}
	}
	for idx, c := range board {
		n := 1 // 本卡触发次数：1 + 家传绝学等重复结算
		for _, r := range eff {
			if !r.empty && m.relicKey(r) == RelicSecretArt && r.Choice == cfg.TradeOf(c.Line) {
				n += reps(r) * int(param(r, "repeats"))
			}
		}
		for t := 0; t < n; t++ {
			s.cardTrig, s.cardTrigs = t+1, n
			scoreCard(idx, c)
		}
		s.cardTrig, s.cardTrigs = 0, 0
	}

	// 2. bench relics
	for _, r := range eff {
		if r.empty || m.behaviour(r.Relic).kind != kindBench {
			continue
		}
		s.relicLoop(reps(r), func() {
			switch m.relicKey(r) {
			case RelicHidden:
				n := 0
				for _, c := range bench {
					if c.Star >= StarTwo {
						n++
					}
				}
				s.add(2, relicName(r), "", r.slot, StepAddMult, param(r, "mult")*float64(n))
			case RelicStorehouse:
				s.add(2, relicName(r), "", r.slot, StepPoints, param(r, "points")*float64(len(bench)))
			}
		})
	}

	// 3. bonds: points → generic tier bonus + add_mult (+ sect token) → times_mult
	for i, st := range bc.Origin {
		if st.Tier >= 0 {
			s.add(3, cfg.Origins[i].Name, "", -1, StepPoints, cfg.Origins[i].Tiers[st.Tier].Values[BondPointsPerCard]*float64(len(board)))
		}
	}
	for i, st := range bc.Trade {
		if st.Tier >= 0 {
			s.add(3, cfg.Trades[i].Name, "", -1, StepPoints, cfg.Trades[i].Tiers[st.Tier].Values[BondPointsPerCard]*float64(len(board)))
		}
	}
	s.add(3, "羁绊", "", -1, StepAddMult, cfg.Rules.Battle.BondTierMult*float64(bc.TierCount()))
	for i, st := range bc.Trade {
		if st.Tier >= 0 {
			s.add(3, cfg.Trades[i].Name, "", -1, StepAddMult, cfg.Trades[i].Tiers[st.Tier].Values[BondAddMult])
		}
	}
	for i, st := range bc.Origin {
		if st.Tier >= 0 {
			s.add(3, cfg.Origins[i].Name, "", -1, StepAddMult, cfg.Origins[i].Tiers[st.Tier].Values[BondAddMult])
		}
	}
	for _, r := range eff {
		if !r.empty && m.relicKey(r) == RelicSectToken && r.Choice >= 0 && bc.Trade[r.Choice].Tier >= 0 {
			s.add(3, relicName(r), "", r.slot, StepAddMult, param(r, "mult"))
		}
	}
	for _, tm := range m.timesModifiers(bc) {
		s.add(3, tm.name, "", -1, StepTimesMult, tm.factor)
	}

	// 4. 3★ skills, left to right
	distinct := map[int]bool{}
	nMax := 0
	var suits [4]int // 按 Suit 计数
	ghosts, reds := 0, 0
	for _, c := range board {
		distinct[c.Line] = true
		if c.Star == cfg.Rules.Merge.MaxStar {
			nMax++
		}
		suits[m.SuitOf(c)]++
		if !m.IsGod(c) {
			ghosts++
		}
		if c.Color == Red {
			reds++
		}
	}
	blacks := len(board) - reds
	foeReds := 0
	for _, c := range foe.Board {
		if c.Color == Red {
			foeReds++
		}
	}
	foeBlacks := len(foe.Board) - foeReds
	foeSuitRelics := 0 // 对手法宝栏里的花色法宝（不含铜镜/拓本复制品）
	for _, r := range foe.Relics {
		if cfg.CategoryOf(r.Relic) == CatSuit {
			foeSuitRelics++
		}
	}
	for _, ls := range m.legendsOnWithLine(board) {
		name := cfg.Lines[ls.line].Card
		pp := ls.skill.Params
		switch ls.skill.Key {
		case LegendMultPerHighStar:
			n := 0
			for _, x := range board {
				if x.Star >= StarTwo {
					n++
				}
			}
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(n))
		case LegendMultPerRelic:
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(len(relics)))
		case LegendMultPerLevel:
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(p.Level))
		case LegendMultPerBond:
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(bc.ActiveBonds()))
		case LegendMultPerLine:
			s.add(4, name, "", -1, StepAddMult, pp["mult"]*float64(len(distinct)))
		case LegendTimesMult:
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
		s.relicLoop(reps(r), func() {
			name := relicName(r)
			switch m.relicKey(r) {
			case RelicFurnace, RelicAccountant, RelicGrudge:
				s.add(5, name, "", r.slot, StepAddMult, r.Value)
			case RelicMasterSeal:
				for k := 0; k < nMax; k++ {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicFellowship:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(bc.ActiveBonds()))
			case RelicEclectic:
				if len(distinct) >= int(param(r, "min_lines")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicLoneWolf:
				if len(board) <= int(param(r, "max_cards")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicLean:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(empty))
			case RelicMultitude:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(len(board)))
			case RelicPiggy:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(p.Gold/int(param(r, "per_gold"))))
			case RelicMomentum:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(p.WinStreak))
			case RelicLastStand:
				if float64(p.HP) < param(r, "hp_below") {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicTailwind:
				s.add(5, name, "", r.slot, StepTimesMult, 1+param(r, "per_level")*float64(p.Level))
			case RelicLiteraryPen:
				if bc.Trade[cfg.tradeIdx[TradeLetter]].Tier >= 0 {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicNightParade:
				if ghosts >= int(param(r, "min_cards")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicOneColor:
				if len(board) >= int(param(r, "min_cards")) && (reds == 0 || blacks == 0) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicYinYang:
				if reds == blacks && reds >= int(param(r, "min_each")) {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			case RelicFourSymbols:
				if suits[SuitRedGod] > 0 && suits[SuitBlackGod] > 0 && suits[SuitRedGhost] > 0 && suits[SuitBlackGhost] > 0 {
					s.add(5, name, "", r.slot, StepAddMult, param(r, "mult"))
				}
			case RelicColorCounter:
				switch {
				case foeReds > foeBlacks:
					s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(reds))
				case foeBlacks > foeReds:
					s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(blacks))
				}
			case RelicUsurp:
				s.add(5, name, "", r.slot, StepAddMult, param(r, "mult")*float64(foeSuitRelics))
			case RelicDefyHeaven:
				for range omens {
					s.add(5, name, "", r.slot, StepTimesMult, param(r, "factor"))
				}
			}
		})
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
