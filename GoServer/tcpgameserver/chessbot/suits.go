package chessbot

import "GoServer/tcpgameserver/rules"

// SuitArchetypes are the suit-school playstyles (plus a suit-blind baseline).
func SuitArchetypes() []Archetype {
	normal := []int{0, 0, 0, 0, 9, 14, 20}
	fam := []string{"merchant", "scholar", "noble"}
	mk := func(a Archetype) Archetype {
		a.LevelPlan, a.Families, a.Trade = normal, fam, "letter"
		if a.SaveTo == 0 {
			a.SaveTo = 30
		}
		return a
	}
	return []Archetype{
		mk(Archetype{Name: "基线", IgnoreSuits: true}),
		mk(Archetype{Name: "纯红", Color: "red",
			Relics: map[string]float64{"cinnabar": 5, "vermilion": 4, "one_color": 5, "four_symbols": 2, "tailwind": 3, "mirror": 3}}),
		mk(Archetype{Name: "纯黑", Color: "black",
			Relics: map[string]float64{"pine_ink": 5, "one_color": 5, "four_symbols": 2, "tailwind": 3, "mirror": 3, "rubbing": 2}}),
		mk(Archetype{Name: "神", Kind: "god", Lines: []string{"smith", "taoist", "monk", "doctor", "trader", "clerk"},
			Relics: map[string]float64{"shrine": 5, "yin_yang": 3, "four_symbols": 3, "tailwind": 3, "mirror": 3}}),
		mk(Archetype{Name: "鬼", Kind: "ghost", Lines: []string{"carpenter", "farmer", "scholar", "artist", "designer", "soldier"},
			Relics: map[string]float64{"night_parade": 5, "yin_yang": 3, "four_symbols": 3, "tailwind": 3, "mirror": 3}}),
		mk(Archetype{Name: "特攻",
			Relics: map[string]float64{"demon_mirror": 5, "god_breaker": 5, "behead": 4, "color_counter": 4, "usurp": 4, "tailwind": 3}}),
		mk(Archetype{Name: "经济", SaveTo: 20,
			Relics: map[string]float64{"treasure": 5, "piggy": 4, "thrifty": 4, "graduation": 4, "market_pass": 4, "academy": 4, "expansion": 5}}),
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
	case "cinnabar", "vermilion":
		return frac(red, total)
	case "pine_ink":
		return frac(black, total)
	case "shrine":
		return frac(god, total)
	case "night_parade":
		return frac(ghost, total)
	case "one_color":
		if red > black {
			return frac(red, total)
		}
		return frac(black, total)
	case "yin_yang":
		if total == 0 {
			return 0.5
		}
		return 1 - float64(absInt(red-black))/float64(total)
	case "four_symbols":
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
	case "demon_mirror":
		return oppFrac(func(c rules.Card) bool { return !m.IsGod(c) })
	case "god_breaker":
		return oppFrac(m.IsGod)
	case "color_counter":
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
	case "behead":
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
	case "usurp":
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
	case "red":
		return rules.Red, true
	case "black":
		return rules.Black, true
	}
	if b.Arch.IgnoreSuits {
		return 0, false
	}
	if holds(m, p, "cinnabar") || holds(m, p, "vermilion") {
		return rules.Red, true
	}
	if holds(m, p, "pine_ink") {
		return rules.Black, true
	}
	if holds(m, p, "one_color") {
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
	if holds(m, p, "shrine") {
		return rules.KindGod, true
	}
	if holds(m, p, "night_parade") {
		return rules.KindGhost, true
	}
	return "", false
}

// omenFlatLoss is the rough expected loss of omens that do not change the score.
var omenFlatLoss = map[string]float64{
	"quake": 0.04, "famine": 0.10, "market_closed": 0.05, "heavy_tax": 0.03,
	"shortage": 0.05, "narrow_path": 0.12,
}

var omenScoring = map[string]bool{
	"solar_eclipse": true, "ghost_rout": true, "red_moon": true,
	"ink_rain": true, "upheaval": true, "discord": true,
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
		if key == "heavy_tax" {
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
