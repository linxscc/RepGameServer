// Package chessbot plays one side of a rules.Match using only the engine's public API.
// Each bot follows an archetype (a preferred set of lines and relics); all bots share
// the same decision loop, ported from the offline rules simulator.
package chessbot

import (
	"math/rand"

	"GoServer/tcpgameserver/rules"
)

// Archetype describes what a bot wants. Keys refer to chess_*.yaml.
type Archetype struct {
	Name        string
	Lines       []string           // preferred line keys; empty = build whatever it already owns
	Wide        bool               // wants many distinct lines
	Relics      map[string]float64 // relic key → desirability; nil = take anything, rarer first
	Families    []string           // family keys, best first
	Trade       string             // trade key for relics that need a choice
	Loop        bool               // buy-merge-sell loop for growth relics
	SaveTo      int                // interest reserve while healthy
	LevelPlan   []int              // LevelPlan[lv] = round from which to be at least lv (0 = never)
	Adaptive    bool               // re-picks its lines every round: what it owns, what is left in the pool, what the opponent is not taking
	Deny        bool               // hate-buys a cheap copy the opponent needs for a 3★
	Color       string             // "" = ignore card color; "red" / "black" = prefer that color
	Kind        string             // "" / "god" / "ghost": prefer lines of that kind
	IgnoreSuits bool               // baseline: no suit synergy; suit/counter relics are worth a fixed 0.1
}

// Archetypes are the playstyles a bot can pick, mirroring the simulator.
func Archetypes() []Archetype {
	normal := []int{0, 0, 0, 0, 9, 14, 20}
	fast := []int{0, 0, 0, 4, 7, 11, 16}
	return []Archetype{
		{Name: "单核爆发", Lines: []string{"soldier", "clerk", "trader"}, SaveTo: 30, LevelPlan: fast, Trade: "court",
			Relics:   map[string]float64{"lone_wolf": 5, "lean": 4, "vanguard": 4, "tailwind": 5, "master_seal": 4, "mirror": 4, "rubbing": 3, "heirloom": 2, "academy": 2},
			Families: []string{"general", "merchant", "hermit"}},
		{Name: "合成循环", Lines: []string{"carpenter", "smith", "farmer"}, Loop: true, SaveTo: 20, LevelPlan: normal, Trade: "craft",
			Relics:   map[string]float64{"furnace": 5, "accountant": 5, "graduation": 4, "tempering": 4, "heirloom": 3, "mirror": 3, "rubbing": 3, "tailwind": 3, "fellowship": 2},
			Families: []string{"artisan", "merchant", "scholar"}},
		{Name: "方外重触发", Lines: []string{"taoist", "monk", "doctor"}, SaveTo: 30, LevelPlan: normal, Trade: "monk",
			Relics:   map[string]float64{"tailwind": 5, "mirror": 5, "master_seal": 4, "fellowship": 3, "rubbing": 3, "multitude": 2, "expansion": 3},
			Families: []string{"artisan", "noble", "scholar"}},
		{Name: "人海羁绊", Lines: []string{"carpenter", "farmer", "monk", "soldier", "taoist", "scholar", "designer", "clerk"}, Wide: true, SaveTo: 20, LevelPlan: fast, Trade: "craft",
			Relics:   map[string]float64{"eclectic": 5, "fellowship": 4, "multitude": 4, "new_wave": 3, "sect_token": 3, "tailwind": 4, "mirror": 3, "market_pass": 3, "academy": 3},
			Families: []string{"general", "merchant", "noble"}},
		{Name: "藏兵于席", Lines: []string{"carpenter", "smith", "farmer", "taoist", "monk", "scholar"}, SaveTo: 20, LevelPlan: normal, Trade: "craft",
			Relics:   map[string]float64{"hidden": 5, "storehouse": 5, "lone_wolf": 4, "vanguard": 3, "tailwind": 3, "mirror": 3, "rubbing": 2},
			Families: []string{"merchant", "noble", "hermit"}},
		{Name: "均衡", SaveTo: 30, LevelPlan: normal, Trade: "letter", Families: []string{"merchant", "scholar", "noble"}},
		{Name: "应变", Adaptive: true, Deny: true, SaveTo: 30, LevelPlan: normal, Trade: "letter",
			Relics:   map[string]float64{"tailwind": 4, "mirror": 4, "master_seal": 3, "fellowship": 3, "furnace": 3, "multitude": 2, "rubbing": 2},
			Families: []string{"merchant", "artisan", "noble"}},
	}
}

// Bot is one bot's resolved preferences for a given config.
type Bot struct {
	Arch   Archetype
	rng    *rand.Rand
	lines  map[int]bool
	relics map[int]float64
	trade  int
}

// New resolves an archetype against the config; unknown keys are ignored.
func New(m *rules.Match, arch Archetype, rng *rand.Rand) *Bot {
	cfg := m.Config()
	b := &Bot{Arch: arch, rng: rng, lines: map[int]bool{}}
	for _, k := range arch.Lines {
		if i, ok := cfg.LineIndex(k); ok {
			b.lines[i] = true
		}
	}
	if arch.Relics != nil {
		b.relics = map[int]float64{}
		for k, v := range arch.Relics {
			if i, ok := cfg.RelicIndex(k); ok {
				b.relics[i] = v
			}
		}
	}
	for i, t := range cfg.Trades {
		if t.Key == arch.Trade {
			b.trade = i
		}
	}
	return b
}

// NewRandom picks a random normal-difficulty archetype. The adaptive one (应变) is much
// stronger than the rest (about 85% against them in tools/balance), so it is left out.
func NewRandom(m *rules.Match, rng *rand.Rand) *Bot {
	var normal []Archetype
	for _, a := range Archetypes() {
		if !a.Adaptive {
			normal = append(normal, a)
		}
	}
	normal = append(normal, SuitArchetypes()...)
	return New(m, normal[rng.Intn(len(normal))], rng)
}

// ---------- phase decisions ----------

// PickFamily chooses the best offered family.
func (b *Bot) PickFamily(m *rules.Match, p *rules.Player) {
	cfg := m.Config()
	for _, want := range b.Arch.Families {
		for i, f := range p.FamilyOffer {
			if cfg.Families[f].Key == want {
				_ = m.PickFamily(p, i)
				return
			}
		}
	}
	_ = m.PickFamily(p, 0)
}

// PickVault takes the most wanted relic, selling the weakest one if the bar is full.
func (b *Bot) PickVault(m *rules.Match, p *rules.Player) {
	best, bv := -1, -1.0
	for i, id := range p.VaultOffer {
		if v := b.relicValue(m, p, id); v > bv {
			best, bv = i, v
		}
	}
	if best < 0 {
		_ = m.SkipVault(p)
		return
	}
	if len(p.Relics) < m.RelicSlots(p) {
		_ = m.PickVault(p, best, -1, b.trade)
		return
	}
	wi, wv := b.worstRelic(m, p)
	if bv > wv+0.5 {
		_ = m.PickVault(p, best, wi, b.trade)
		return
	}
	_ = m.SkipVault(p)
}

// Prep plays a whole prep phase and marks the bot ready.
func (b *Bot) Prep(m *rules.Match, p *rules.Player) {
	if m.Phase != rules.PhasePrep {
		return
	}
	if b.Arch.Adaptive {
		b.retarget(m, p)
	}
	b.relicPass(m, p)
	for lv := 2; lv < len(b.Arch.LevelPlan) && lv <= m.Config().MaxLevel(); lv++ {
		for p.Level < lv && b.Arch.LevelPlan[lv] > 0 && m.Round >= b.Arch.LevelPlan[lv] && p.Gold-m.Config().Rules.Economy.XPCost >= b.reserve(p)/2 {
			if m.BuyXP(p) != nil {
				break
			}
		}
	}
	b.buyPass(m, p)
	if b.Arch.Deny {
		b.denyPass(m, p)
	}
	for rolls := 0; rolls < 12; rolls++ {
		free := m.FreeRefreshesLeft(p) > 0
		if !free && (m.Round < 4 || p.Gold-m.Config().Rules.Economy.RefreshCost < b.reserve(p)+2) {
			break
		}
		if m.Refresh(p) != nil {
			break
		}
		b.relicPass(m, p)
		b.buyPass(m, p)
	}
	if b.Arch.Loop {
		b.loopSell(m, p)
	}
	b.arrange(m, p)
	for _, l := range p.Shop {
		if l >= 0 && b.wants(p, m, l) {
			_ = m.SetLock(p, true)
			break
		}
	}
	_ = m.SetReady(p, true)
}

// ---------- adaptive play ----------

// retarget picks the three lines worth building this round: copies already owned weigh
// most, then copies still in the pool; lines the opponent fielded last round are avoided.
func (b *Bot) retarget(m *rules.Match, p *rules.Player) {
	cfg := m.Config()
	opp := m.Opponent(p)
	oppHas := map[int]int{}
	for _, c := range opp.LastBoard {
		oppHas[c.Line] += c.Star
	}
	type cand struct {
		line  int
		score float64
	}
	var cs []cand
	for l, def := range cfg.Lines {
		if m.HasThreeStar(p, l) {
			continue
		}
		owned := 0
		for _, cards := range [][]rules.Card{p.Board, p.Bench} {
			for _, c := range cards {
				if c.Line == l {
					owned += pow(cfg.Rules.Merge.GroupSize, c.Star-1)
				}
			}
		}
		left := float64(m.Pool[l]) / float64(def.Pool)
		s := 3*float64(owned) + 4*left - 2.5*float64(oppHas[l])
		if odds := cfg.Level(p.Level).ShopOdds; def.Cost-1 < len(odds) && odds[def.Cost-1] == 0 {
			s -= 2 // the shop cannot offer this cost at the current level
		}
		cs = append(cs, cand{l, s})
	}
	for i := range cs { // simple selection of the top 3
		for j := i + 1; j < len(cs); j++ {
			if cs[j].score > cs[i].score {
				cs[i], cs[j] = cs[j], cs[i]
			}
		}
	}
	b.lines = map[int]bool{}
	for i := 0; i < len(cs) && i < 3; i++ {
		b.lines[cs[i].line] = true
	}
}

// denyPass buys a cheap shop card the opponent fielded at 2★ last round when the pool is
// nearly out, so they cannot finish the 3★. Only with spare gold and bench room.
func (b *Bot) denyPass(m *rules.Match, p *rules.Player) {
	cfg := m.Config()
	need := cfg.Rules.Merge.GroupSize
	opp := m.Opponent(p)
	threat := map[int]bool{}
	for _, c := range opp.LastBoard {
		if c.Star == cfg.Rules.Merge.MaxStar-1 {
			threat[c.Line] = true
		}
	}
	for i, l := range p.Shop {
		if l < 0 || !threat[l] || m.Pool[l] > need {
			continue
		}
		if p.Gold-cfg.Lines[l].Cost < b.reserve(p) || len(p.Bench) >= cfg.Rules.Board.BenchSize {
			continue
		}
		_, _ = m.BuyCard(p, i)
	}
}

// offPrefOK says whether a card of the non-preferred color may still be bought (merge
// completions are handled by the caller): never when the preferred color of the same line is
// on offer, and otherwise only while the roster is so thin the board would be left hollow.
func (b *Bot) offPrefOK(m *rules.Match, p *rules.Player, line int) bool {
	pc, _ := b.prefColor(m, p)
	for j, l2 := range p.Shop {
		if l2 == line && j < len(p.ShopColor) && p.ShopColor[j] == pc {
			return false
		}
	}
	return len(p.Board)+len(p.Bench) < m.Slots(p)-1 // only to avoid a hollow board
}

// ColorPeak is the largest number of fielded cards sharing one color.
func ColorPeak(m *rules.Match, p *rules.Player) int {
	var n [2]int
	for _, c := range p.Board {
		if c.Color == rules.Red || c.Color == rules.Black {
			n[c.Color]++
		}
	}
	if n[0] > n[1] {
		return n[0]
	}
	return n[1]
}

func pow(b, e int) int {
	r := 1
	for i := 0; i < e; i++ {
		r *= b
	}
	return r
}

// ---------- helpers ----------

func (b *Bot) reserve(p *rules.Player) int {
	if p.HP <= 35 {
		return 0
	}
	return b.Arch.SaveTo
}

func (b *Bot) relicValue(m *rules.Match, p *rules.Player, id int) float64 {
	cfg := m.Config()
	if b.Arch.IgnoreSuits {
		if cat := cfg.CategoryOf(id); cat == rules.CatSuit || cat == rules.CatCounter {
			return 0.1
		}
	}
	rarity := float64(cfg.RarityOf(id))
	var v float64
	if b.relics == nil {
		v = 1 + 0.5*rarity
	} else if bv, ok := b.relics[id]; ok {
		v = bv
	} else {
		v = 0.3 + 0.3*rarity
	}
	if !b.Arch.IgnoreSuits {
		v += 4 * synergy(m, p, id)
	}
	return v
}

func (b *Bot) worstRelic(m *rules.Match, p *rules.Player) (int, float64) {
	idx, worst := -1, 1e9
	for i, r := range p.Relics {
		v := b.relicValue(m, p, r.Relic) + r.Value/4
		if v < worst {
			idx, worst = i, v
		}
	}
	return idx, worst
}

func ownsLine(p *rules.Player, line int) bool {
	for _, cs := range [][]rules.Card{p.Board, p.Bench} {
		for _, c := range cs {
			if c.Line == line {
				return true
			}
		}
	}
	return false
}

func (b *Bot) wants(p *rules.Player, m *rules.Match, line int) bool {
	if len(b.lines) == 0 {
		return ownsLine(p, line)
	}
	return b.lines[line]
}

func (b *Bot) cardWorth(m *rules.Match, p *rules.Player, c rules.Card) float64 {
	v := float64(m.Points(c))
	if b.wants(p, m, c.Line) {
		v += 20
	}
	return v
}

func (b *Bot) makeRoom(m *rules.Match, p *rules.Player, forLine int) bool {
	if len(p.Bench) < m.Config().Rules.Board.BenchSize {
		return true
	}
	uid, worst := "", 1e9
	for _, c := range p.Bench {
		if c.Line == forLine {
			continue
		}
		if w := b.cardWorth(m, p, c); w < worst {
			uid, worst = c.UID, w
		}
	}
	if uid == "" || worst >= 20 {
		return false
	}
	return m.SellCard(p, uid) == nil
}

func (b *Bot) buyPass(m *rules.Match, p *rules.Player) {
	cfg := m.Config()
	res := b.reserve(p)
	for i, l := range p.Shop {
		if l < 0 {
			continue
		}
		cost := cfg.Lines[l].Cost
		pair := m.CountCopies(p, l, 1) >= 1 || m.CountCopies(p, l, 2) >= 1
		completes := m.CountCopies(p, l, 1) >= cfg.Rules.Merge.GroupSize-1
		pk, hasKind := b.prefKind(m, p)
		want := b.wants(p, m, l) || (hasKind && cfg.Lines[l].Kind == pk) || (b.Arch.Wide && !ownsLine(p, l)) || (b.Arch.Loop && cost == 1 && pair) ||
			(m.Round <= 3 && cost == 1) || (completes && cost <= 2) ||
			len(p.Board)+len(p.Bench) < m.Slots(p)+1
		if !want {
			continue
		}
		if pc, ok := b.prefColor(m, p); ok && !completes && i < len(p.ShopColor) && p.ShopColor[i] != pc && !b.offPrefOK(m, p, l) {
			continue
		}
		limit := res
		if completes {
			limit = res - 6
		}
		if p.Gold-cost < limit {
			continue
		}
		if !completes && !b.makeRoom(m, p, l) {
			continue
		}
		_, _ = m.BuyCard(p, i)
	}
}

func (b *Bot) relicPass(m *rules.Match, p *rules.Player) {
	id := p.ShopRelic
	if id < 0 {
		return
	}
	v := b.relicValue(m, p, id)
	if v < 1 || p.Gold-m.Config().RelicPrice(id) < b.reserve(p)/2 {
		return
	}
	if len(p.Relics) >= m.RelicSlots(p) {
		wi, wv := b.worstRelic(m, p)
		if v <= wv+0.5 || m.SellRelic(p, wi) != nil {
			return
		}
	}
	_ = m.BuyRelic(p, b.trade)
}

func (b *Bot) loopSell(m *rules.Match, p *rules.Player) {
	hasAccountant := false
	for _, r := range p.Relics {
		if m.Config().Relics[r.Relic].Key == "accountant" {
			hasAccountant = true
		}
	}
	if !hasAccountant {
		return
	}
	for i := len(p.Bench) - 1; i >= 0; i-- {
		if c := p.Bench[i]; c.Star == 2 && !b.wants(p, m, c.Line) {
			_ = m.SellCard(p, c.UID)
		}
	}
}

// arrange picks the board and relic order that maximise this round's score.
func (b *Bot) arrange(m *rules.Match, p *rules.Player) {
	all := append(append([]rules.Card{}, p.Board...), p.Bench...)
	slots := m.Slots(p)
	benchSize := m.Config().Rules.Board.BenchSize
	split := func(idx []int) ([]rules.Card, []rules.Card) {
		used := make([]bool, len(all))
		board := make([]rules.Card, 0, len(idx))
		for _, i := range idx {
			used[i] = true
			board = append(board, all[i])
		}
		bench := make([]rules.Card, 0, len(all))
		for i, c := range all {
			if !used[i] {
				bench = append(bench, c)
			}
		}
		return board, bench
	}
	pc, colored := b.prefColor(m, p)
	eval := func(idx []int, relics []rules.RelicInst) float64 {
		board, bench := split(idx)
		s := m.ScoreWith(p, board, bench, relics).Score
		if colored { // tie-break only: among equal scores, field the preferred color
			for _, c := range board {
				if c.Color == pc {
					s += 1e-6
				}
			}
		}
		return s
	}
	order := identity(len(p.Relics))
	relics := p.Relics
	var best []int
	for pass := 0; pass < 2; pass++ {
		best = greedyBoard(len(all), slots, benchSize, all, func(idx []int) float64 { return eval(idx, relics) })
		if len(p.Relics) > 1 {
			bs := -1.0
			permute(identity(len(p.Relics)), func(perm []int) {
				rs := make([]rules.RelicInst, len(perm))
				for i, k := range perm {
					rs[i] = p.Relics[k]
				}
				if s := eval(best, rs); s > bs {
					bs, order = s, append([]int{}, perm...)
				}
			})
			relics = make([]rules.RelicInst, len(order))
			for i, k := range order {
				relics[i] = p.Relics[k]
			}
		}
	}
	uids := make([]string, len(best))
	for i, k := range best {
		uids[i] = all[k].UID
	}
	_ = m.Arrange(p, uids, order)
}

func identity(n int) []int {
	x := make([]int, n)
	for i := range x {
		x[i] = i
	}
	return x
}

func permute(a []int, f func([]int)) {
	var rec func(k int)
	rec = func(k int) {
		if k == len(a) {
			f(a)
			return
		}
		for i := k; i < len(a); i++ {
			a[k], a[i] = a[i], a[k]
			rec(k + 1)
			a[k], a[i] = a[i], a[k]
		}
	}
	rec(0)
}

// greedyBoard adds the best card while it helps, then tries replacements, removals and
// position swaps; finally fills the board if the bench would overflow.
func greedyBoard(n, slots, benchSize int, all []rules.Card, eval func([]int) float64) []int {
	var cur []int
	in := make([]bool, n)
	curS := 0.0
	for len(cur) < slots {
		bi, bs := -1, curS
		for i := 0; i < n; i++ {
			if in[i] {
				continue
			}
			if s := eval(append(append([]int{}, cur...), i)); s > bs {
				bi, bs = i, s
			}
		}
		if bi < 0 {
			break
		}
		cur, in[bi], curS = append(cur, bi), true, bs
	}
	for improved, iter := true, 0; improved && iter < 20; iter++ {
		improved = false
		for k := 0; k < len(cur) && !improved; k++ {
			for i := 0; i < n; i++ {
				if in[i] {
					continue
				}
				old := cur[k]
				cur[k] = i
				if s := eval(cur); s > curS {
					in[old], in[i], curS, improved = false, true, s, true
					break
				}
				cur[k] = old
			}
		}
	}
	for improved := true; improved; {
		improved = false
		for k := range cur {
			cand := append(append([]int{}, cur[:k]...), cur[k+1:]...)
			if s := eval(cand); s > curS {
				in[cur[k]] = false
				cur, curS, improved = cand, s, true
				break
			}
		}
	}
	for improved := true; improved; {
		improved = false
		for i := 0; i < len(cur) && !improved; i++ {
			for j := i + 1; j < len(cur); j++ {
				cur[i], cur[j] = cur[j], cur[i]
				if s := eval(cur); s > curS {
					curS, improved = s, true
					break
				}
				cur[i], cur[j] = cur[j], cur[i]
			}
		}
	}
	for n-len(cur) > benchSize && len(cur) < slots {
		bi := -1
		for i := 0; i < n; i++ {
			if !in[i] && (bi < 0 || all[i].Star > all[bi].Star) {
				bi = i
			}
		}
		cur, in[bi] = append(cur, bi), true
	}
	return cur
}
