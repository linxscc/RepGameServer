package chessbot

import "GoServer/tcpgameserver/logic/rules"

// SuitArchetypes are the suit-school playstyles (plus a suit-blind baseline).
func SuitArchetypes() []Archetype {
	normal := []int{0, 0, 0, 0, 9, 14, 20}
	fam := []string{rules.FamilyMerchant, rules.FamilyScholar, rules.FamilyNoble}
	mk := func(a Archetype) Archetype {
		a.LevelPlan, a.Families, a.Trade = normal, fam, rules.TradeLetter
		if a.SaveTo == 0 {
			a.SaveTo = 30
		}
		return a
	}
	return []Archetype{
		mk(Archetype{Name: "基线", IgnoreSuits: true}),
		mk(Archetype{Name: "纯红", Color: rules.ColorNameRed,
			Relics: map[string]float64{rules.RelicCinnabar: 5, rules.RelicVermilion: 4, rules.RelicOneColor: 5, rules.RelicFourSymbols: 2, rules.RelicTailwind: 3, rules.RelicMirror: 3}}),
		mk(Archetype{Name: "纯黑", Color: rules.ColorNameBlack,
			Relics: map[string]float64{rules.RelicPineInk: 5, rules.RelicOneColor: 5, rules.RelicFourSymbols: 2, rules.RelicTailwind: 3, rules.RelicMirror: 3, rules.RelicRubbing: 2}}),
		mk(Archetype{Name: "神", Kind: rules.KindGod, Lines: []string{rules.LineSmith, rules.LineTaoist, rules.LineMonk, rules.LineDoctor, rules.LineTrader, rules.LineClerk},
			Relics: map[string]float64{rules.RelicShrine: 5, rules.RelicYinYang: 3, rules.RelicFourSymbols: 3, rules.RelicTailwind: 3, rules.RelicMirror: 3}}),
		mk(Archetype{Name: "鬼", Kind: rules.KindGhost, Lines: []string{rules.LineCarpenter, rules.LineFarmer, rules.LineScholar, rules.LineArtist, rules.LineDesigner, rules.LineSoldier},
			Relics: map[string]float64{rules.RelicNightParade: 5, rules.RelicYinYang: 3, rules.RelicFourSymbols: 3, rules.RelicTailwind: 3, rules.RelicMirror: 3}}),
		mk(Archetype{Name: "特攻",
			Relics: map[string]float64{rules.RelicDemonMirror: 5, rules.RelicGodBreaker: 5, rules.RelicBehead: 4, rules.RelicColorCounter: 4, rules.RelicUsurp: 4, rules.RelicTailwind: 3}}),
		mk(Archetype{Name: "经济", SaveTo: 20,
			Relics: map[string]float64{rules.RelicTreasure: 5, rules.RelicPiggy: 4, rules.RelicThrifty: 4, rules.RelicGraduation: 4, rules.RelicMarketPass: 4, rules.RelicAcademy: 4, rules.RelicExpansion: 5}}),
	}
}

// holds reports whether p owns the relic with the given key.
func holds(m *rules.Match, p *rules.Player, key string) bool {
	for _, r := range p.Relics {
		if m.Config().Relics[r.Relic].Key == key {
			return true
		}
	}
	return false
}

// ownCounts tallies the fielded + benched cards.
func ownCounts(m *rules.Match, p *rules.Player) (total, red, black, god, ghost int, suits [4]bool) {
	for _, cs := range [][]rules.Card{p.Board, p.Bench} {
		for _, c := range cs {
			total++
			if c.Color == rules.Red {
				red++
			} else {
				black++
			}
			if m.IsGod(c) {
				god++
			} else {
				ghost++
			}
			suits[m.SuitOf(c)] = true
		}
	}
	return
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// synergy scores 0..1 how well a relic fits what the bot holds (or what the opponent fielded).
// With nothing to judge by (no cards) it returns 0.5.
func synergy(m *rules.Match, p *rules.Player, id int) float64 {
	key := m.Config().Relics[id].Key
	total, red, black, god, ghost, suits := ownCounts(m, p)
	frac := func(n, d int) float64 {
		if d == 0 {
			return 0.5
		}
		return float64(n) / float64(d)
	}
	opp := m.Opponent(p)
	oppBoard := opp.LastBoard
	oppFrac := func(pred func(rules.Card) bool) float64 {
		if len(oppBoard) == 0 {
			return 0.5
		}
		n := 0
		for _, c := range oppBoard {
			if pred(c) {
				n++
			}
		}
		return float64(n) / float64(len(oppBoard))
	}
	switch key {
	case rules.RelicCinnabar, rules.RelicVermilion:
		return frac(red, total)
	case rules.RelicPineInk:
		return frac(black, total)
	case rules.RelicShrine:
		return frac(god, total)
	case rules.RelicNightParade:
		return frac(ghost, total)
	case rules.RelicOneColor:
		if red > black {
			return frac(red, total)
		}
		return frac(black, total)
	case rules.RelicYinYang:
		if total == 0 {
			return 0.5
		}
		return 1 - float64(absInt(red-black))/float64(total)
	case rules.RelicFourSymbols:
		if total == 0 {
			return 0.5
		}
		n := 0
		for _, s := range suits {
			if s {
				n++
			}
		}
		return float64(n) / 4
	case rules.RelicDemonMirror:
		return oppFrac(func(c rules.Card) bool { return !m.IsGod(c) })
	case rules.RelicGodBreaker:
		return oppFrac(m.IsGod)
	case rules.RelicColorCounter:
		if len(oppBoard) == 0 {
			return 0.5
		}
		r := 0
		for _, c := range oppBoard {
			if c.Color == rules.Red {
				r++
			}
		}
		return float64(absInt(r-(len(oppBoard)-r))) / float64(len(oppBoard))
	case rules.RelicBehead:
		if len(oppBoard) == 0 {
			return 0.5
		}
		sum := 0
		for _, c := range oppBoard {
			sum += m.Points(c)
		}
		if sum <= 0 {
			return 0
		}
		return float64(m.Points(oppBoard[0])) / float64(sum)
	case rules.RelicUsurp:
		n := 0
		for _, r := range opp.Relics {
			if m.Config().CategoryOf(r.Relic) == rules.CatSuit {
				n++
			}
		}
		if n >= 2 {
			return 1
		}
		return float64(n) / 2
	}
	return 0
}

// prefColor is the preferred color; ok is false when the bot has no color preference.
func (b *Bot) prefColor(m *rules.Match, p *rules.Player) (rules.Color, bool) {
	switch b.Arch.Color {
	case rules.ColorNameRed:
		return rules.Red, true
	case rules.ColorNameBlack:
		return rules.Black, true
	}
	if b.Arch.IgnoreSuits {
		return 0, false
	}
	if holds(m, p, rules.RelicCinnabar) || holds(m, p, rules.RelicVermilion) {
		return rules.Red, true
	}
	if holds(m, p, rules.RelicPineInk) {
		return rules.Black, true
	}
	if holds(m, p, rules.RelicOneColor) {
		red, black := 0, 0
		for _, c := range p.Board {
			if c.Color == rules.Red {
				red++
			} else {
				black++
			}
		}
		if red == 0 && black == 0 {
			_, red, black, _, _, _ = ownCounts(m, p)
		}
		if black > red {
			return rules.Black, true
		}
		return rules.Red, true
	}
	return 0, false
}

// prefKind is the preferred line kind ("god" / "ghost"); ok is false when there is none.
func (b *Bot) prefKind(m *rules.Match, p *rules.Player) (string, bool) {
	if b.Arch.Kind != "" {
		return b.Arch.Kind, true
	}
	if b.Arch.IgnoreSuits {
		return "", false
	}
	if holds(m, p, rules.RelicShrine) {
		return rules.KindGod, true
	}
	if holds(m, p, rules.RelicNightParade) {
		return rules.KindGhost, true
	}
	return "", false
}

// omenFlatLoss is the rough expected loss of omens that do not change the score.
var omenFlatLoss = map[string]float64{
	rules.OmenQuake: 0.04, rules.OmenFamine: 0.10, rules.OmenMarketClosed: 0.05, rules.OmenHeavyTax: 0.03,
	rules.OmenShortage: 0.05, rules.OmenNarrowPath: 0.12,
}

var omenScoring = map[string]bool{
	rules.OmenSolarEclipse: true, rules.OmenGhostRout: true, rules.OmenRedMoon: true,
	rules.OmenInkRain: true, rules.OmenUpheaval: true, rules.OmenDiscord: true,
}

// omenLosses returns the expected loss of p's layout if the next omen's point follows
// 1d6 (one) or 2d6 (two); shifting past active omens is ignored.
func omenLosses(m *rules.Match, p *rules.Player, board []rules.Card) (one, two float64) {
	cfg := m.Config()
	opp := m.Opponent(p)
	foe := rules.Foe{Board: opp.LastBoard, Relics: opp.Relics}
	s := m.ScoreVs(p, board, p.Bench, p.Relics, foe, nil).Score
	sides := cfg.Rules.Omen.DieSides
	loss := func(point int) float64 {
		oi := cfg.OmenByPoint(point)
		if oi < 0 {
			return 0
		}
		key := cfg.Omens[oi].Key
		if omenScoring[key] {
			if s == 0 {
				return 0
			}
			s2 := m.ScoreVs(p, board, p.Bench, p.Relics, foe, []int{oi}).Score
			return 1 - s2/s
		}
		l := omenFlatLoss[key]
		if key == rules.OmenHeavyTax {
			l += 0.02 * float64(m.Interest(p))
		}
		return l
	}
	for d := 1; d <= sides; d++ {
		one += loss(d) / float64(sides)
		for e := 1; e <= sides; e++ {
			two += loss(d+e) / float64(sides*sides)
		}
	}
	return
}

// omenPick reports whether to roll. The opponent is predicted to roll when 2d6 hurts them
// less than 1d6. If they roll, I get 2d6 by rolling and 1d6 by skipping; if they skip, I get
// 1d6 by rolling and the system's 2d6 by skipping. Ties skip.
func omenPick(myOne, myTwo, oppOne, oppTwo float64) bool {
	if oppTwo < oppOne {
		return myTwo < myOne
	}
	return myOne < myTwo
}

// ChooseOmen decides roll or skip for this bot's seat by expected loss.
func (b *Bot) ChooseOmen(m *rules.Match, p *rules.Player) {
	myOne, myTwo := omenLosses(m, p, p.Board)
	opp := m.Opponent(p)
	oppOne, oppTwo := omenLosses(m, opp, opp.LastBoard)
	_ = m.ChooseOmen(p, omenPick(myOne, myTwo, oppOne, oppTwo))
}
