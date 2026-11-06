package rules

import (
	"math"
	"math/rand"
)

// NewMatch creates a match in the family-pick phase. rng makes it deterministic.
func NewMatch(cfg *Config, rng *rand.Rand, ids [2]string) *Match {
	m := &Match{cfg: cfg, rng: rng, Phase: PhaseFamily, Winner: -1, Pool: make([]int, len(cfg.Lines))}
	for i, l := range cfg.Lines {
		m.Pool[i] = l.Pool
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
	for _, i := range m.rng.Perm(2) {
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
		m.Pool[l] -= need
		c := Card{UID: m.newUID(), Line: l, Star: star}
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

// startRound pays income, grants xp, refreshes shops and enters the prep phase.
func (m *Match) startRound() {
	m.Round++
	m.Phase = PhasePrep
	e := m.cfg.Rules.Economy
	for _, p := range m.P {
		p.triggers = map[string]int{}
		p.Refreshed = false
		p.FreeRefreshes = 0
		p.Ready = false
		p.VaultOffer = nil

		income := m.Interest(p) + e.BaseIncome
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
		p.Gold += income

		if m.Round >= e.AutoXPFromRound {
			m.addXP(p, e.AutoXP+int(m.modifier(m.OpsBonds(p), "xp_per_round")))
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

// fitBoard moves cards that no longer fit onto the bench (or sells them if the bench is full).
func (m *Match) fitBoard(p *Player) {
	for len(p.Board) > m.Slots(p) {
		c := p.Board[len(p.Board)-1]
		p.Board = p.Board[:len(p.Board)-1]
		if len(p.Bench) < m.cfg.Rules.Board.BenchSize {
			p.Bench = append(p.Bench, c)
		} else {
			m.disposeCard(p, c)
		}
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

// ---------- vault ----------

func (m *Match) openVault() {
	m.Phase = PhaseVault
	for _, p := range m.P {
		n := m.cfg.Rules.Vault.Offer
		if v, ok := m.familyParam(p, "noble", "extra_offer"); ok {
			n += int(v)
		}
		ex := map[int]bool{}
		p.VaultOffer = nil
		for len(p.VaultOffer) < n {
			id := m.randomRelic(p, m.cfg.VaultOddsFor(m.Round), ex)
			if id < 0 {
				break
			}
			ex[id] = true
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
	for _, p := range m.P {
		for _, l := range p.Shop {
			if l >= 0 {
				total[l]++
			}
		}
		for _, cs := range [][]Card{p.Board, p.Bench} {
			for _, c := range cs {
				total[c.Line] += m.copiesOf(c.Star)
			}
		}
	}
	for l, def := range m.cfg.Lines {
		if total[l] != def.Pool {
			return false
		}
	}
	return true
}
