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
