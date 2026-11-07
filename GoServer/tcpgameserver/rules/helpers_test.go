package rules

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// loadTestConfig reads the real tables from config/tables (the package cannot import
// config here without an import cycle, so it decodes the same files directly).
func loadTestConfig(t *testing.T) *Config {
	t.Helper()
	dir := filepath.Join("..", "config", "tables")
	var (
		r  RulesTable
		l  LinesFile
		b  BondsFile
		rl RelicsFile
		f  FamiliesFile
	)
	for name, dst := range map[string]any{
		"chess_rules.yaml": &r, "chess_lines.yaml": &l, "chess_bonds.yaml": &b,
		"chess_relics.yaml": &rl, "chess_families.yaml": &f,
	} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal(data, dst); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	cfg, err := NewConfig(r, l, b, rl, f)
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
