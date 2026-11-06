// Package rules implements the RepGame card-battle rules (shop, merge, bonds,
// relics, scoring, battle) as a pure, deterministic engine with no networking.
// Every number comes from Config, which is loaded from config/tables/chess_*.yaml.
package rules

import (
	"errors"
	"fmt"
	"sort"
)

// ---------- chess_rules.yaml ----------

type Tier struct {
	At     int                `yaml:"at"`     // 达到几个不同的系
	Desc   string             `yaml:"desc"`   // 给玩家看的效果说明
	Values map[string]float64 `yaml:"values"` // 通用修正项，见 bondModifiers
}

type LevelDef struct {
	Level     int   `yaml:"level"`
	XPToNext  int   `yaml:"xp_to_next"` // 升到下一级所需经验；满级为 0
	Slots     int   `yaml:"slots"`      // 出阵格
	ShopOdds  []int `yaml:"shop_odds"`  // 各费用档概率（%），下标 = 费用 - 1
	RelicOdds []int `yaml:"relic_odds"` // 商店法宝各稀有度概率（%），下标见 rarities
}

type StreakGold struct {
	At   int `yaml:"at"`
	Gold int `yaml:"gold"`
}

type StageDamage struct {
	UntilRound int `yaml:"until_round"`
	Damage     int `yaml:"damage"`
}

type Suppression struct {
	Ratio  float64 `yaml:"ratio"`  // 胜方得分 ÷ 败方得分 ≥ ratio
	Factor float64 `yaml:"factor"` // 扣血系数
}

type VaultOdds struct {
	FromRound int   `yaml:"from_round"`
	Odds      []int `yaml:"odds"`
}

type RarityDef struct {
	Key   string `yaml:"key"`
	Name  string `yaml:"name"`
	Price int    `yaml:"price"`
}

type RulesTable struct {
	Player struct {
		StartHP       int `yaml:"start_hp"`
		StartGold     int `yaml:"start_gold"`
		StartCards    int `yaml:"start_cards"`     // 开局发的不同 1★ 卡数量
		StartCardCost int `yaml:"start_card_cost"` // 开局卡的费用档
		FamilyOffer   int `yaml:"family_offer"`    // 开局家世几选一
	} `yaml:"player"`
	Economy struct {
		BaseIncome      int          `yaml:"base_income"`
		InterestStep    int          `yaml:"interest_step"`
		InterestCap     int          `yaml:"interest_cap"`
		WinBonus        int          `yaml:"win_bonus"`
		StreakGold      []StreakGold `yaml:"streak_gold"`
		RefreshCost     int          `yaml:"refresh_cost"`
		XPCost          int          `yaml:"xp_cost"`
		XPPerBuy        int          `yaml:"xp_per_buy"`
		AutoXP          int          `yaml:"auto_xp"`
		AutoXPFromRound int          `yaml:"auto_xp_from_round"`
		MergedSellLoss  int          `yaml:"merged_sell_loss"` // 合成卡卖出时少返还的金币
	} `yaml:"economy"`
	Levels []LevelDef `yaml:"levels"`
	Board  struct {
		BenchSize     int  `yaml:"bench_size"`
		RelicSlots    int  `yaml:"relic_slots"`
		ShopSize      int  `yaml:"shop_size"`
		MaxLockRounds int  `yaml:"max_lock_rounds"`
		MinSlots      int  `yaml:"min_slots"`
		ShopRelic     bool `yaml:"shop_relic"` // 商店是否出售法宝（关闭时法宝只能从宝库获得）
	} `yaml:"board"`
	Merge struct {
		GroupSize      int `yaml:"group_size"`
		MaxStar        int `yaml:"max_star"`
		TriggerCap     int `yaml:"trigger_cap"` // “每次 X 时”奖励每回合每来源上限
		StarMultiplier int `yaml:"star_multiplier"`
	} `yaml:"merge"`
	Battle struct {
		Stages         []StageDamage `yaml:"stages"`
		LateDamageStep int           `yaml:"late_damage_step"`
		MaxRounds      int           `yaml:"max_rounds"`
		Suppression    []Suppression `yaml:"suppression"`
		ScoreCap       float64       `yaml:"score_cap"`
		BondTierMult   float64       `yaml:"bond_tier_mult"` // 每个生效羁绊档位 +倍率
	} `yaml:"battle"`
	Vault struct {
		Rounds   []int       `yaml:"rounds"`
		Offer    int         `yaml:"offer"`
		SkipGold int         `yaml:"skip_gold"`
		Odds     []VaultOdds `yaml:"odds"`
	} `yaml:"vault"`
	Rarities []RarityDef `yaml:"rarities"`
	Timers   struct {
		FamilySeconds int `yaml:"family_seconds"`
		PrepSeconds   int `yaml:"prep_seconds"`
		SettleSeconds int `yaml:"settle_seconds"`
		VaultSeconds  int `yaml:"vault_seconds"`
	} `yaml:"timers"`
}

// ---------- chess_lines.yaml ----------

type Skill struct {
	Key    string             `yaml:"key"`
	Desc   string             `yaml:"desc"`
	Params map[string]float64 `yaml:"params"`
}

type LineDef struct {
	Key    string `yaml:"key"`
	Name   string `yaml:"name"`
	Card   string `yaml:"card"` // 卡名（各星级相同），与客户端卡面同名
	Cost   int    `yaml:"cost"`
	Points int    `yaml:"points"` // 1★ 点数
	Pool   int    `yaml:"pool"`
	Trade  string `yaml:"trade"`
	Origin string `yaml:"origin"`
	Legend Skill  `yaml:"legend"` // 3★ 技能
}

type LinesFile struct {
	Lines []LineDef `yaml:"lines"`
}

// ---------- chess_bonds.yaml ----------

type BondDef struct {
	Key   string `yaml:"key"`
	Name  string `yaml:"name"`
	Desc  string `yaml:"desc"`
	Tiers []Tier `yaml:"tiers"`
}

type BondsFile struct {
	Trades  []BondDef `yaml:"trades"`
	Origins []BondDef `yaml:"origins"`
}

// ---------- chess_relics.yaml ----------

type RelicDef struct {
	Key    string             `yaml:"key"`
	Name   string             `yaml:"name"`
	Rarity string             `yaml:"rarity"`
	Desc   string             `yaml:"desc"`
	Params map[string]float64 `yaml:"params"`
}

type RelicsFile struct {
	Relics []RelicDef `yaml:"relics"`
}

// ---------- chess_families.yaml ----------

type FamilyDef struct {
	Key    string             `yaml:"key"`
	Name   string             `yaml:"name"`
	Desc   string             `yaml:"desc"`
	Params map[string]float64 `yaml:"params"`
}

type FamiliesFile struct {
	Families []FamilyDef `yaml:"families"`
}

// ---------- assembled config ----------

// Config is the full, validated rule set. Build it with NewConfig.
type Config struct {
	Rules    RulesTable
	Lines    []LineDef
	Trades   []BondDef
	Origins  []BondDef
	Relics   []RelicDef
	Families []FamilyDef

	tradeIdx  map[string]int
	originIdx map[string]int
	lineIdx   map[string]int
	relicIdx  map[string]int
	rarityIdx map[string]int
	familyIdx map[string]int
	vaultSet  map[int]bool
}

// NewConfig assembles and validates the tables.
func NewConfig(r RulesTable, l LinesFile, b BondsFile, rel RelicsFile, f FamiliesFile) (*Config, error) {
	c := &Config{Rules: r, Lines: l.Lines, Trades: b.Trades, Origins: b.Origins, Relics: rel.Relics, Families: f.Families}
	if err := c.index(); err != nil {
		return nil, err
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func indexKeys[T any](items []T, key func(T) string, what string, errs *[]error) map[string]int {
	m := make(map[string]int, len(items))
	for i, it := range items {
		k := key(it)
		if k == "" {
			*errs = append(*errs, fmt.Errorf("%s #%d: empty key", what, i))
			continue
		}
		if _, dup := m[k]; dup {
			*errs = append(*errs, fmt.Errorf("%s: duplicate key %q", what, k))
		}
		m[k] = i
	}
	return m
}

func (c *Config) index() error {
	var errs []error
	c.tradeIdx = indexKeys(c.Trades, func(b BondDef) string { return b.Key }, "chess_bonds.trades", &errs)
	c.originIdx = indexKeys(c.Origins, func(b BondDef) string { return b.Key }, "chess_bonds.origins", &errs)
	c.lineIdx = indexKeys(c.Lines, func(l LineDef) string { return l.Key }, "chess_lines", &errs)
	c.relicIdx = indexKeys(c.Relics, func(r RelicDef) string { return r.Key }, "chess_relics", &errs)
	c.rarityIdx = indexKeys(c.Rules.Rarities, func(r RarityDef) string { return r.Key }, "chess_rules.rarities", &errs)
	c.familyIdx = indexKeys(c.Families, func(f FamilyDef) string { return f.Key }, "chess_families", &errs)
	c.vaultSet = map[int]bool{}
	for _, r := range c.Rules.Vault.Rounds {
		c.vaultSet[r] = true
	}
	return errors.Join(errs...)
}

func (c *Config) validate() error {
	var errs []error
	check := func(ok bool, format string, a ...any) {
		if !ok {
			errs = append(errs, fmt.Errorf(format, a...))
		}
	}
	r := c.Rules
	nr := len(r.Rarities)
	check(nr > 0, "chess_rules.rarities must not be empty")
	check(r.Player.StartHP > 0, "chess_rules.player.start_hp must be > 0")
	check(r.Player.StartCards >= 0, "chess_rules.player.start_cards must be >= 0")
	check(r.Player.FamilyOffer >= 1 && r.Player.FamilyOffer <= len(c.Families), "chess_rules.player.family_offer must be 1..%d", len(c.Families))
	check(r.Economy.InterestStep > 0, "chess_rules.economy.interest_step must be > 0")
	check(r.Economy.RefreshCost >= 0 && r.Economy.XPCost > 0 && r.Economy.XPPerBuy > 0, "chess_rules.economy costs are invalid")
	check(r.Economy.MergedSellLoss >= 1, "chess_rules.economy.merged_sell_loss must be >= 1 (blocks the buy-merge-sell loop)")
	check(len(r.Levels) > 0, "chess_rules.levels must not be empty")
	maxCost := 0
	for _, l := range c.Lines {
		if l.Cost > maxCost {
			maxCost = l.Cost
		}
	}
	for i, lv := range r.Levels {
		check(lv.Level == i+1, "chess_rules.levels[%d].level must be %d", i, i+1)
		check(lv.Slots >= 1, "chess_rules.levels[%d].slots must be >= 1", i)
		check(len(lv.ShopOdds) == maxCost, "chess_rules.levels[%d].shop_odds must have %d entries (one per cost)", i, maxCost)
		check(len(lv.RelicOdds) == nr, "chess_rules.levels[%d].relic_odds must have %d entries (one per rarity)", i, nr)
		check(sum(lv.ShopOdds) > 0, "chess_rules.levels[%d].shop_odds must not be all zero", i)
		last := i == len(r.Levels)-1
		check(last == (lv.XPToNext == 0), "chess_rules.levels[%d].xp_to_next must be 0 only on the last level", i)
	}
	b := r.Board
	check(b.BenchSize >= 1 && b.RelicSlots >= 1 && b.ShopSize >= 1, "chess_rules.board sizes must be >= 1")
	check(b.MaxLockRounds >= 0 && b.MinSlots >= 1, "chess_rules.board lock/min slots invalid")
	m := r.Merge
	check(m.GroupSize >= 2, "chess_rules.merge.group_size must be >= 2")
	check(m.MaxStar >= 2, "chess_rules.merge.max_star must be >= 2")
	check(m.TriggerCap >= 1, "chess_rules.merge.trigger_cap must be >= 1")
	check(m.StarMultiplier >= 1, "chess_rules.merge.star_multiplier must be >= 1")
	bt := r.Battle
	check(len(bt.Stages) > 0, "chess_rules.battle.stages must not be empty")
	for i := 1; i < len(bt.Stages); i++ {
		check(bt.Stages[i].UntilRound > bt.Stages[i-1].UntilRound, "chess_rules.battle.stages must be sorted by until_round")
	}
	check(bt.MaxRounds > 0, "chess_rules.battle.max_rounds must be > 0")
	check(bt.ScoreCap > 0, "chess_rules.battle.score_cap must be > 0")
	for i := 1; i < len(bt.Suppression); i++ {
		check(bt.Suppression[i].Ratio > bt.Suppression[i-1].Ratio, "chess_rules.battle.suppression must be sorted by ratio")
	}
	check(r.Vault.Offer >= 1, "chess_rules.vault.offer must be >= 1")
	check(len(r.Vault.Odds) > 0 && r.Vault.Odds[0].FromRound <= 1, "chess_rules.vault.odds must start at round 1")
	for i, v := range r.Vault.Odds {
		check(len(v.Odds) == nr, "chess_rules.vault.odds[%d] must have %d entries", i, nr)
		check(i == 0 || v.FromRound > r.Vault.Odds[i-1].FromRound, "chess_rules.vault.odds must be sorted by from_round")
	}
	for _, rr := range r.Rarities {
		check(rr.Price > 0, "chess_rules.rarities %s price must be > 0", rr.Key)
	}
	t := r.Timers
	check(t.FamilySeconds > 0 && t.PrepSeconds >= 5 && t.SettleSeconds >= 0 && t.VaultSeconds > 0, "chess_rules.timers are invalid")

	// lines
	check(len(c.Lines) > 0, "chess_lines must not be empty")
	for _, l := range c.Lines {
		check(l.Card != "", "chess_lines %s needs a card name", l.Key)
		check(l.Cost >= 1 && l.Cost <= maxCost, "chess_lines %s cost invalid", l.Key)
		check(l.Points > 0 && l.Pool > 0, "chess_lines %s points/pool must be > 0", l.Key)
		_, okT := c.tradeIdx[l.Trade]
		_, okO := c.originIdx[l.Origin]
		check(okT, "chess_lines %s: unknown trade %q", l.Key, l.Trade)
		check(okO, "chess_lines %s: unknown origin %q", l.Key, l.Origin)
		if err := validateParams("legend", l.Legend.Key, l.Legend.Params, legendParams); err != nil {
			errs = append(errs, fmt.Errorf("chess_lines %s: %w", l.Key, err))
		}
	}
	// bonds
	for _, group := range [][]BondDef{c.Trades, c.Origins} {
		for _, bd := range group {
			check(len(bd.Tiers) > 0, "chess_bonds %s must have tiers", bd.Key)
			for i, tr := range bd.Tiers {
				check(tr.At >= 1, "chess_bonds %s tier %d: at must be >= 1", bd.Key, i)
				check(tr.Desc != "", "chess_bonds %s tier %d: desc must not be empty", bd.Key, i)
				check(i == 0 || tr.At > bd.Tiers[i-1].At, "chess_bonds %s tiers must be sorted by at", bd.Key)
				for k := range tr.Values {
					_, ok := bondModifiers[k]
					check(ok, "chess_bonds %s tier %d: unknown modifier %q", bd.Key, i, k)
				}
			}
		}
	}
	// relics
	check(len(c.Relics) > 0, "chess_relics must not be empty")
	for _, rl := range c.Relics {
		_, ok := c.rarityIdx[rl.Rarity]
		check(ok, "chess_relics %s: unknown rarity %q", rl.Key, rl.Rarity)
		if err := validateParams("relic", rl.Key, rl.Params, relicParams); err != nil {
			errs = append(errs, fmt.Errorf("chess_relics: %w", err))
		}
		switch rl.Key {
		case "literary_pen":
			_, ok := c.tradeIdx["letter"]
			check(ok, "chess_relics literary_pen needs a trade with key \"letter\"")
		case "piggy":
			check(rl.Params["per_gold"] > 0, "chess_relics piggy.per_gold must be > 0")
		}
		for _, k := range []string{"factor"} {
			if v, ok := rl.Params[k]; ok {
				check(v > 0, "chess_relics %s.%s must be > 0", rl.Key, k)
			}
		}
	}
	// families
	for _, f := range c.Families {
		if err := validateParams("family", f.Key, f.Params, familyParams); err != nil {
			errs = append(errs, fmt.Errorf("chess_families: %w", err))
		}
	}
	return errors.Join(errs...)
}

func sum(xs []int) int {
	s := 0
	for _, x := range xs {
		s += x
	}
	return s
}

// validateParams checks a behaviour key is known and has exactly the params it needs.
func validateParams(kind, key string, params map[string]float64, registry map[string][]string) error {
	want, ok := registry[key]
	if !ok {
		return fmt.Errorf("unknown %s key %q", kind, key)
	}
	var errs []error
	for _, p := range want {
		if _, ok := params[p]; !ok {
			errs = append(errs, fmt.Errorf("%s %q: missing param %q", kind, key, p))
		}
	}
	allowed := map[string]bool{}
	for _, p := range want {
		allowed[p] = true
	}
	extra := []string{}
	for p := range params {
		if !allowed[p] {
			extra = append(extra, p)
		}
	}
	sort.Strings(extra)
	for _, p := range extra {
		errs = append(errs, fmt.Errorf("%s %q: unknown param %q", kind, key, p))
	}
	return errors.Join(errs...)
}

// ---------- lookups ----------

func (c *Config) Level(level int) LevelDef {
	if level < 1 {
		level = 1
	}
	if level > len(c.Rules.Levels) {
		level = len(c.Rules.Levels)
	}
	return c.Rules.Levels[level-1]
}

func (c *Config) MaxLevel() int { return len(c.Rules.Levels) }

func (c *Config) LineIndex(key string) (int, bool) { i, ok := c.lineIdx[key]; return i, ok }

func (c *Config) RelicIndex(key string) (int, bool) { i, ok := c.relicIdx[key]; return i, ok }

func (c *Config) FamilyIndex(key string) (int, bool) { i, ok := c.familyIdx[key]; return i, ok }

func (c *Config) RarityOf(relic int) int { return c.rarityIdx[c.Relics[relic].Rarity] }

func (c *Config) RelicPrice(relic int) int { return c.Rules.Rarities[c.RarityOf(relic)].Price }

func (c *Config) IsVaultRound(round int) bool { return c.vaultSet[round] }

func (c *Config) TradeOf(line int) int { return c.tradeIdx[c.Lines[line].Trade] }

func (c *Config) OriginOf(line int) int { return c.originIdx[c.Lines[line].Origin] }

func (c *Config) VaultOddsFor(round int) []int {
	odds := c.Rules.Vault.Odds[0].Odds
	for _, v := range c.Rules.Vault.Odds {
		if round >= v.FromRound {
			odds = v.Odds
		}
	}
	return odds
}

func (c *Config) StageDamage(round int) int {
	st := c.Rules.Battle.Stages
	for _, s := range st {
		if round <= s.UntilRound {
			return s.Damage
		}
	}
	last := st[len(st)-1]
	return last.Damage + c.Rules.Battle.LateDamageStep*(round-last.UntilRound)
}

func (c *Config) SuppressionFactor(ratio float64) float64 {
	f := 1.0
	for _, s := range c.Rules.Battle.Suppression {
		if ratio >= s.Ratio {
			f = s.Factor
		}
	}
	return f
}

func (c *Config) StreakGold(n int) int {
	g := 0
	for _, s := range c.Rules.Economy.StreakGold {
		if n >= s.At {
			g = s.Gold
		}
	}
	return g
}
