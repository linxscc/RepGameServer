package rules

import (
	"errors"
	"fmt"
	"math/rand"
)

// Errors returned by player actions. Callers map them to protocol error codes.
var (
	ErrWrongPhase     = errors.New("rules: action not allowed in this phase")
	ErrNotEnoughGold  = errors.New("rules: not enough gold")
	ErrBenchFull      = errors.New("rules: bench is full")
	ErrRelicSlotsFull = errors.New("rules: relic slots are full")
	ErrNoSuchItem     = errors.New("rules: no such card, relic or slot")
	ErrMaxLevel       = errors.New("rules: already at max level")
	ErrBadLayout      = errors.New("rules: invalid board layout")
	ErrBadChoice      = errors.New("rules: invalid choice")
	ErrAlreadyChosen  = errors.New("rules: already chosen")
)

type Phase int

const (
	PhaseFamily Phase = iota // 选家世
	PhasePrep                // 准备阶段
	PhaseVault               // 宝库三选一
	PhaseOver                // 对局结束
	PhaseOmen                // 天象掷骰（天象回合开始时，准备阶段之前）
)

// OmenChoice is a player's secret roll/skip decision in the omen phase.
type OmenChoice int

const (
	OmenUndecided OmenChoice = 0
	OmenRolls     OmenChoice = 1
	OmenSkips     OmenChoice = 2
)

// ActiveOmen is one omen that appeared; it is in effect for rounds From..Until.
type ActiveOmen struct {
	Omen, From, Until int // Omen indexes Config.Omens
}

// OmenRoll records the last omen dice.
type OmenRoll struct {
	Round           int
	Dice            [2]int // per seat; 0 = did not roll
	Forced, Shifted bool   // Forced: both skipped so the system rolled; Shifted: duplicate moved the point
	Point, Omen     int    // final point; Omen indexes Config.Omens (-1 = none yet)
}

// Color is a card's suit color.
type Color int

const (
	Red   Color = 0
	Black Color = 1
)

// Card is one card instance owned by a player.
type Card struct {
	UID   string
	Line  int // index into Config.Lines
	Star  int // 1..MaxStar
	Bonus int // 永久点数加成
	Color Color
	Red   int // red 1★ copies contained in this card (1★: 1 for red, 0 for black)
}

// RelicInst is one relic a player owns.
type RelicInst struct {
	Relic  int     // index into Config.Relics
	Value  float64 // 成长值
	Choice int     // 选定的行当（门派令 / 家传绝学），-1 表示无
}

// Player is one side of a match. Fields are read-only for callers; mutate through Match methods.
type Player struct {
	ID     string
	HP     int
	Gold   int
	Level  int
	XP     int
	Board  []Card
	Bench  []Card
	Relics []RelicInst
	Family int // -1 until chosen

	WinStreak, LossStreak int
	WonLast, LostLast     bool

	Shop       []int   // line index per slot, -1 = empty; each shown card is reserved from the pool
	ShopColor  []Color // same length as Shop; meaningless where Shop[i] < 0
	ShopRelic  int     // relic index, -1 = none
	Locked     bool
	LockStreak int

	LastBoard     []Card // board at the last settlement; drives ops bonds
	Refreshed     bool
	FreeRefreshes int
	Ready         bool

	FamilyOffer []int // family indexes offered at start
	VaultOffer  []int // relic indexes offered in the vault
	OmenChoice  OmenChoice

	triggers map[string]int
	three    []bool // line → owns a max-star card
}

// Match is a deterministic 1v1 game. It is not safe for concurrent use.
type Match struct {
	cfg     *Config
	rng     *rand.Rand
	Round   int
	Phase   Phase
	Pool    []int // remaining copies per line (red + black)
	PoolRed []int // remaining red copies per line; black = Pool[l] - PoolRed[l]
	P       [2]*Player
	Winner  int // 0/1, -1 = draw (valid once Phase == PhaseOver)
	// FirstDealt is the seat that drew its opening cards first (for balance analysis).
	FirstDealt int
	// Surrendered is set when the match ended because a player gave up.
	Surrendered bool
	// Omens is every omen that appeared, in order (history is kept).
	Omens []ActiveOmen
	// LastRoll is the most recent omen roll (Omen == -1 before the first).
	LastRoll OmenRoll

	uidSeq int
}

func (m *Match) Config() *Config { return m.cfg }

// ActiveOmens returns the omens in effect this round, in order of appearance.
func (m *Match) ActiveOmens() []ActiveOmen {
	var out []ActiveOmen
	for _, o := range m.Omens {
		if o.From <= m.Round && m.Round <= o.Until {
			out = append(out, o)
		}
	}
	return out
}

// OmenActive reports whether the omen with this key is in effect this round.
func (m *Match) OmenActive(key string) bool {
	i, ok := m.cfg.OmenIndex(key)
	if !ok {
		return false
	}
	return m.omenActiveIdx(i)
}

// omenSum adds a param over the active omens with this key plus any in extra
// (omen indexes treated as active in addition, e.g. hypothetical scoring).
func (m *Match) omenSum(key, param string, extra []int) float64 {
	t := 0.0
	for _, i := range m.omenList(extra) {
		if o := m.cfg.Omens[i]; o.Key == key {
			t += o.Params[param]
		}
	}
	return t
}

// omenList is the active omens (in order of appearance) followed by extra.
func (m *Match) omenList(extra []int) []int {
	var out []int
	for _, o := range m.ActiveOmens() {
		out = append(out, o.Omen)
	}
	return append(out, extra...)
}

// BenchSize is the bench capacity this round (地动 shrinks it), at least 1.
func (m *Match) BenchSize(p *Player) int {
	n := m.cfg.Rules.Board.BenchSize + int(m.omenSum(OmenQuake, "bench", nil))
	if n < 1 {
		n = 1
	}
	return n
}

// ShopSize is the number of shop card slots a roll deals this round (断货), at least 1.
// A locked shop keeps its slots until the next roll.
func (m *Match) ShopSize() int {
	n := m.cfg.Rules.Board.ShopSize + int(m.omenSum(OmenShortage, "shop", nil))
	if n < 1 {
		n = 1
	}
	return n
}

// RefreshCost is what a paid shop refresh costs this round (闭市).
func (m *Match) RefreshCost(p *Player) int {
	n := m.cfg.Rules.Economy.RefreshCost + int(m.omenSum(OmenMarketClosed, "refresh", nil))
	if n < 0 {
		n = 0
	}
	return n
}

// BondAtDelta is how much every bond tier's required count is raised this round (失和).
func (m *Match) BondAtDelta() int { return int(m.omenSum(OmenDiscord, "at", nil)) }

func (m *Match) omenActiveIdx(i int) bool {
	for _, o := range m.ActiveOmens() {
		if o.Omen == i {
			return true
		}
	}
	return false
}

func (m *Match) newUID() string {
	m.uidSeq++
	return fmt.Sprintf("c%d", m.uidSeq)
}

// Player returns the player with the given ID, or nil.
func (m *Match) Player(id string) *Player {
	for _, p := range m.P {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Opponent returns the other player.
func (m *Match) Opponent(p *Player) *Player {
	if m.P[0] == p {
		return m.P[1]
	}
	return m.P[0]
}

// ---------- card values ----------

func (m *Match) copiesOf(star int) int {
	n := 1
	for s := 1; s < star; s++ {
		n *= m.cfg.Rules.Merge.GroupSize
	}
	return n
}

func (m *Match) starFactor(star int) int {
	n := 1
	for s := 1; s < star; s++ {
		n *= m.cfg.Rules.Merge.StarMultiplier
	}
	return n
}

// Points is a card's own points (star points + permanent bonus).
func (m *Match) Points(c Card) int {
	return m.cfg.Lines[c.Line].Points*m.starFactor(c.Star) + c.Bonus
}

// SellPrice is what selling a card returns.
func (m *Match) SellPrice(c Card) int {
	cost := m.cfg.Lines[c.Line].Cost
	if c.Star == StarOne {
		return cost
	}
	return m.copiesOf(c.Star)*cost - m.cfg.Rules.Economy.MergedSellLoss
}

func (m *Match) CardName(c Card) string { return m.cfg.Lines[c.Line].Card }

// Suit is one of the four card suits: Kind*2 + Color (god=0, ghost=1).
type Suit int

const (
	SuitRedGod     Suit = 0
	SuitBlackGod   Suit = 1
	SuitRedGhost   Suit = 2
	SuitBlackGhost Suit = 3
)

// IsGod reports whether the card's line is a god (神) line.
func (m *Match) IsGod(c Card) bool { return m.cfg.Lines[c.Line].Kind == KindGod }

// SuitOf returns the card's suit.
func (m *Match) SuitOf(c Card) Suit {
	k := 1
	if m.IsGod(c) {
		k = 0
	}
	return Suit(k*2 + int(c.Color))
}

// ---------- family helpers ----------

func (m *Match) familyParam(p *Player, key, param string) (float64, bool) {
	if p.Family < 0 {
		return 0, false
	}
	f := m.cfg.Families[p.Family]
	if f.Key != key {
		return 0, false
	}
	return f.Params[param], true
}

// ---------- legend helpers ----------

// legendOn reports whether the board holds a max-star card whose skill key matches, returning its params.
func (m *Match) legendsOn(board []Card) []Skill {
	var out []Skill
	seen := map[int]bool{}
	for _, c := range board {
		if c.Star == m.cfg.Rules.Merge.MaxStar && !seen[c.Line] {
			seen[c.Line] = true
			out = append(out, m.cfg.Lines[c.Line].Legend)
		}
	}
	return out
}

func (m *Match) legendSum(board []Card, key, param string) (float64, bool) {
	total, found := 0.0, false
	for _, s := range m.legendsOn(board) {
		if s.Key == key {
			total += s.Params[param]
			found = true
		}
	}
	return total, found
}

// ---------- relics: copy resolution ----------

// effRelic is a relic as it acts after copy resolution.
type effRelic struct {
	RelicInst
	slot   int  // position in the relic bar
	copied bool // came from 铜镜 / 拓本 (does not grow)
	empty  bool // a copy with nothing valid to copy
}

func (m *Match) behaviour(relic int) relicBehaviour {
	return relicBehaviours[m.cfg.Relics[relic].Key]
}

func (m *Match) effective(relics []RelicInst) []effRelic {
	out := make([]effRelic, len(relics))
	isCopy := func(r RelicInst) bool { return m.behaviour(r.Relic).kind == kindCopy }
	for i, r := range relics {
		out[i] = effRelic{RelicInst: r, slot: i}
		switch m.cfg.Relics[r.Relic].Key {
		case RelicMirror:
			if i+1 < len(relics) && !isCopy(relics[i+1]) {
				out[i] = effRelic{RelicInst: relics[i+1], slot: i, copied: true}
			} else {
				out[i].empty = true
			}
		case RelicRubbing:
			if i > 0 && !isCopy(relics[0]) {
				out[i] = effRelic{RelicInst: relics[0], slot: i, copied: true}
			} else {
				out[i].empty = true
			}
		}
	}
	return out
}

func (m *Match) relicKey(r effRelic) string { return m.cfg.Relics[r.Relic].Key }

// sumEffParam adds a param over every effective (non-empty) relic with the given key.
func (m *Match) sumEffParam(eff []effRelic, key, param string) float64 {
	t := 0.0
	for _, r := range eff {
		if !r.empty && m.relicKey(r) == key {
			t += m.cfg.Relics[r.Relic].Params[param]
		}
	}
	return t
}

// hasEff reports whether any effective (non-empty) relic has the given key.
func (m *Match) hasEff(eff []effRelic, key string) bool {
	for _, r := range eff {
		if !r.empty && m.relicKey(r) == key {
			return true
		}
	}
	return false
}

// ---------- bonds ----------

// BondState is one bond's count and active tier (-1 = inactive).
type BondState struct {
	Count int
	Tier  int
}

type BondCounts struct {
	Trade  []BondState
	Origin []BondState
}

func activeTier(tiers []Tier, count, delta int) int {
	t := -1
	for i, tr := range tiers {
		if count >= tr.At+delta {
			t = i
		}
	}
	return t
}

// bondCounts counts the board's bonds; delta raises every tier's required count (失和).
func (m *Match) bondCounts(board []Card, relics []RelicInst, delta int) BondCounts {
	cfg := m.cfg
	bc := BondCounts{Trade: make([]BondState, len(cfg.Trades)), Origin: make([]BondState, len(cfg.Origins))}
	seen := map[int]bool{}
	for _, c := range board {
		if seen[c.Line] {
			continue
		}
		seen[c.Line] = true
		bc.Trade[cfg.TradeOf(c.Line)].Count++
		bc.Origin[cfg.OriginOf(c.Line)].Count++
	}
	for _, s := range m.legendsOnWithLine(board) {
		if s.skill.Key == LegendOwnOriginCount {
			bc.Origin[cfg.OriginOf(s.line)].Count += int(s.skill.Params["count"])
		}
	}
	for _, r := range m.effective(relics) {
		if !r.empty && m.relicKey(r) == RelicSectToken && r.Choice >= 0 && bc.Trade[r.Choice].Count > 0 {
			bc.Trade[r.Choice].Count += int(cfg.Relics[r.Relic].Params["count"])
		}
	}
	for i := range bc.Trade {
		bc.Trade[i].Tier = activeTier(cfg.Trades[i].Tiers, bc.Trade[i].Count, delta)
	}
	for i := range bc.Origin {
		bc.Origin[i].Tier = activeTier(cfg.Origins[i].Tiers, bc.Origin[i].Count, delta)
	}
	return bc
}

type lineSkill struct {
	line  int
	skill Skill
}

func (m *Match) legendsOnWithLine(board []Card) []lineSkill {
	var out []lineSkill
	seen := map[int]bool{}
	for _, c := range board {
		if c.Star == m.cfg.Rules.Merge.MaxStar && !seen[c.Line] {
			seen[c.Line] = true
			out = append(out, lineSkill{c.Line, m.cfg.Lines[c.Line].Legend})
		}
	}
	return out
}

// ActiveBonds counts bonds with any active tier.
func (bc BondCounts) ActiveBonds() int {
	n := 0
	for _, s := range append(append([]BondState{}, bc.Trade...), bc.Origin...) {
		if s.Tier >= 0 {
			n++
		}
	}
	return n
}

// TierCount counts active tiers; reaching a bond's top tier counts as 2 (tier index + 1).
func (bc BondCounts) TierCount() int {
	n := 0
	for _, s := range append(append([]BondState{}, bc.Trade...), bc.Origin...) {
		n += s.Tier + 1
	}
	return n
}

// modifier sums a generic bond modifier over all active tiers.
func (m *Match) modifier(bc BondCounts, name string) float64 {
	t := 0.0
	for i, s := range bc.Trade {
		if s.Tier >= 0 {
			t += m.cfg.Trades[i].Tiers[s.Tier].Values[name]
		}
	}
	for i, s := range bc.Origin {
		if s.Tier >= 0 {
			t += m.cfg.Origins[i].Tiers[s.Tier].Values[name]
		}
	}
	return t
}

// timesModifier multiplies every active times_mult value.
func (m *Match) timesModifiers(bc BondCounts) []struct {
	name   string
	factor float64
} {
	var out []struct {
		name   string
		factor float64
	}
	for i, s := range bc.Trade {
		if s.Tier >= 0 {
			if f, ok := m.cfg.Trades[i].Tiers[s.Tier].Values[BondTimesMult]; ok {
				out = append(out, struct {
					name   string
					factor float64
				}{m.cfg.Trades[i].Name, f})
			}
		}
	}
	for i, s := range bc.Origin {
		if s.Tier >= 0 {
			if f, ok := m.cfg.Origins[i].Tiers[s.Tier].Values[BondTimesMult]; ok {
				out = append(out, struct {
					name   string
					factor float64
				}{m.cfg.Origins[i].Name, f})
			}
		}
	}
	return out
}

// OpsBonds are the bond counts that drive economy effects: the last settled board.
func (m *Match) OpsBonds(p *Player) BondCounts {
	return m.bondCounts(p.LastBoard, p.Relics, m.BondAtDelta())
}

// Bonds are the bond counts of the current board (what the next settlement scores).
func (m *Match) Bonds(p *Player) BondCounts { return m.bondCounts(p.Board, p.Relics, m.BondAtDelta()) }

// ---------- slots, interest ----------

func (m *Match) Slots(p *Player) int {
	n := m.cfg.Level(p.Level).Slots + int(m.sumEffParam(m.effective(p.Relics), RelicExpansion, "slots"))
	if v, ok := m.legendSum(p.LastBoard, LegendExtraSlots, "slots"); ok {
		n += int(v)
	}
	if v, ok := m.familyParam(p, FamilyHermit, "board_slots"); ok {
		n += int(v)
	}
	n += int(m.omenSum(OmenNarrowPath, "slots", nil))
	if n < m.cfg.Rules.Board.MinSlots {
		n = m.cfg.Rules.Board.MinSlots
	}
	return n
}

func (m *Match) RelicSlots(p *Player) int {
	n := m.cfg.Rules.Board.RelicSlots
	if v, ok := m.familyParam(p, FamilyHermit, "relic_slots"); ok {
		n += int(v)
	}
	return n
}

// Interest is what the player's current gold would earn at the next round start.
// 苛税 (heavy_tax) makes it 0.
func (m *Match) Interest(p *Player) int {
	if m.OmenActive(OmenHeavyTax) {
		return 0
	}
	e := m.cfg.Rules.Economy
	rate, capV := 1.0, float64(e.InterestCap)
	for _, s := range m.legendsOn(p.LastBoard) {
		if s.Key == LegendInterest {
			rate, capV = s.Params["rate"], s.Params["cap"]
		}
	}
	capV += m.sumEffParam(m.effective(p.Relics), RelicTreasure, "cap")
	if v, ok := m.familyParam(p, FamilyMiser, "interest_cap"); ok {
		capV += v
	}
	v := float64(p.Gold/e.InterestStep) * rate
	if v > capV {
		v = capV
	}
	return int(v)
}

func (m *Match) freeRefreshTotal(p *Player) int {
	return int(m.sumEffParam(m.effective(p.Relics), RelicMarketPass, "free"))
}

func (m *Match) FreeRefreshesLeft(p *Player) int {
	n := m.freeRefreshTotal(p) - p.FreeRefreshes
	if n < 0 {
		return 0
	}
	return n
}

func (m *Match) trigger(p *Player, key string) bool {
	if p.triggers[key] >= m.cfg.Rules.Merge.TriggerCap {
		return false
	}
	p.triggers[key]++
	return true
}

// CountCopiesColor is CountCopies restricted to one color.
func (m *Match) CountCopiesColor(p *Player, line, star int, c Color) int {
	n := 0
	for _, cs := range [][]Card{p.Board, p.Bench} {
		for _, cd := range cs {
			if cd.Line == line && cd.Star == star && cd.Color == c {
				n++
			}
		}
	}
	return n
}

func (m *Match) CountCopies(p *Player, line, star int) int {
	n := 0
	for _, cs := range [][]Card{p.Board, p.Bench} {
		for _, c := range cs {
			if c.Line == line && c.Star == star {
				n++
			}
		}
	}
	return n
}

// XPToNext is the experience needed for the next level (0 at max level).
func (m *Match) XPToNext(p *Player) int { return m.cfg.Level(p.Level).XPToNext }

// RelicNeedsChoice reports whether acquiring the relic requires choosing a trade.
func (m *Match) RelicNeedsChoice(relic int) bool { return m.behaviour(relic).choice }

// HasThreeStar reports whether the player owns a max-star card of the line.
func (m *Match) HasThreeStar(p *Player, line int) bool { return p.three[line] }

func (m *Match) addXP(p *Player, n int) {
	for n > 0 && p.Level < m.cfg.MaxLevel() {
		need := m.cfg.Level(p.Level).XPToNext - p.XP
		if n < need {
			p.XP += n
			return
		}
		n -= need
		p.XP = 0
		p.Level++
	}
	if p.Level >= m.cfg.MaxLevel() {
		p.XP = 0
	}
}
