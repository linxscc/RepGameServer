package rules

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// testTables holds the raw decoded tables so tests can tweak one and rebuild.
type testTables struct {
	R   RulesTable
	L   LinesFile
	B   BondsFile
	Rel RelicsFile
	F   FamiliesFile
	O   OmensFile
}

// loadTestTables reads the real tables from data/config/tables (the package cannot import
// config here without an import cycle, so it decodes the same files directly).
func loadTestTables(t *testing.T) testTables {
	t.Helper()
	dir := filepath.Join("..", "..", "data", "config", "tables")
	var tt testTables
	for name, dst := range map[string]any{
		"chess_rules.yaml": &tt.R, "chess_lines.yaml": &tt.L, "chess_bonds.yaml": &tt.B,
		"chess_relics.yaml": &tt.Rel, "chess_families.yaml": &tt.F, "chess_omens.yaml": &tt.O,
	} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal(data, dst); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	return tt
}

func (tt testTables) build() (*Config, error) {
	return NewConfig(tt.R, tt.L, tt.B, tt.Rel, tt.F, tt.O)
}

func loadTestConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := loadTestTables(t).build()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// newPrepMatch returns a match in round 1 prep with neutral families and empty boards
// (start cards are returned to the pool so card conservation still holds).
func newPrepMatch(t *testing.T, seed int64) *Match {
	t.Helper()
	cfg := loadTestConfig(t)
	m := NewMatch(cfg, rand.New(rand.NewSource(seed)), [2]string{"a", "b"})
	for _, p := range m.P {
		p.FamilyOffer = []int{mustFamily(t, cfg, "noble")}
		if err := m.PickFamily(p, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Begin(); err != nil {
		t.Fatal(err)
	}
	for _, p := range m.P {
		for _, c := range append(append([]Card{}, p.Board...), p.Bench...) {
			m.Pool[c.Line] += m.copiesOf(c.Star)
			m.PoolRed[c.Line] += c.Red
		}
		p.Board, p.Bench = nil, nil
		p.Gold = 100
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved after setup")
	}
	return m
}

func mustLine(t *testing.T, cfg *Config, key string) int {
	t.Helper()
	i, ok := cfg.LineIndex(key)
	if !ok {
		t.Fatalf("no line %q", key)
	}
	return i
}

func mustRelic(t *testing.T, cfg *Config, key string) int {
	t.Helper()
	i, ok := cfg.RelicIndex(key)
	if !ok {
		t.Fatalf("no relic %q", key)
	}
	return i
}

func mustFamily(t *testing.T, cfg *Config, key string) int {
	t.Helper()
	i, ok := cfg.FamilyIndex(key)
	if !ok {
		t.Fatalf("no family %q", key)
	}
	return i
}

// forceShop replaces the player's shop with the given lines, all Red (Black once red is exhausted),
// keeping pool accounting exact.
func forceShop(t *testing.T, m *Match, p *Player, keys ...string) {
	t.Helper()
	m.returnShop(p)
	p.Shop, p.ShopColor = nil, nil
	for _, k := range keys {
		l := mustLine(t, m.cfg, k)
		if m.Pool[l] == 0 {
			t.Fatalf("pool empty for %s", k)
		}
		c := Red
		if m.PoolRed[l] == 0 {
			c = Black
		}
		takeFromPool(m, l, c)
		p.Shop = append(p.Shop, l)
		p.ShopColor = append(p.ShopColor, c)
	}
}

// forceShopColored is forceShop with explicit slot colors.
func forceShopColored(t *testing.T, m *Match, p *Player, keys []string, colors []Color) {
	t.Helper()
	m.returnShop(p)
	p.Shop, p.ShopColor = nil, nil
	for i, k := range keys {
		l := mustLine(t, m.cfg, k)
		if m.Pool[l] == 0 || (colors[i] == Red && m.PoolRed[l] == 0) || (colors[i] == Black && m.Pool[l]-m.PoolRed[l] == 0) {
			t.Fatalf("pool empty for %s color %d", k, colors[i])
		}
		takeFromPool(m, l, colors[i])
		p.Shop = append(p.Shop, l)
		p.ShopColor = append(p.ShopColor, colors[i])
	}
}

func takeFromPool(m *Match, l int, c Color) {
	m.Pool[l]--
	if c == Red {
		m.PoolRed[l]--
	}
}

// coloredCard builds a card of the given color without touching the pool.
func coloredCard(m *Match, t *testing.T, key string, star int, c Color) Card {
	cd := Card{UID: m.newUID(), Line: mustLine(t, m.cfg, key), Star: star, Color: c}
	if c == Red {
		cd.Red = m.copiesOf(star)
	}
	return cd
}

func card(m *Match, t *testing.T, key string, star int) Card {
	return Card{UID: m.newUID(), Line: mustLine(t, m.cfg, key), Star: star}
}

func relic(t *testing.T, m *Match, key string, value float64) RelicInst {
	return RelicInst{Relic: mustRelic(t, m.cfg, key), Value: value, Choice: -1}
}

// stepsBy returns the score steps whose source is src.
func stepsBy(r ScoreResult, src string) []Step {
	var out []Step
	for _, s := range r.Steps {
		if s.Source == src {
			out = append(out, s)
		}
	}
	return out
}

// advance leaves the current phase: Prep → Settle, Vault → CloseVault, Omen → ResolveOmen.
func advance(t *testing.T, m *Match) {
	t.Helper()
	var err error
	switch m.Phase {
	case PhasePrep:
		_, err = m.Settle()
	case PhaseVault:
		err = m.CloseVault()
	case PhaseOmen:
		err = m.ResolveOmen()
	default:
		t.Fatalf("advance: cannot leave phase %d", m.Phase)
	}
	if err != nil {
		t.Fatal(err)
	}
}

// forceOmen makes the omen active from the current round for the configured duration.
func forceOmen(t *testing.T, m *Match, key string) {
	t.Helper()
	i, ok := m.cfg.OmenIndex(key)
	if !ok {
		t.Fatalf("no omen %q", key)
	}
	m.Omens = append(m.Omens, ActiveOmen{Omen: i, From: m.Round, Until: m.Round + m.cfg.Rules.Omen.Duration - 1})
}
