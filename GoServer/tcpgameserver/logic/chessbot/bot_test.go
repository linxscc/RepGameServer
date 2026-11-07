package chessbot

import (
	"math/rand"
	"testing"

	"GoServer/tcpgameserver/data/config"
	"GoServer/tcpgameserver/logic/rules"
)

// Two bots play full matches through the engine; every match must end, keep the
// pool conserved and never leave a board larger than its slots.
func TestBotsFinishMatches(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	rounds := 0
	for game := 0; game < 40; game++ {
		rng := rand.New(rand.NewSource(int64(game)))
		m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(int64(game)+100)), [2]string{"a", "b"})
		bots := [2]*Bot{NewRandom(m, rng), NewRandom(m, rng)}
		for i, p := range m.P {
			bots[i].PickFamily(m, p)
		}
		if err := m.Begin(); err != nil {
			t.Fatal(err)
		}
		for m.Phase != rules.PhaseOver {
			switch m.Phase {
			case rules.PhasePrep:
				for i, p := range m.P {
					bots[i].Prep(m, p)
					if len(p.Board) > m.Slots(p) || !p.Ready {
						t.Fatalf("game %d: bot left an illegal or unready state", game)
					}
				}
				if _, err := m.Settle(); err != nil {
					t.Fatal(err)
				}
			case rules.PhaseOmen:
				for i, p := range m.P {
					bots[i].ChooseOmen(m, p)
				}
				for oi, op := range m.P {
					bots[oi].ChooseOmen(m, op)
				}
				_ = m.ResolveOmen()
			case rules.PhaseVault:
				for i, p := range m.P {
					bots[i].PickVault(m, p)
				}
				if err := m.CloseVault(); err != nil {
					t.Fatal(err)
				}
			}
			if !m.CheckPool() {
				t.Fatalf("game %d round %d: pool not conserved", game, m.Round)
			}
		}
		rounds += m.Round
	}
	if avg := float64(rounds) / 40; avg < 12 || avg > 28 {
		t.Fatalf("average match length %.1f rounds is outside the expected range", avg)
	}
}

// A bot that prefers red must field clearly more same-color cards by round 15
// than a bot that ignores color.
func TestColorBotFieldsMostlyItsColor(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	base := Archetypes()[5] // 均衡
	redArch, plainArch := base, base
	redArch.Color = "red"
	plainArch.Color = ""
	plainArch.IgnoreSuits = true
	var redSum, plainSum float64
	samples, skipped := 0, 0
	for seed := int64(0); seed < 16 && samples < 10; seed++ {
		rng := rand.New(rand.NewSource(seed))
		m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(seed+100)), [2]string{"a", "b"})
		bots := [2]*Bot{New(m, redArch, rng), New(m, plainArch, rng)}
		for i, p := range m.P {
			bots[i].PickFamily(m, p)
		}
		if err := m.Begin(); err != nil {
			t.Fatal(err)
		}
		got := false
		var peaks [2]int
		for m.Phase != rules.PhaseOver {
			switch m.Phase {
			case rules.PhasePrep:
				for i, p := range m.P {
					bots[i].Prep(m, p)
				}
				res, err := m.Settle()
				if err != nil {
					t.Fatal(err)
				}
				if res.Round == 15 {
					got = true
					for i, p := range m.P {
						peaks[i] = ColorPeak(m, p)
					}
				}
			case rules.PhaseOmen:
				for i, p := range m.P {
					bots[i].ChooseOmen(m, p)
				}
				for oi, op := range m.P {
					bots[oi].ChooseOmen(m, op)
				}
				_ = m.ResolveOmen()
			case rules.PhaseVault:
				for i, p := range m.P {
					bots[i].PickVault(m, p)
				}
				if err := m.CloseVault(); err != nil {
					t.Fatal(err)
				}
			}
		}
		if !got {
			skipped++
			continue
		}
		samples++
		redSum += float64(peaks[0])
		plainSum += float64(peaks[1])
	}
	if samples < 8 {
		t.Fatalf("only %d samples reached round 15 (%d skipped)", samples, skipped)
	}
	ra, pa := redSum/float64(samples), plainSum/float64(samples)
	t.Logf("samples=%d skipped=%d red-bot ColorPeak avg=%.2f plain-bot avg=%.2f", samples, skipped, ra, pa)
	if ra < pa+1 {
		t.Fatalf("red bot avg ColorPeak %.2f not >= plain %.2f + 1", ra, pa)
	}
}

func TestArchetypeRelicKeysExist(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range Archetypes() {
		for k := range a.Relics {
			if _, ok := tables.Chess.RelicIndex(k); !ok {
				t.Errorf("archetype %s: unknown relic %q", a.Name, k)
			}
		}
	}
}

// ---------- suit / counter relic valuation ----------

func suitTestMatch(t *testing.T) (*rules.Match, *rules.Player, *rules.Player) {
	t.Helper()
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(7)), [2]string{"a", "b"})
	return m, m.P[0], m.P[1]
}

func lineOfKind(m *rules.Match, kind string) int {
	for i, l := range m.Config().Lines {
		if l.Kind == kind {
			return i
		}
	}
	return 0
}

func relicIdx(t *testing.T, m *rules.Match, key string) int {
	t.Helper()
	i, ok := m.Config().RelicIndex(key)
	if !ok {
		t.Fatalf("unknown relic %s", key)
	}
	return i
}

func TestSynergyFollowsHoldings(t *testing.T) {
	m, p, o := suitTestMatch(t)
	ghost := lineOfKind(m, rules.KindGhost)
	for i := 0; i < 4; i++ {
		p.Board = append(p.Board, rules.Card{UID: "r", Line: ghost, Star: 1, Color: rules.Red})
	}
	for i := 0; i < 3; i++ {
		o.LastBoard = append(o.LastBoard, rules.Card{UID: "g", Line: ghost, Star: 1, Color: rules.Black})
	}
	cin, pine := relicIdx(t, m, "cinnabar"), relicIdx(t, m, "pine_ink")
	if a, b := synergy(m, p, cin), synergy(m, p, pine); !(a > b) {
		t.Fatalf("cinnabar %.2f should beat pine_ink %.2f", a, b)
	}
	dm, gb := relicIdx(t, m, "demon_mirror"), relicIdx(t, m, "god_breaker")
	if a, b := synergy(m, p, dm), synergy(m, p, gb); !(a > b) {
		t.Fatalf("demon_mirror %.2f should beat god_breaker %.2f", a, b)
	}
}

func TestPrefColorFromSuitRelic(t *testing.T) {
	m, p, _ := suitTestMatch(t)
	b := New(m, Archetypes()[5], rand.New(rand.NewSource(1)))
	if _, ok := b.prefColor(m, p); ok {
		t.Fatal("no suit relic: expected no color preference")
	}
	p.Relics = append(p.Relics, rules.RelicInst{Relic: relicIdx(t, m, "pine_ink"), Choice: -1})
	if c, ok := b.prefColor(m, p); !ok || c != rules.Black {
		t.Fatalf("got (%v,%v), want (Black,true)", c, ok)
	}
}

func TestBaselineIgnoresSuits(t *testing.T) {
	m, p, _ := suitTestMatch(t)
	ghost := lineOfKind(m, rules.KindGhost)
	for i := 0; i < 4; i++ {
		p.Board = append(p.Board, rules.Card{UID: "r", Line: ghost, Star: 1, Color: rules.Red})
	}
	arch := SuitArchetypes()[0]
	b := New(m, arch, rand.New(rand.NewSource(1)))
	if v := b.relicValue(m, p, relicIdx(t, m, "cinnabar")); v != 0.1 {
		t.Fatalf("baseline cinnabar value %v, want 0.1", v)
	}
	arch.IgnoreSuits = false
	b2 := New(m, arch, rand.New(rand.NewSource(1)))
	if v := b2.relicValue(m, p, relicIdx(t, m, "cinnabar")); v <= 0.1 {
		t.Fatalf("non-baseline cinnabar value %v should exceed 0.1", v)
	}
}

func TestSuitArchetypesFinishMatches(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	archs := SuitArchetypes()
	if len(archs) != 7 {
		t.Fatalf("want 7 suit archetypes, got %d", len(archs))
	}
	for _, a := range archs {
		for k := range a.Relics {
			if _, ok := tables.Chess.RelicIndex(k); !ok {
				t.Errorf("archetype %s: unknown relic %q", a.Name, k)
			}
		}
	}
	seed := int64(1)
	for i := range archs {
		for j := i + 1; j < len(archs); j++ {
			seed++
			rng := rand.New(rand.NewSource(seed))
			m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(seed+100)), [2]string{"a", "b"})
			bots := [2]*Bot{New(m, archs[i], rng), New(m, archs[j], rng)}
			for k, p := range m.P {
				bots[k].PickFamily(m, p)
			}
			if err := m.Begin(); err != nil {
				t.Fatal(err)
			}
			for m.Phase != rules.PhaseOver {
				switch m.Phase {
				case rules.PhasePrep:
					for k, p := range m.P {
						bots[k].Prep(m, p)
					}
					if _, err := m.Settle(); err != nil {
						t.Fatal(err)
					}
				case rules.PhaseOmen:
					for i, p := range m.P {
						bots[i].ChooseOmen(m, p)
					}
					for oi, op := range m.P {
						bots[oi].ChooseOmen(m, op)
					}
					_ = m.ResolveOmen()
				case rules.PhaseVault:
					for k, p := range m.P {
						bots[k].PickVault(m, p)
					}
					if err := m.CloseVault(); err != nil {
						t.Fatal(err)
					}
				}
				if !m.CheckPool() {
					t.Fatalf("%s vs %s round %d: pool not conserved", archs[i].Name, archs[j].Name, m.Round)
				}
			}
		}
	}
}

func TestOmenPickTruthTable(t *testing.T) {
	cases := []struct {
		myOne, myTwo, oppOne, oppTwo float64
		want                         bool
	}{
		{0.3, 0.1, 0.2, 0.1, true},
		{0.1, 0.3, 0.2, 0.1, false},
		{0.3, 0.1, 0.1, 0.2, false},
		{0.1, 0.3, 0.1, 0.2, true},
		{0.2, 0.2, 0.1, 0.2, false},
	}
	for i, c := range cases {
		if got := omenPick(c.myOne, c.myTwo, c.oppOne, c.oppTwo); got != c.want {
			t.Fatalf("case %d got %v", i, got)
		}
	}
}

// A board of five red 1-star cards fears red_moon (point 4): 1/6 in 1d6, 3/36 in 2d6.
func TestRedBoardFearsSmallSums(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(1)), [2]string{"a", "b"})
	p := m.P[0]
	p.Board = nil
	for i := 0; i < 5; i++ {
		p.Board = append(p.Board, rules.Card{UID: string(rune('a' + i)), Line: 0, Star: 1, Color: rules.Red, Red: 1})
	}
	p.Bench = nil
	one, two := omenLosses(m, p, p.Board)
	if !(one > two) {
		t.Fatalf("one=%v two=%v, want one > two", one, two)
	}
}

func TestBotsFinishMatchesWithOmens(t *testing.T) {
	tables, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	sawThree := false
	for game := 0; game < 10; game++ {
		rng := rand.New(rand.NewSource(int64(game)))
		m := rules.NewMatch(tables.Chess, rand.New(rand.NewSource(int64(game)+100)), [2]string{"a", "b"})
		bots := [2]*Bot{NewRandom(m, rng), NewRandom(m, rng)}
		for i, p := range m.P {
			bots[i].PickFamily(m, p)
		}
		if err := m.Begin(); err != nil {
			t.Fatal(err)
		}
		for m.Phase != rules.PhaseOver {
			switch m.Phase {
			case rules.PhasePrep:
				for i, p := range m.P {
					bots[i].Prep(m, p)
				}
				if _, err := m.Settle(); err != nil {
					t.Fatal(err)
				}
			case rules.PhaseOmen:
				for i, p := range m.P {
					bots[i].ChooseOmen(m, p)
				}
				if err := m.ResolveOmen(); err != nil {
					t.Fatal(err)
				}
			case rules.PhaseVault:
				for i, p := range m.P {
					bots[i].PickVault(m, p)
				}
				if err := m.CloseVault(); err != nil {
					t.Fatal(err)
				}
			}
			if !m.CheckPool() {
				t.Fatalf("game %d round %d: pool not conserved", game, m.Round)
			}
		}
		if len(m.Omens) >= 3 {
			sawThree = true
		}
	}
	if !sawThree {
		t.Fatal("no match reached 3 omens")
	}
}
