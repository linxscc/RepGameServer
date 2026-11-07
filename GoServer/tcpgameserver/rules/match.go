package rules

import (
	"math"
	"math/rand"
)

// NewMatch creates a match in the family-pick phase. rng makes it deterministic.
func NewMatch(cfg *Config, rng *rand.Rand, ids [2]string) *Match {
	m := &Match{cfg: cfg, rng: rng, Phase: PhaseFamily, Winner: -1, Pool: make([]int, len(cfg.Lines)), PoolRed: make([]int, len(cfg.Lines)), LastRoll: OmenRoll{Omen: -1}}
	for i, l := range cfg.Lines {
		m.Pool[i] = l.Pool
		m.PoolRed[i] = l.Pool / 2
	}
	for i := range m.P {
		p := &Player{
			ID: ids[i], HP: cfg.Rules.Player.StartHP, Gold: cfg.Rules.Player.StartGold, Level: 1,
			Family: -1, ShopRelic: -1, triggers: map[string]int{}, three: make([]bool, len(cfg.Lines)),
		}
		perm := rng.Perm(len(cfg.Families))
		p.FamilyOffer = perm[:cfg.Rules.Player.FamilyOffer]
		m.P[i] = p
	}
	return m
}

// PickFamily chooses one of the offered families (index into FamilyOffer).
func (m *Match) PickFamily(p *Player, offerIdx int) error {
	if m.Phase != PhaseFamily || p == nil {
		return ErrWrongPhase
	}
	if p.Family >= 0 {
		return ErrAlreadyChosen
	}
	if offerIdx < 0 || offerIdx >= len(p.FamilyOffer) {
		return ErrBadChoice
	}
	p.Family = p.FamilyOffer[offerIdx]
	p.Ready = true
	return nil
}

// FamiliesChosen reports whether both players picked a family.
func (m *Match) FamiliesChosen() bool { return m.P[0].Family >= 0 && m.P[1].Family >= 0 }

// Begin applies families, deals the opening cards and starts round 1.
// Players who have not picked get a random offered family.
func (m *Match) Begin() error {
	if m.Phase != PhaseFamily {
		return ErrWrongPhase
	}
	for _, p := range m.P {
		if p.Family < 0 {
			p.Family = p.FamilyOffer[m.rng.Intn(len(p.FamilyOffer))]
		}
		m.applyFamily(p)
	}
	order := m.rng.Perm(2)
	m.FirstDealt = order[0]
	for _, i := range order {
		m.dealStart(m.P[i])
	}
	m.startRound()
	return nil
}

func (m *Match) applyFamily(p *Player) {
	f := m.cfg.Families[p.Family]
	switch f.Key {
	case "merchant":
		p.Gold += int(f.Params["gold"])
	case "scholar":
		odds := make([]int, len(m.cfg.Rules.Rarities))
		odds[0] = 1
		for n := 0; n < int(f.Params["relics"]) && len(p.Relics) < m.RelicSlots(p); n++ {
			if id := m.randomRelic(p, odds, nil); id >= 0 {
				inst, err := m.newRelic(id, m.rng.Intn(len(m.cfg.Trades)))
				if err == nil {
					p.Relics = append(p.Relics, inst)
				}
			}
		}
	case "general":
		if lv := int(f.Params["level"]); lv > p.Level && lv <= m.cfg.MaxLevel() {
			p.Level = lv
		}
	}
}

func (m *Match) dealStart(p *Player) {
	r := m.cfg.Rules.Player
	var got []int
	artisanStar := 1
	if v, ok := m.familyParam(p, "artisan", "star"); ok {
		artisanStar = int(v)
	}
	for len(got) < r.StartCards {
		w := make([]int, len(m.cfg.Lines))
		for l, def := range m.cfg.Lines {
			if def.Cost == r.StartCardCost && !containsInt(got, l) {
				w[l] = m.Pool[l]
			}
		}
		star := 1
		if len(got) == 0 {
			star = artisanStar
		}
		need := m.copiesOf(star)
		for l := range w {
			if m.Pool[l] < need {
				w[l] = 0
			}
		}
		l := m.pickWeighted(w)
		if l < 0 {
			return
		}
		got = append(got, l)
		red := 0
		for i := 0; i < need; i++ {
			if m.drawColor(l) == Red { // drawColor also takes the copy out of Pool
				red++
			}
		}
		c := Card{UID: m.newUID(), Line: l, Star: star, Red: red, Color: majorityColor(red, need)}
		if star == m.cfg.Rules.Merge.MaxStar {
			p.three[l] = true
		}
		if len(p.Board) < m.Slots(p) {
			p.Board = append(p.Board, c)
		} else {
			p.Bench = append(p.Bench, c)
		}
	}
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// startRound advances the round and enters the omen phase on omen rounds, else the prep phase.
func (m *Match) startRound() {
	m.Round++
	for _, p := range m.P {
		p.OmenChoice = OmenUndecided
		p.Ready = false
		p.VaultOffer = nil
	}
	if m.cfg.IsOmenRound(m.Round) {
		m.Phase = PhaseOmen
		return
	}
	m.beginPrep()
}

// ChooseOmen records a player's secret roll (true) or skip (false) decision.
func (m *Match) ChooseOmen(p *Player, roll bool) error {
	if m.Phase != PhaseOmen || p == nil {
		return ErrWrongPhase
	}
	if p.OmenChoice != OmenUndecided {
		return ErrAlreadyChosen
	}
	p.OmenChoice = OmenSkips
	if roll {
		p.OmenChoice = OmenRolls
	}
	p.Ready = true
	return nil
}

// ResolveOmen rolls the dice (undecided counts as skip; if both skip, the system rolls
// for both), adds the omen and starts the prep phase.
func (m *Match) ResolveOmen() error {
	if m.Phase != PhaseOmen {
		return ErrWrongPhase
	}
	sides := m.cfg.Rules.Omen.DieSides
	var dice [2]int
	for i, p := range m.P {
		if p.OmenChoice == OmenRolls {
			dice[i] = m.rng.Intn(sides) + 1
		}
	}
	forced := dice[0] == 0 && dice[1] == 0
	if forced {
		for i := range dice {
			dice[i] = m.rng.Intn(sides) + 1
		}
	}
	m.resolveOmenDice(dice, forced)
	return nil
}

// resolveOmenDice turns dice into an omen: a point whose omen is already active moves up
// by one (wrapping from MaxOmenPoint to 1) until it is not, at most MaxOmenPoint times.
func (m *Match) resolveOmenDice(dice [2]int, forced bool) {
	point := dice[0] + dice[1]
	shifted := false
	maxPt := m.cfg.MaxOmenPoint()
	for n := 0; n < maxPt && m.omenActiveIdx(m.cfg.OmenByPoint(point)); n++ {
		point = point%maxPt + 1
		shifted = true
	}
	omen := m.cfg.OmenByPoint(point)
	m.Omens = append(m.Omens, ActiveOmen{Omen: omen, From: m.Round, Until: m.Round + m.cfg.Rules.Omen.Duration - 1})
	m.LastRoll = OmenRoll{Round: m.Round, Dice: dice, Forced: forced, Shifted: shifted, Point: point, Omen: omen}
	m.beginPrep()
}

// beginPrep pays income, grants xp, refreshes shops and enters the prep phase.
func (m *Match) beginPrep() {
	m.Phase = PhasePrep
	e := m.cfg.Rules.Economy
	for _, p := range m.P {
		p.triggers = map[string]int{}
		p.Refreshed = false
		p.FreeRefreshes = 0
		p.Ready = false

		income := m.Interest(p) + e.BaseIncome + int(m.omenSum("famine", "income", nil))
		if v, ok := m.familyParam(p, "miser", "income_delta"); ok {
			income += int(v)
		}
		streak := p.WinStreak
		if p.LossStreak > streak {
			streak = p.LossStreak
		}
		income += m.cfg.StreakGold(streak)
		if p.WonLast {
			income += e.WinBonus
		}
		if v, ok := m.familyParam(p, "poor", "gold_per_loss"); ok && p.LostLast {
			income += int(v)
		}
		if income < 0 {
			income = 0
		}
		p.Gold += income

		if m.Round >= e.AutoXPFromRound {
			m.addXP(p, e.AutoXP+int(m.sumEffParam(m.effective(p.Relics), "academy", "xp")))
		}

		if p.Locked && p.LockStreak < m.cfg.Rules.Board.MaxLockRounds && p.Shop != nil {
			p.LockStreak++
		} else {
			p.LockStreak = 0
			m.rollShop(p)
		}
		p.Locked = false
		m.fitBoard(p)
	}
}

// fitBoard moves board cards beyond the slots onto the bench, then sells bench cards
// from the right end until the bench fits (地动 shrinks it; sells refund the sell price).
func (m *Match) fitBoard(p *Player) {
	for len(p.Board) > m.Slots(p) {
		c := p.Board[len(p.Board)-1]
		p.Board = p.Board[:len(p.Board)-1]
		p.Bench = append(p.Bench, c)
	}
	for len(p.Bench) > m.BenchSize(p) {
		c := p.Bench[len(p.Bench)-1]
		p.Bench = p.Bench[:len(p.Bench)-1]
		m.disposeCard(p, c)
	}
}

// RoundResult is what a settlement produced.
type RoundResult struct {
	Round  int
	Scores [2]ScoreResult
	Damage [2]int // damage each player took
	Winner int    // round winner, -1 = tie
	Over   bool
	Vault  bool // a vault pick follows
}

// Settle ends the prep phase: applies end-of-prep effects, scores both boards,
// deals damage, updates streaks and decides whether the match is over.
func (m *Match) Settle() (RoundResult, error) {
	if m.Phase != PhasePrep {
		return RoundResult{}, ErrWrongPhase
	}
	res := RoundResult{Round: m.Round, Winner: -1}
	for _, p := range m.P {
		for _, r := range m.effective(p.Relics) {
			if !r.empty && m.relicKey(r) == "thrifty" && !p.Refreshed {
				p.Gold += int(m.cfg.Relics[r.Relic].Params["gold"])
			}
		}
	}
	for i, p := range m.P {
		res.Scores[i] = m.Score(p)
	}
	stage := m.cfg.StageDamage(m.Round)
	s0, s1 := res.Scores[0].Score, res.Scores[1].Score
	if s0 == s1 {
		for i, p := range m.P {
			res.Damage[i] = m.damageTo(p, stage)
			m.lose(p)
		}
	} else {
		w := 0
		if s1 > s0 {
			w = 1
		}
		l := 1 - w
		hi, lo := res.Scores[w].Score, res.Scores[l].Score
		if lo <= 0 {
			lo = 1
		}
		raw := int(math.Ceil(float64(stage) * m.cfg.SuppressionFactor(hi/lo)))
		res.Damage[l] = m.damageTo(m.P[l], raw)
		res.Winner = w
		m.lose(m.P[l])
		m.win(m.P[w])
	}
	for i, p := range m.P {
		p.HP -= res.Damage[i]
		p.LastBoard = append([]Card{}, p.Board...)
		p.Ready = false
	}
	a, b := m.P[0].HP, m.P[1].HP
	if a <= 0 || b <= 0 || m.Round >= m.cfg.Rules.Battle.MaxRounds {
		m.Phase = PhaseOver
		switch {
		case a > b:
			m.Winner = 0
		case b > a:
			m.Winner = 1
		default:
			m.Winner = -1
		}
		res.Over = true
		return res, nil
	}
	if m.cfg.IsVaultRound(m.Round) {
		m.openVault()
		res.Vault = true
		return res, nil
	}
	m.startRound()
	return res, nil
}

func (m *Match) damageTo(p *Player, dmg int) int {
	for _, s := range m.legendsOn(p.Board) {
		if s.Key == "damage_taken" {
			return int(math.Ceil(float64(dmg) * s.Params["factor"]))
		}
	}
	return dmg
}

func (m *Match) lose(p *Player) {
	p.WinStreak = 0
	p.LossStreak++
	p.WonLast, p.LostLast = false, true
	for i := range p.Relics {
		def := m.cfg.Relics[p.Relics[i].Relic]
		if def.Key == "grudge" {
			p.Relics[i].Value = math.Min(p.Relics[i].Value+def.Params["per_loss"], def.Params["cap"])
		}
	}
}

func (m *Match) win(p *Player) {
	p.LossStreak = 0
	p.WinStreak++
	p.WonLast, p.LostLast = true, false
}

// Surrender ends the match at once; the surrendering player loses.
func (m *Match) Surrender(p *Player) error {
	if m.Phase == PhaseOver || p == nil {
		return ErrWrongPhase
	}
	m.Phase = PhaseOver
	m.Winner = 0
	if m.P[0] == p {
		m.Winner = 1
	}
	m.Surrendered = true
	return nil
}

// ---------- vault ----------

func (m *Match) openVault() {
	m.Phase = PhaseVault
	for _, p := range m.P {
		n := m.cfg.Rules.Vault.Offer
		if v, ok := m.familyParam(p, "noble", "extra_offer"); ok {
			n += int(v)
		}
		ex := map[int]bool{}
		usedCat := map[string]bool{}
		p.VaultOffer = nil
		for len(p.VaultOffer) < n {
			id := m.randomRelicCat(p, m.cfg.VaultOddsFor(m.Round), ex, usedCat)
			if id < 0 {
				break
			}
			ex[id] = true
			usedCat[m.cfg.CategoryOf(id)] = true
			p.VaultOffer = append(p.VaultOffer, id)
		}
		p.Ready = len(p.VaultOffer) == 0
	}
}

// PickVault takes offer[offerIdx]; if the relic bar is full, sellSlot names the relic to sell
// first (-1 = refuse). choice is the trade for relics that need one.
func (m *Match) PickVault(p *Player, offerIdx, sellSlot, choice int) error {
	if m.Phase != PhaseVault || p == nil {
		return ErrWrongPhase
	}
	if p.Ready {
		return ErrAlreadyChosen
	}
	if offerIdx < 0 || offerIdx >= len(p.VaultOffer) {
		return ErrBadChoice
	}
	id := p.VaultOffer[offerIdx]
	inst, err := m.newRelic(id, choice)
	if err != nil {
		return err
	}
	if len(p.Relics) >= m.RelicSlots(p) {
		if sellSlot < 0 {
			return ErrRelicSlotsFull
		}
		if err := m.SellRelic(p, sellSlot); err != nil {
			return err
		}
	}
	p.Relics = append(p.Relics, inst)
	p.VaultOffer = nil
	p.Ready = true
	return nil
}

// SkipVault declines the vault for gold.
func (m *Match) SkipVault(p *Player) error {
	if m.Phase != PhaseVault || p == nil {
		return ErrWrongPhase
	}
	if p.Ready {
		return ErrAlreadyChosen
	}
	p.Gold += m.cfg.Rules.Vault.SkipGold
	p.VaultOffer = nil
	p.Ready = true
	return nil
}

// AutoVault resolves an undecided player on timeout: a random pick, or skip if the bar is full.
func (m *Match) AutoVault(p *Player) {
	if m.Phase != PhaseVault || p.Ready {
		return
	}
	if len(p.Relics) >= m.RelicSlots(p) || len(p.VaultOffer) == 0 {
		_ = m.SkipVault(p)
		return
	}
	_ = m.PickVault(p, m.rng.Intn(len(p.VaultOffer)), -1, m.rng.Intn(len(m.cfg.Trades)))
}

// CloseVault ends the vault phase (auto-resolving anyone undecided) and starts the next round.
func (m *Match) CloseVault() error {
	if m.Phase != PhaseVault {
		return ErrWrongPhase
	}
	for _, p := range m.P {
		m.AutoVault(p)
	}
	m.startRound()
	return nil
}

// BothReady reports whether both players marked ready for the current phase.
func (m *Match) BothReady() bool { return m.P[0].Ready && m.P[1].Ready }

// CheckPool verifies card conservation (pool + shops + owned copies == configured pool).
func (m *Match) CheckPool() bool {
	total := append([]int{}, m.Pool...)
	red := append([]int{}, m.PoolRed...)
	for _, p := range m.P {
		for i, l := range p.Shop {
			if l >= 0 {
				total[l]++
				if i < len(p.ShopColor) && p.ShopColor[i] == Red {
					red[l]++
				}
			}
		}
		for _, cs := range [][]Card{p.Board, p.Bench} {
			for _, c := range cs {
				total[c.Line] += m.copiesOf(c.Star)
				red[c.Line] += c.Red
			}
		}
	}
	for l, def := range m.cfg.Lines {
		if total[l] != def.Pool || red[l] != def.Pool/2 || m.PoolRed[l] < 0 || m.PoolRed[l] > m.Pool[l] {
			return false
		}
	}
	return true
}
