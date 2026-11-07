package chessbot

import (
	"math/rand"
	"testing"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/rules"
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
