package rules

import "fmt"

// ---------- random helpers ----------

func (m *Match) pickWeighted(w []int) int {
	total := 0
	for _, v := range w {
		if v > 0 {
			total += v
		}
	}
	if total <= 0 {
		return -1
	}
	x := m.rng.Intn(total)
	for i, v := range w {
		if v <= 0 {
			continue
		}
		if x < v {
			return i
		}
		x -= v
	}
	return -1
}

// ---------- shop ----------

// drawColor takes one copy of the line out of the red/black split (weighted by what remains)
// and returns its color. It also decrements Pool[l].
func (m *Match) drawColor(l int) Color {
	red, black := m.PoolRed[l], m.Pool[l]-m.PoolRed[l]
	if red < 0 {
		red = 0
	}
	if black < 0 {
		black = 0
	}
	m.Pool[l]--
	if m.pickWeighted([]int{red, black}) == 1 {
		return Black
	}
	m.PoolRed[l]--
	return Red
}

// majorityColor is the color held by most of n cards, red of which are red.
// With the default group size of 3 a majority always exists; for an even group
// size a tie goes to Red.
func majorityColor(red, n int) Color {
	if red*2 >= n {
		return Red
	}
	return Black
}

func (m *Match) drawShopCard(p *Player) (int, Color) {
	odds := m.cfg.Level(p.Level).ShopOdds
	tierW := make([]int, len(odds))
	for t := range odds {
		for l, def := range m.cfg.Lines {
			if def.Cost == t+1 && m.Pool[l] > 0 && !p.three[l] {
				tierW[t] = odds[t]
				break
			}
		}
	}
	t := m.pickWeighted(tierW)
	if t < 0 {
		return -1, Black
	}
	w := make([]int, len(m.cfg.Lines))
	for l, def := range m.cfg.Lines {
		if def.Cost == t+1 && !p.three[l] {
			w[l] = m.Pool[l]
		}
	}
	l := m.pickWeighted(w)
	if l < 0 {
		return -1, Black
	}
	return l, m.drawColor(l)
}

func (m *Match) returnShop(p *Player) {
	for i, l := range p.Shop {
		if l >= 0 {
			m.Pool[l]++
			if p.ShopColor[i] == Red {
				m.PoolRed[l]++
			}
		}
	}
	p.Shop, p.ShopColor = nil, nil
}

// randomRelic picks a relic the player does not own, by rarity odds, skipping excluded ones.
func (m *Match) randomRelic(p *Player, odds []int, exclude map[int]bool) int {
	return m.randomRelicFiltered(p, odds, exclude, nil)
}

// randomRelicCat prefers relics whose category is not in usedCat; if none qualify it falls
// back to randomRelic.
func (m *Match) randomRelicCat(p *Player, odds []int, exclude map[int]bool, usedCat map[string]bool) int {
	if len(usedCat) > 0 {
		if id := m.randomRelicFiltered(p, odds, exclude, usedCat); id >= 0 {
			return id
		}
	}
	return m.randomRelic(p, odds, exclude)
}

func (m *Match) randomRelicFiltered(p *Player, odds []int, exclude map[int]bool, usedCat map[string]bool) int {
	byR := make([][]int, len(m.cfg.Rules.Rarities))
	owned := map[int]bool{}
	for _, r := range p.Relics {
		owned[r.Relic] = true
	}
	for i := range m.cfg.Relics {
		if owned[i] || exclude[i] || usedCat[m.cfg.CategoryOf(i)] {
			continue
		}
		r := m.cfg.RarityOf(i)
		byR[r] = append(byR[r], i)
	}
	w := append([]int{}, odds...)
	for i := range w {
		if len(byR[i]) == 0 {
			w[i] = 0
		}
	}
	r := m.pickWeighted(w)
	if r < 0 {
		for i := range byR {
			if len(byR[i]) > 0 {
				r = i
				break
			}
		}
		if r < 0 {
			return -1
		}
	}
	return byR[r][m.rng.Intn(len(byR[r]))]
}

func (m *Match) rollShop(p *Player) {
	m.returnShop(p)
	p.Shop = make([]int, m.ShopSize())
	p.ShopColor = make([]Color, len(p.Shop))
	for i := range p.Shop {
		p.Shop[i], p.ShopColor[i] = m.drawShopCard(p)
	}
	p.ShopRelic = -1
	if m.cfg.Rules.Board.ShopRelic {
		p.ShopRelic = m.randomRelic(p, m.cfg.Level(p.Level).RelicOdds, nil)
	}
}

func (m *Match) requirePrep(p *Player) error {
	if m.Phase != PhasePrep || p == nil {
		return ErrWrongPhase
	}
	return nil
}

// Refresh rerolls the shop (cards and relic), using a free refresh first.
func (m *Match) Refresh(p *Player) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	if m.FreeRefreshesLeft(p) > 0 {
		p.FreeRefreshes++
	} else {
		cost := m.RefreshCost(p)
		if p.Gold < cost {
			return ErrNotEnoughGold
		}
		p.Gold -= cost
	}
	p.Refreshed = true
	p.Locked = false
	m.rollShop(p)
	return nil
}

// BuyXP spends gold for experience.
func (m *Match) BuyXP(p *Player) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	e := m.cfg.Rules.Economy
	if p.Level >= m.cfg.MaxLevel() {
		return ErrMaxLevel
	}
	if p.Gold < e.XPCost {
		return ErrNotEnoughGold
	}
	p.Gold -= e.XPCost
	m.addXP(p, e.XPPerBuy)
	return nil
}

// SetLock locks or unlocks the shop for the next round start.
func (m *Match) SetLock(p *Player, locked bool) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	p.Locked = locked
	return nil
}

// SetReady marks the player as done with this phase.
func (m *Match) SetReady(p *Player, ready bool) error {
	if m.Phase == PhaseOver || p == nil {
		return ErrWrongPhase
	}
	p.Ready = ready
	return nil
}

// CanBuyCard reports whether shop slot i can be bought right now.
func (m *Match) CanBuyCard(p *Player, i int) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	if i < 0 || i >= len(p.Shop) || p.Shop[i] < 0 {
		return ErrNoSuchItem
	}
	l := p.Shop[i]
	if p.Gold < m.cfg.Lines[l].Cost {
		return ErrNotEnoughGold
	}
	if len(p.Bench) < m.BenchSize(p) {
		return nil
	}
	if m.copiesForMerge(p, l, 1, p.ShopColor[i]) >= m.cfg.Rules.Merge.GroupSize-1 {
		return nil // completes a merge: total cards does not grow
	}
	return ErrBenchFull
}

// BuyCard buys shop slot i onto the bench, then auto-merges. Returns merged cards (in order).
func (m *Match) BuyCard(p *Player, i int) ([]Card, error) {
	if err := m.CanBuyCard(p, i); err != nil {
		return nil, err
	}
	l := p.Shop[i]
	p.Gold -= m.cfg.Lines[l].Cost
	col := p.ShopColor[i]
	p.Shop[i] = -1 // ShopColor[i] is left as is (meaningless for empty slots)
	nc := Card{UID: m.newUID(), Line: l, Star: 1, Color: col}
	if col == Red {
		nc.Red = 1
	}
	p.Bench = append(p.Bench, nc)
	return m.checkMerges(p), nil
}

// ---------- merge ----------

// copiesForMerge counts the cards that would join a merge group with a card of the given color:
// same line and star, and also same color when Merge.SameColor is set.
func (m *Match) copiesForMerge(p *Player, line, star int, c Color) int {
	if m.cfg.Rules.Merge.SameColor {
		return m.CountCopiesColor(p, line, star, c)
	}
	return m.CountCopies(p, line, star)
}

func (m *Match) checkMerges(p *Player) []Card {
	var made []Card
	for {
		merged := false
		for l := range m.cfg.Lines {
			for s := 1; s < m.cfg.Rules.Merge.MaxStar && !merged; s++ {
				for _, col := range []Color{Red, Black} {
					if m.copiesForMerge(p, l, s, col) >= m.cfg.Rules.Merge.GroupSize {
						made = append(made, m.merge(p, l, s, col))
						merged = true
						break
					}
				}
			}
			if merged {
				break
			}
		}
		if !merged {
			return made
		}
	}
}

// merge combines GroupSize cards of (line, star). With Merge.SameColor only cards of color col
// take part; otherwise col is ignored. The result takes the majority color of the cards used
// (ties, possible only with an even group size, go to Red) and the sum of their Red counts.
// With 赤化 (vermilion) the result is always Red; its Red count is still the sum.
func (m *Match) merge(p *Player, line, star int, col Color) Card {
	need := m.cfg.Rules.Merge.GroupSize
	sameColor := m.cfg.Rules.Merge.SameColor
	joins := func(c Card) bool {
		return c.Line == line && c.Star == star && (!sameColor || c.Color == col)
	}
	taken, bonus, pos, red, redCards := 0, 0, -1, 0, 0
	var board []Card
	for i, c := range p.Board {
		if taken < need && joins(c) {
			taken++
			bonus += c.Bonus
			red += c.Red
			if c.Color == Red {
				redCards++
			}
			if pos < 0 {
				pos = i
				board = append(board, Card{}) // placeholder for the result
			}
			continue
		}
		board = append(board, c)
	}
	var bench []Card
	for _, c := range p.Bench {
		if taken < need && joins(c) {
			taken++
			bonus += c.Bonus
			red += c.Red
			if c.Color == Red {
				redCards++
			}
			continue
		}
		bench = append(bench, c)
	}
	nc := Card{UID: m.newUID(), Line: line, Star: star + 1, Bonus: bonus, Red: red, Color: majorityColor(redCards, need)}

	if m.hasEff(m.effective(p.Relics), "vermilion") {
		nc.Color = Red // 赤化：颜色变红，Red 计数不变（卡池按 Red 归还）
	}
	nc.Bonus += int(m.sumEffParam(m.effective(p.Relics), "tempering", "points"))
	for i := range p.Relics {
		key := m.cfg.Relics[p.Relics[i].Relic].Key
		if key == "furnace" && m.trigger(p, "furnace") {
			p.Relics[i].Value += m.cfg.Relics[p.Relics[i].Relic].Params["per_merge"]
		}
	}
	for _, r := range m.effective(p.Relics) {
		if !r.empty && m.relicKey(r) == "graduation" && m.trigger(p, fmt.Sprintf("graduation%d", r.slot)) {
			p.Gold += int(m.cfg.Relics[r.Relic].Params["gold"])
		}
	}

	placeholder := -1
	for i := range board {
		if board[i].UID == "" {
			placeholder = i
			break
		}
	}
	if placeholder >= 0 {
		board[placeholder] = nc
	} else {
		bench = append(bench, nc)
	}
	p.Board, p.Bench = board, bench

	if nc.Star == m.cfg.Rules.Merge.MaxStar {
		p.three[line] = true
	}
	// 铸剑师 grows on every merge after it exists (not the merge that made it)
	if !(nc.Star == m.cfg.Rules.Merge.MaxStar && m.cfg.Lines[line].Legend.Key == "grow_on_merge") {
		for _, cards := range [][]Card{p.Board, p.Bench} {
			for i := range cards {
				c := &cards[i]
				lg := m.cfg.Lines[c.Line].Legend
				if c.Star == m.cfg.Rules.Merge.MaxStar && lg.Key == "grow_on_merge" && m.trigger(p, "grow"+c.UID) {
					c.Bonus += int(lg.Params["points"])
				}
			}
		}
	}
	return nc
}

// ---------- sell ----------

func (m *Match) findCard(p *Player, uid string) (onBoard bool, idx int, ok bool) {
	for i, c := range p.Board {
		if c.UID == uid {
			return true, i, true
		}
	}
	for i, c := range p.Bench {
		if c.UID == uid {
			return false, i, true
		}
	}
	return false, -1, false
}

func (m *Match) growAccountant(p *Player) {
	for i := range p.Relics {
		def := m.cfg.Relics[p.Relics[i].Relic]
		if def.Key == "accountant" && m.trigger(p, "accountant") {
			p.Relics[i].Value += def.Params["per_sell"]
		}
	}
}

// SellCard sells a card from the board or bench.
func (m *Match) SellCard(p *Player, uid string) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	onBoard, i, ok := m.findCard(p, uid)
	if !ok {
		return ErrNoSuchItem
	}
	var c Card
	if onBoard {
		c = p.Board[i]
		p.Board = append(p.Board[:i:i], p.Board[i+1:]...)
	} else {
		c = p.Bench[i]
		p.Bench = append(p.Bench[:i:i], p.Bench[i+1:]...)
	}
	m.disposeCard(p, c)
	if c.Star >= 2 {
		m.growAccountant(p)
	}
	return nil
}

// disposeCard refunds a card and returns its copies to the pool.
func (m *Match) disposeCard(p *Player, c Card) {
	p.Gold += m.SellPrice(c)
	m.Pool[c.Line] += m.copiesOf(c.Star)
	m.PoolRed[c.Line] += c.Red
	if c.Star == m.cfg.Rules.Merge.MaxStar {
		p.three[c.Line] = false
	}
}

// ---------- relics ----------

// BuyRelic buys the shop relic. choice is a trade index for relics that need one (else ignored).
func (m *Match) BuyRelic(p *Player, choice int) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	id := p.ShopRelic
	if id < 0 {
		return ErrNoSuchItem
	}
	if len(p.Relics) >= m.RelicSlots(p) {
		return ErrRelicSlotsFull
	}
	price := m.cfg.RelicPrice(id)
	if p.Gold < price {
		return ErrNotEnoughGold
	}
	inst, err := m.newRelic(id, choice)
	if err != nil {
		return err
	}
	p.Gold -= price
	p.Relics = append(p.Relics, inst)
	p.ShopRelic = -1
	return nil
}

func (m *Match) newRelic(id, choice int) (RelicInst, error) {
	inst := RelicInst{Relic: id, Choice: -1}
	if m.behaviour(id).choice {
		if choice < 0 || choice >= len(m.cfg.Trades) {
			return inst, ErrBadChoice
		}
		inst.Choice = choice
	}
	return inst, nil
}

// SellRelic sells the relic in bar slot i for half its price.
func (m *Match) SellRelic(p *Player, i int) error {
	if m.Phase != PhasePrep && m.Phase != PhaseVault {
		return ErrWrongPhase
	}
	if i < 0 || i >= len(p.Relics) {
		return ErrNoSuchItem
	}
	r := p.Relics[i]
	p.Relics = append(p.Relics[:i:i], p.Relics[i+1:]...)
	p.Gold += m.cfg.RelicPrice(r.Relic) / 2
	m.growAccountant(p)
	m.fitBoard(p)
	return nil
}

// ---------- layout ----------

// Arrange sets the board to exactly boardUIDs (left to right); every other card goes to the bench
// in its current order. relicOrder, if non-nil, is a permutation of relic bar slots.
func (m *Match) Arrange(p *Player, boardUIDs []string, relicOrder []int) error {
	if err := m.requirePrep(p); err != nil {
		return err
	}
	all := append(append([]Card{}, p.Board...), p.Bench...)
	byUID := map[string]int{}
	for i, c := range all {
		byUID[c.UID] = i
	}
	used := make([]bool, len(all))
	board := make([]Card, 0, len(boardUIDs))
	for _, uid := range boardUIDs {
		i, ok := byUID[uid]
		if !ok || used[i] {
			return ErrBadLayout
		}
		used[i] = true
		board = append(board, all[i])
	}
	var bench []Card
	for i, c := range all {
		if !used[i] {
			bench = append(bench, c)
		}
	}
	if len(bench) > m.BenchSize(p) {
		return ErrBadLayout
	}
	newRelics := p.Relics
	if relicOrder != nil {
		if len(relicOrder) != len(p.Relics) {
			return ErrBadLayout
		}
		seen := make([]bool, len(p.Relics))
		newRelics = make([]RelicInst, 0, len(p.Relics))
		for _, i := range relicOrder {
			if i < 0 || i >= len(p.Relics) || seen[i] {
				return ErrBadLayout
			}
			seen[i] = true
			newRelics = append(newRelics, p.Relics[i])
		}
	}
	// Relic order changes which relics a mirror/rubbing copies, so the slot limit is
	// checked under the NEW order; nothing is mutated unless the layout is valid.
	old := p.Relics
	p.Relics = newRelics
	if len(board) > m.Slots(p) {
		p.Relics = old
		return ErrBadLayout
	}
	p.Board, p.Bench = board, bench
	return nil
}
