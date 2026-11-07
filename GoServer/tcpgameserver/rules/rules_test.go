package rules

import (
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// The design doc's worked example: 统领 + 游商 + 官员 with 先锋旗/精兵简政/独行侠/乘风 at level 6.
// pts 45×3+9+15 = 159; mult 1 → +2 (市朝 3 = 2 tiers) → ×2 (市朝) → ×2 (统领) = 12
// → +3×5 empty = 27 → ×2.5 = 67.5 → ×1.9 = 128.25; 159 × 128.25 = 20391.
func TestScoreDocExample(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 6
	board := []Card{card(m, t, "soldier", 3), card(m, t, "trader", 2), card(m, t, "clerk", 2)}
	relics := []RelicInst{relic(t, m, "vanguard", 0), relic(t, m, "lean", 0), relic(t, m, "lone_wolf", 0), relic(t, m, "tailwind", 0)}
	if got := m.ScoreWith(p, board, nil, relics).Score; got != 20391 {
		t.Fatalf("score = %v, want 20391", got)
	}
	if got := m.ScoreWith(p, board, nil, nil).Score; got != 828 {
		t.Fatalf("no-relic score = %v, want 828", got)
	}
}

func TestScoreStepsReplayToTotal(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 6
	board := []Card{card(m, t, "soldier", 3), card(m, t, "trader", 2), card(m, t, "clerk", 2)}
	res := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "vanguard", 0), relic(t, m, "tailwind", 0)})
	last := res.Steps[len(res.Steps)-1]
	if last.Points != res.Points || last.Mult != res.Mult {
		t.Fatalf("last step %+v does not match result %v × %v", last, res.Points, res.Mult)
	}
	if res.Steps[0].Phase != 1 || res.Steps[0].CardUID != board[0].UID {
		t.Fatalf("first step should be the leftmost card: %+v", res.Steps[0])
	}
}

func TestRelicOrderMatters(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 4                                                        // 乘风 ×1.6
	board := []Card{card(m, t, "scholar", 1), card(m, t, "artist", 1)} // pts 5; mult 1 + 1 tier + 2 文墨 = 4
	a := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "furnace", 4), relic(t, m, "tailwind", 0)}).Score
	b := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "tailwind", 0), relic(t, m, "furnace", 4)}).Score
	if a != 64 || b != 52 {
		t.Fatalf("got %v / %v, want 64 / 52", a, b)
	}
}

func TestRetriggerSkipsTimesRelics(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 4
	// 天师 3★（taoist）：最左边法宝额外触发 1 次；×气势类法宝只触发一次
	board := []Card{card(m, t, "taoist", 3), card(m, t, "monk", 1)}
	r := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "tailwind", 0)})
	if n := len(stepsBy(r, "乘风")); n != 1 {
		t.Fatalf("乘风 steps = %d, want 1", n)
	}
	r = m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "furnace", 4)})
	if n := len(stepsBy(r, "熔炉")); n != 2 {
		t.Fatalf("熔炉 steps = %d, want 2", n)
	}
	board[0] = card(m, t, "taoist", 1)
	r = m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "furnace", 4)})
	if n := len(stepsBy(r, "熔炉")); n != 1 {
		t.Fatalf("no retrigger: 熔炉 steps = %d, want 1", n)
	}
}

func TestCopyCannotCopyCopy(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	board := []Card{card(m, t, "carpenter", 1)}
	base := m.ScoreWith(p, board, nil, nil).Score
	if got := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "mirror", 0), relic(t, m, "rubbing", 0)}).Score; got != base {
		t.Fatalf("mirror+rubbing = %v, want %v (both empty)", got, base)
	}
	eff := m.effective([]RelicInst{relic(t, m, "tailwind", 0), relic(t, m, "rubbing", 0), relic(t, m, "mirror", 0)})
	if m.relicKey(eff[1]) != "tailwind" || !eff[2].empty {
		t.Fatalf("unexpected copy resolution: %+v", eff)
	}
}

func TestMergeSellLoopLosesGoldAndIsCapped(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Relics = []RelicInst{relic(t, m, "furnace", 0), relic(t, m, "accountant", 0)}
	for loop := 0; loop < 5; loop++ {
		forceShop(t, m, p, "carpenter", "carpenter", "carpenter")
		for i := 0; i < 3; i++ {
			if _, err := m.BuyCard(p, i); err != nil {
				t.Fatal(err)
			}
		}
		if len(p.Bench) != 1 || p.Bench[0].Star != 2 {
			t.Fatalf("want one 2★, bench=%+v", p.Bench)
		}
		if err := m.SellCard(p, p.Bench[0].UID); err != nil {
			t.Fatal(err)
		}
	}
	if p.Gold != 95 {
		t.Fatalf("gold = %d, want 95", p.Gold)
	}
	if p.Relics[0].Value != 6 || p.Relics[1].Value != 6 {
		t.Fatalf("growth must cap at 3 hits × 2: %+v", p.Relics)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestChainMergeKeepsBoardPosition(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Board = []Card{card(m, t, "scholar", 1), card(m, t, "soldier", 2)}
	p.Bench = []Card{card(m, t, "soldier", 2), card(m, t, "soldier", 1), card(m, t, "soldier", 1)}
	soldier := mustLine(t, m.cfg, "soldier")
	scholar := mustLine(t, m.cfg, "scholar")
	// 8 张拷贝里黑色只有 6 张（12 张红黑各半），所以让一张 2★ 含 2 张红色拷贝
	p.Board[1].Color, p.Board[1].Red = Red, 2
	m.Pool[soldier] -= 8
	m.PoolRed[soldier] -= 2
	m.Pool[scholar]--
	forceShop(t, m, p, "soldier")
	made, err := m.BuyCard(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(made) != 2 || len(p.Board) != 2 || p.Board[1].Line != soldier || p.Board[1].Star != 3 || len(p.Bench) != 0 {
		t.Fatalf("made=%d board=%+v bench=%+v", len(made), p.Board, p.Bench)
	}
	if !p.three[soldier] || !m.CheckPool() {
		t.Fatal("three-star flag or pool wrong")
	}
	for _, l := range p.Shop {
		if l == soldier {
			t.Fatal("a line the player has at 3★ must not appear in the shop")
		}
	}
	m.rollShop(p)
	for _, l := range p.Shop {
		if l == soldier {
			t.Fatal("rolled a 3★ line into the shop")
		}
	}
}

func TestFullBenchBuyOnlyWhenMerging(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	for i := 0; i < m.cfg.Rules.Board.BenchSize; i++ {
		p.Bench = append(p.Bench, card(m, t, "farmer", 1))
	}
	p.Bench[0] = card(m, t, "monk", 1)
	p.Bench[1] = card(m, t, "monk", 1)
	m.Pool[mustLine(t, m.cfg, "farmer")] -= 6
	m.Pool[mustLine(t, m.cfg, "monk")] -= 2
	forceShop(t, m, p, "taoist", "monk")
	if err := m.CanBuyCard(p, 0); err != ErrBenchFull {
		t.Fatalf("want ErrBenchFull, got %v", err)
	}
	if err := m.CanBuyCard(p, 1); err != nil {
		t.Fatalf("merging buy should be allowed: %v", err)
	}
}

func TestThreeStarSkills(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 5
	p.Board = []Card{card(m, t, "doctor", 3)}
	if d := m.damageTo(p, 9); d != 7 {
		t.Fatalf("医圣 damage = %d, want 7", d)
	}
	bc := m.bondCounts([]Card{card(m, t, "farmer", 3), card(m, t, "monk", 1), card(m, t, "soldier", 1)}, nil, 0)
	if bc.Origin[m.cfg.originIdx["rural"]].Count != 4 {
		t.Fatalf("神农 rural count = %d, want 4", bc.Origin[m.cfg.originIdx["rural"]].Count)
	}
	if got := m.ScoreWith(p, []Card{card(m, t, "scholar", 3)}, nil, nil).Score; got != 18*11 {
		t.Fatalf("教授 = %v, want 198", got)
	}
	// 铸剑师 grows on later merges, 3 times per round at most
	q := m.P[1]
	smith := card(m, t, "smith", 3)
	q.Board = []Card{smith}
	m.Pool[smith.Line] -= 9
	q.three[smith.Line] = true
	for loop := 0; loop < 4; loop++ {
		forceShop(t, m, q, "farmer", "farmer", "farmer")
		for i := 0; i < 3; i++ {
			if _, err := m.BuyCard(q, i); err != nil {
				t.Fatal(err)
			}
		}
		if err := m.SellCard(q, q.Bench[len(q.Bench)-1].UID); err != nil {
			t.Fatal(err)
		}
	}
	if q.Board[0].Bonus != 18 {
		t.Fatalf("铸剑师 bonus = %d, want 18", q.Board[0].Bonus)
	}
}

func TestShopLockLimit(t *testing.T) {
	m := newPrepMatch(t, 3)
	p := m.P[0]
	first := append([]int{}, p.Shop...)
	next := func() {
		advance(t, m)
		for m.Phase != PhasePrep {
			advance(t, m)
		}
	}
	for r := 0; r < m.cfg.Rules.Board.MaxLockRounds; r++ {
		if err := m.SetLock(p, true); err != nil {
			t.Fatal(err)
		}
		next()
		for i := range first {
			if p.Shop[i] != first[i] {
				t.Fatalf("locked shop changed after lock %d", r+1)
			}
		}
	}
	_ = m.SetLock(p, true)
	next()
	if p.LockStreak != 0 {
		t.Fatal("lock beyond the limit must force a refresh")
	}
}

func TestSettleDamageAndTie(t *testing.T) {
	m := newPrepMatch(t, 5)
	a, b := m.P[0], m.P[1]
	a.Board = []Card{card(m, t, "soldier", 1)}   // 5
	b.Board = []Card{card(m, t, "carpenter", 1)} // 2 → ratio 2.5 → ×1.25
	m.Pool[a.Board[0].Line]--
	m.Pool[b.Board[0].Line]--
	res, err := m.Settle()
	if err != nil {
		t.Fatal(err)
	}
	if res.Winner != 0 || res.Damage[1] != 3 || res.Damage[0] != 0 { // ceil(2 × 1.25)
		t.Fatalf("unexpected result %+v", res)
	}
	if a.WinStreak != 1 || b.LossStreak != 1 {
		t.Fatal("streaks not updated")
	}
}

// TestRandomPlayKeepsInvariants plays many games with random legal and illegal actions
// and checks conservation and limits after every step.
func TestRandomPlayKeepsInvariants(t *testing.T) {
	cfg := loadTestConfig(t)
	for game := 0; game < 300; game++ {
		rng := rand.New(rand.NewSource(int64(game)))
		m := NewMatch(cfg, rand.New(rand.NewSource(int64(game)*7+1)), [2]string{"a", "b"})
		for _, p := range m.P {
			if rng.Intn(3) > 0 {
				_ = m.PickFamily(p, rng.Intn(len(p.FamilyOffer)))
			}
		}
		if err := m.Begin(); err != nil {
			t.Fatal(err)
		}
		check := func(where string) {
			t.Helper()
			if !m.CheckPool() {
				t.Fatalf("game %d %s: pool not conserved", game, where)
			}
			for _, p := range m.P {
				if len(p.Bench) > cfg.Rules.Board.BenchSize || len(p.Relics) > m.RelicSlots(p) || p.Gold < 0 {
					t.Fatalf("game %d %s: illegal state bench=%d relics=%d gold=%d", game, where, len(p.Bench), len(p.Relics), p.Gold)
				}
				if m.Phase == PhasePrep && len(p.Board) > m.Slots(p) {
					t.Fatalf("game %d %s: board %d > slots %d", game, where, len(p.Board), m.Slots(p))
				}
			}
		}
		for m.Phase != PhaseOver {
			if m.Round > cfg.Rules.Battle.MaxRounds {
				t.Fatalf("game %d ran past max rounds", game)
			}
			switch m.Phase {
			case PhasePrep:
				for _, p := range m.P {
					for k := 0; k < 25; k++ {
						randomAction(m, p, rng)
						check("prep")
					}
				}
				if _, err := m.Settle(); err != nil {
					t.Fatal(err)
				}
			case PhaseVault:
				for _, p := range m.P {
					switch rng.Intn(3) {
					case 0:
						_ = m.SkipVault(p)
					case 1:
						if len(p.VaultOffer) > 0 {
							_ = m.PickVault(p, rng.Intn(len(p.VaultOffer)), rng.Intn(len(p.Relics)+1)-1, rng.Intn(len(cfg.Trades)))
						}
					}
				}
				if err := m.CloseVault(); err != nil {
					t.Fatal(err)
				}
			case PhaseOmen:
				for _, p := range m.P {
					if k := rng.Intn(3); k > 0 {
						_ = m.ChooseOmen(p, k == 1)
					}
				}
				if err := m.ResolveOmen(); err != nil {
					t.Fatal(err)
				}
				if m.LastRoll.Point < 1 || m.LastRoll.Point > cfg.MaxOmenPoint() || m.LastRoll.Omen < 0 {
					t.Fatalf("game %d: bad omen roll %+v", game, m.LastRoll)
				}
			}
			check("round")
		}
	}
}

func randomAction(m *Match, p *Player, rng *rand.Rand) {
	all := append(append([]Card{}, p.Board...), p.Bench...)
	switch rng.Intn(9) {
	case 0, 1, 2:
		_, _ = m.BuyCard(p, rng.Intn(len(p.Shop)+1)-1)
	case 3:
		if len(all) > 0 {
			_ = m.SellCard(p, all[rng.Intn(len(all))].UID)
		}
	case 4:
		_ = m.Refresh(p)
	case 5:
		_ = m.BuyXP(p)
	case 6:
		if rng.Intn(2) == 0 {
			_ = m.BuyRelic(p, rng.Intn(len(m.cfg.Trades)))
		} else if len(p.Relics) > 0 {
			_ = m.SellRelic(p, rng.Intn(len(p.Relics)))
		}
	case 7:
		perm := rng.Perm(len(all))
		n := rng.Intn(m.Slots(p) + 2)
		if n > len(perm) {
			n = len(perm)
		}
		uids := make([]string, n)
		for i := 0; i < n; i++ {
			uids[i] = all[perm[i]].UID
		}
		_ = m.Arrange(p, uids, rng.Perm(len(p.Relics)))
	case 8:
		_ = m.SetLock(p, rng.Intn(2) == 0)
	}
}

func TestScoreStepsCarryTriggerCounts(t *testing.T) {
	m := newPrepMatch(t, 7)
	p := m.P[0]
	cfg := m.cfg

	// 家传绝学选方外：方外卡（黄巾力士）逐卡结算 2 次，步骤记 1/2、2/2
	art := relic(t, m, "secret_art", 0)
	art.Choice = cfg.TradeOf(mustLine(t, cfg, "taoist"))
	c := card(m, t, "taoist", 1)
	res := m.ScoreWith(p, []Card{c}, nil, []RelicInst{art})
	var seen []int
	for _, s := range res.Steps {
		if s.CardUID == c.UID && s.Relic < 0 {
			if s.Triggers != 2 {
				t.Fatalf("card step triggers %d want 2", s.Triggers)
			}
			seen = append(seen, s.Trigger)
		}
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("card triggers %v", seen)
	}

	// 3★ 黄巾力士（天师）：最左边的法宝人多势众额外触发 1 次 → 步骤记 1/2、2/2
	three := card(m, t, "taoist", 3)
	res = m.ScoreWith(p, []Card{three}, nil, []RelicInst{relic(t, m, "multitude", 0)})
	seen = nil
	for _, s := range res.Steps {
		if s.Relic == 0 {
			if s.Triggers != 2 {
				t.Fatalf("relic step triggers %d want 2", s.Triggers)
			}
			seen = append(seen, s.Trigger)
		}
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("relic triggers %v", seen)
	}

	// 没有重复触发时都是 1/1
	res = m.ScoreWith(p, []Card{card(m, t, "smith", 1)}, nil, nil)
	for _, s := range res.Steps {
		if s.Trigger != 1 || s.Triggers != 1 {
			t.Fatalf("plain step %+v", s)
		}
	}
}

func TestLinesHaveKindAndEvenPool(t *testing.T) {
	cfg := loadTestConfig(t)
	gods := map[int]int{}
	for _, l := range cfg.Lines {
		if l.Kind != KindGod && l.Kind != KindGhost {
			t.Fatalf("%s kind %q", l.Key, l.Kind)
		}
		if l.Pool%2 != 0 {
			t.Fatalf("%s pool %d not even", l.Key, l.Pool)
		}
		want := map[int]int{1: 16, 2: 14, 3: 12}[l.Cost]
		if l.Pool != want {
			t.Fatalf("%s pool %d want %d", l.Key, l.Pool, want)
		}
		if l.Kind == KindGod {
			gods[l.Cost]++
		}
	}
	if gods[1] != 3 || gods[2] != 2 || gods[3] != 1 {
		t.Fatalf("gods per cost %v", gods)
	}
	if cfg.Rules.Merge.SameColor {
		t.Fatal("same_color must default to false")
	}
}

func TestShopSlotsCarryColor(t *testing.T) {
	m := newPrepMatch(t, 11)
	p := m.P[0]
	for i := 0; i < 50; i++ {
		m.rollShop(p)
		if len(p.ShopColor) != len(p.Shop) {
			t.Fatalf("ShopColor len %d != Shop len %d", len(p.ShopColor), len(p.Shop))
		}
		if !m.CheckPool() {
			t.Fatalf("pool not conserved after roll %d", i)
		}
	}
}

func TestMergeTakesMajorityColor(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	soldier := mustLine(t, m.cfg, "soldier")
	p.Bench = []Card{coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "soldier", 1, Red)}
	m.Pool[soldier] -= 2
	m.PoolRed[soldier] -= 2
	forceShopColored(t, m, p, []string{"soldier"}, []Color{Black})
	if _, err := m.BuyCard(p, 0); err != nil {
		t.Fatal(err)
	}
	if len(p.Bench) != 1 {
		t.Fatalf("bench = %+v", p.Bench)
	}
	c := p.Bench[0]
	if c.Color != Red || c.Red != 2 || c.Star != 2 {
		t.Fatalf("merged = %+v", c)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestSellMergedReturnsExactColors(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	soldier := mustLine(t, m.cfg, "soldier")
	p.Bench = []Card{coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "soldier", 1, Red)}
	m.Pool[soldier] -= 2
	m.PoolRed[soldier] -= 2
	forceShopColored(t, m, p, []string{"soldier"}, []Color{Black})
	if _, err := m.BuyCard(p, 0); err != nil {
		t.Fatal(err)
	}
	pool, red := m.Pool[soldier], m.PoolRed[soldier]
	if err := m.SellCard(p, p.Bench[0].UID); err != nil {
		t.Fatal(err)
	}
	if m.Pool[soldier] != pool+3 || m.PoolRed[soldier] != red+2 {
		t.Fatalf("pool %d->%d red %d->%d", pool, m.Pool[soldier], red, m.PoolRed[soldier])
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestSameColorMergeRule(t *testing.T) {
	m := newPrepMatch(t, 1)
	m.cfg.Rules.Merge.SameColor = true
	p := m.P[0]
	soldier := mustLine(t, m.cfg, "soldier")
	p.Bench = []Card{coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "soldier", 1, Red)}
	m.Pool[soldier] -= 2
	m.PoolRed[soldier] -= 2
	forceShopColored(t, m, p, []string{"soldier", "soldier"}, []Color{Black, Red})
	if _, err := m.BuyCard(p, 0); err != nil {
		t.Fatal(err)
	}
	if len(p.Bench) != 3 {
		t.Fatalf("red,red,black must not merge: %+v", p.Bench)
	}
	if _, err := m.BuyCard(p, 1); err != nil {
		t.Fatal(err)
	}
	if len(p.Bench) != 2 {
		t.Fatalf("bench = %+v", p.Bench)
	}
	var merged *Card
	for i := range p.Bench {
		if p.Bench[i].Star == 2 {
			merged = &p.Bench[i]
		}
	}
	if merged == nil || merged.Red != 3 || merged.Color != Red {
		t.Fatalf("bench = %+v", p.Bench)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestShopDrawsOtherColorWhenOneExhausted(t *testing.T) {
	m := newPrepMatch(t, 2)
	p := m.P[0]
	m.returnShop(p)
	p.Shop, p.ShopColor = nil, nil
	// exhaust red in every line; the draw must fall back to Black without ever going negative
	for l := range m.cfg.Lines {
		m.Pool[l] -= m.PoolRed[l]
		m.PoolRed[l] = 0
	}
	for i := 0; i < 200; i++ {
		l, c := m.drawShopCard(p)
		if l < 0 {
			t.Fatal("no card drawn")
		}
		if c != Black {
			t.Fatal("drew Red from an exhausted red pool")
		}
		for k := range m.PoolRed {
			if m.PoolRed[k] < 0 {
				t.Fatalf("PoolRed[%d] = %d", k, m.PoolRed[k])
			}
		}
		m.Pool[l]++ // undo so the pool never runs dry
	}
}

func TestLockedShopKeepsColors(t *testing.T) {
	m := newPrepMatch(t, 3)
	p := m.P[0]
	firstShop := append([]int{}, p.Shop...)
	firstColor := append([]Color{}, p.ShopColor...)
	if err := m.SetLock(p, true); err != nil {
		t.Fatal(err)
	}
	advance(t, m)
	for m.Phase != PhasePrep {
		advance(t, m)
	}
	for i := range firstShop {
		if p.Shop[i] != firstShop[i] || p.ShopColor[i] != firstColor[i] {
			t.Fatalf("slot %d changed: %d/%d -> %d/%d", i, firstShop[i], firstColor[i], p.Shop[i], p.ShopColor[i])
		}
	}
}

func TestArtisanStartCardColorConsistent(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		cfg := loadTestConfig(t)
		m := NewMatch(cfg, rand.New(rand.NewSource(seed)), [2]string{"a", "b"})
		for _, p := range m.P {
			p.FamilyOffer = []int{mustFamily(t, cfg, "artisan")}
			if err := m.PickFamily(p, 0); err != nil {
				t.Fatal(err)
			}
		}
		if err := m.Begin(); err != nil {
			t.Fatal(err)
		}
		for _, p := range m.P {
			c := append(append([]Card{}, p.Board...), p.Bench...)[0]
			if c.Star != 2 || c.Red < 0 || c.Red > 3 {
				t.Fatalf("seed %d: card %+v", seed, c)
			}
			if (c.Red >= 2) != (c.Color == Red) {
				t.Fatalf("seed %d: color %d inconsistent with Red=%d", seed, c.Color, c.Red)
			}
		}
		if !m.CheckPool() {
			t.Fatalf("seed %d: pool not conserved", seed)
		}
	}
}

func TestEveryRelicHasCategory(t *testing.T) {
	cfg := loadTestConfig(t)
	want := map[string]string{"graduation": CatEconomy, "treasure": CatEconomy, "thrifty": CatEconomy, "piggy": CatEconomy,
		"furnace": CatGrowth, "accountant": CatGrowth, "grudge": CatGrowth, "mirror": CatTactic, "vanguard": CatTactic}
	for i, r := range cfg.Relics {
		if CategoryName(r.Category) == "" {
			t.Fatalf("%s category %q", r.Key, r.Category)
		}
		if w, ok := want[r.Key]; ok && cfg.CategoryOf(i) != w {
			t.Fatalf("%s category %s want %s", r.Key, r.Category, w)
		}
	}
}

func TestUnknownRelicCategoryRejected(t *testing.T) {
	tt := loadTestTables(t)
	tt.Rel.Relics = append([]RelicDef{}, tt.Rel.Relics...)
	tt.Rel.Relics[0].Category = "magic"
	_, err := tt.build()
	if err == nil || !strings.Contains(err.Error(), `unknown category "magic"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestBondModifiersNumericOnly(t *testing.T) {
	for _, k := range []string{"gold_per_merge", "retrigger_left", "free_refresh", "xp_per_round", "extra_slots", "merged_bonus_points"} {
		tt := loadTestTables(t)
		tt.B.Trades = append([]BondDef{}, tt.B.Trades...)
		craft := tt.B.Trades[0]
		craft.Tiers = append([]Tier{}, craft.Tiers...)
		craft.Tiers[0].Values = map[string]float64{k: 1}
		tt.B.Trades[0] = craft
		_, err := tt.build()
		if err == nil || !strings.Contains(err.Error(), `unknown modifier "`+k+`"`) {
			t.Fatalf("%s: err = %v", k, err)
		}
	}
}

func TestNewBondValues(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 4
	r := m.ScoreWith(p, []Card{card(m, t, "carpenter", 1), card(m, t, "smith", 1)}, nil, nil)
	found := false
	for _, s := range stepsBy(r, "天工") {
		if s.Kind == StepAddMult && s.Value == 4 {
			found = true
		}
	}
	if !found {
		t.Fatalf("天工 2: %+v", stepsBy(r, "天工"))
	}
	r = m.ScoreWith(p, []Card{card(m, t, "taoist", 1), card(m, t, "monk", 1)}, nil, nil)
	found = false
	for _, s := range stepsBy(r, "方外") {
		if s.Kind == StepPoints && s.Value == 3*2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("方外 2: %+v", stepsBy(r, "方外"))
	}
}

func TestMarketPassFreeRefresh(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Gold = 10
	if m.FreeRefreshesLeft(p) != 0 {
		t.Fatal("no free refresh without the relic")
	}
	p.Relics = []RelicInst{relic(t, m, "market_pass", 0)}
	if m.FreeRefreshesLeft(p) != 1 {
		t.Fatalf("left = %d, want 1", m.FreeRefreshesLeft(p))
	}
	if err := m.Refresh(p); err != nil || p.Gold != 10 {
		t.Fatalf("free refresh: %v gold %d", err, p.Gold)
	}
	if m.FreeRefreshesLeft(p) != 0 {
		t.Fatal("want 0 left")
	}
	if err := m.Refresh(p); err != nil || p.Gold != 10-m.cfg.Rules.Economy.RefreshCost {
		t.Fatalf("second refresh: %v gold %d", err, p.Gold)
	}
}

func totalXP(m *Match, p *Player) int {
	n := p.XP
	for lv := 1; lv < p.Level; lv++ {
		n += m.cfg.Level(lv).XPToNext
	}
	return n
}

func TestAcademyXP(t *testing.T) {
	m := newPrepMatch(t, 1)
	m.Round = m.cfg.Rules.Economy.AutoXPFromRound - 1
	m.P[0].Relics = []RelicInst{relic(t, m, "academy", 0)}
	before := [2]int{totalXP(m, m.P[0]), totalXP(m, m.P[1])}
	m.startRound()
	auto := m.cfg.Rules.Economy.AutoXP
	if d := totalXP(m, m.P[0]) - before[0]; d != auto+2 {
		t.Fatalf("academy xp gain = %d, want %d", d, auto+2)
	}
	if d := totalXP(m, m.P[1]) - before[1]; d != auto {
		t.Fatalf("plain xp gain = %d, want %d", d, auto)
	}
}

func TestExpansionSlot(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 1
	if m.Slots(p) != 3 {
		t.Fatalf("slots = %d, want 3", m.Slots(p))
	}
	p.Relics = []RelicInst{relic(t, m, "expansion", 0)}
	if m.Slots(p) != 4 {
		t.Fatalf("slots = %d, want 4", m.Slots(p))
	}
}

func TestSellExpansionFitsBoard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 1
	p.Relics = []RelicInst{relic(t, m, "expansion", 0)}
	for _, k := range []string{"carpenter", "smith", "farmer", "taoist"} {
		p.Board = append(p.Board, card(m, t, k, 1))
		m.Pool[mustLine(t, m.cfg, k)]--
	}
	last := p.Board[3].UID
	if len(p.Board) != m.Slots(p) {
		t.Fatalf("board %d slots %d", len(p.Board), m.Slots(p))
	}
	if err := m.SellRelic(p, 0); err != nil {
		t.Fatal(err)
	}
	if len(p.Board) != 3 || len(p.Bench) != 1 || p.Bench[len(p.Bench)-1].UID != last {
		t.Fatalf("board=%d bench=%+v", len(p.Board), p.Bench)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestTemperingMergedBonus(t *testing.T) {
	for _, has := range []bool{true, false} {
		m := newPrepMatch(t, 1)
		p := m.P[0]
		if has {
			p.Relics = []RelicInst{relic(t, m, "tempering", 0)}
		}
		forceShop(t, m, p, "soldier", "soldier", "soldier")
		for i := 0; i < 3; i++ {
			if _, err := m.BuyCard(p, i); err != nil {
				t.Fatal(err)
			}
		}
		var got *Card
		for i := range p.Bench {
			if p.Bench[i].Star == 2 {
				got = &p.Bench[i]
			}
		}
		for i := range p.Board {
			if p.Board[i].Star == 2 {
				got = &p.Board[i]
			}
		}
		if got == nil {
			t.Fatalf("no 2★: board=%+v bench=%+v", p.Board, p.Bench)
		}
		want := 0
		if has {
			want = 3
		}
		if got.Bonus != want {
			t.Fatalf("tempering=%v bonus = %d, want %d", has, got.Bonus, want)
		}
	}
}

func distinctCats(t *testing.T, m *Match, offer []int) {
	t.Helper()
	seen := map[string]bool{}
	for _, id := range offer {
		c := m.cfg.CategoryOf(id)
		if seen[c] {
			t.Fatalf("duplicate category %q in %v", c, offer)
		}
		seen[c] = true
	}
}

func TestVaultOffersDistinctCategories(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		m := newPrepMatch(t, seed)
		for _, p := range m.P {
			p.Family = mustFamily(t, m.cfg, "noble")
		}
		m.openVault()
		for _, p := range m.P {
			if len(p.VaultOffer) < 3 {
				t.Fatalf("seed %d: offer %v", seed, p.VaultOffer)
			}
			distinctCats(t, m, p.VaultOffer[:3])
		}
	}
}

func TestVaultFallsBackWhenCategoriesRunOut(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Relics = nil
	for i := range m.cfg.Relics {
		if m.cfg.CategoryOf(i) != "tactic" {
			p.Relics = append(p.Relics, RelicInst{Relic: i, Choice: -1})
		}
	}
	m.openVault()
	if len(p.VaultOffer) < 3 {
		t.Fatalf("offer %v", p.VaultOffer)
	}
	for _, id := range p.VaultOffer {
		if m.cfg.CategoryOf(id) != "tactic" {
			t.Fatalf("non-tactic %d in offer", id)
		}
	}
}

func TestNobleVaultOffersDistinctCategories(t *testing.T) {
	cfg := loadTestConfig(t)
	cats := map[string]bool{}
	for i := range cfg.Relics {
		cats[cfg.CategoryOf(i)] = true
	}
	noble := mustFamily(t, cfg, "noble")
	for seed := int64(1); seed <= 100; seed++ {
		m := newPrepMatch(t, seed)
		for _, p := range m.P {
			p.Family = noble
		}
		m.openVault()
		for _, p := range m.P {
			if len(p.VaultOffer) != 4 {
				t.Fatalf("seed %d: offer %v", seed, p.VaultOffer)
			}
			// Only 3 relic categories exist in the data so far: the first 3 must differ,
			// the 4th may repeat; with 4+ categories all four would differ.
			distinctCats(t, m, p.VaultOffer[:min(3, len(cats))])
			seen := map[string]bool{}
			for _, id := range p.VaultOffer {
				seen[m.cfg.CategoryOf(id)] = true
			}
			if want := min(4, len(cats)); len(seen) != want {
				t.Fatalf("seed %d: %d distinct categories, want %d", seed, len(seen), want)
			}
		}
	}
}

// arrangeSlotsSetup gives a player [expansion, mirror] (mirror last = no copy, +1 slot) or
// [mirror, expansion] (mirror copies expansion, +2 slots) and enough cards for slots+1.
func arrangeSlotsSetup(t *testing.T, mirrorFirst bool) (*Match, *Player, []RelicInst) {
	t.Helper()
	m := newPrepMatch(t, 1)
	p := m.P[0]
	ex := RelicInst{Relic: mustRelic(t, m.cfg, "expansion"), Choice: -1}
	mi := RelicInst{Relic: mustRelic(t, m.cfg, "mirror"), Choice: -1}
	if mirrorFirst {
		p.Relics = []RelicInst{mi, ex}
	} else {
		p.Relics = []RelicInst{ex, mi}
	}
	for i := 0; i < m.Slots(p)+1; i++ {
		p.Bench = append(p.Bench, card(m, t, "soldier", 1))
	}
	return m, p, append([]RelicInst{}, p.Relics...)
}

func TestArrangeRelicOrderShrinkingSlotsRejected(t *testing.T) {
	m, p, relics := arrangeSlotsSetup(t, true) // [mirror, expansion]
	n := m.Slots(p)
	var uids []string
	for _, c := range p.Bench[:n] {
		uids = append(uids, c.UID)
	}
	if err := m.Arrange(p, uids, nil); err != nil {
		t.Fatal(err)
	}
	board0, bench0 := append([]Card{}, p.Board...), append([]Card{}, p.Bench...)
	if err := m.Arrange(p, uids, []int{1, 0}); err != ErrBadLayout { // mirror moved last: slots shrink
		t.Fatalf("err = %v, want ErrBadLayout", err)
	}
	if !reflect.DeepEqual(p.Relics, relics) || !reflect.DeepEqual(p.Board, board0) || !reflect.DeepEqual(p.Bench, bench0) {
		t.Fatal("state mutated by rejected Arrange")
	}
}

func TestArrangeRelicOrderExpandingSlotsAccepted(t *testing.T) {
	m, p, _ := arrangeSlotsSetup(t, false) // [expansion, mirror]
	n := m.Slots(p)
	var uids []string
	for _, c := range p.Bench[:n+1] {
		uids = append(uids, c.UID)
	}
	if err := m.Arrange(p, uids, nil); err != ErrBadLayout {
		t.Fatalf("without reorder err = %v, want ErrBadLayout", err)
	}
	if err := m.Arrange(p, uids, []int{1, 0}); err != nil { // mirror first: slots grow
		t.Fatal(err)
	}
	if len(p.Board) != n+1 || m.Slots(p) != n+1 {
		t.Fatalf("board %d slots %d", len(p.Board), m.Slots(p))
	}
}

// ---------- suit relics ----------

func TestCinnabarPerRedCard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	board := []Card{coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "clerk", 1, Red), coloredCard(m, t, "monk", 1, Black)}
	st := stepsBy(m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "cinnabar", 0)}), "朱砂")
	if len(st) != 2 {
		t.Fatalf("朱砂 steps %+v, want 2", st)
	}
	for i, s := range st {
		if s.Phase != 1 || s.Kind != StepAddMult || s.Value != 3 || s.CardUID != board[i].UID {
			t.Fatalf("step %+v", s)
		}
	}
}

func TestPineInkPerBlackCard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	board := []Card{coloredCard(m, t, "soldier", 1, Black), coloredCard(m, t, "clerk", 1, Red)}
	st := stepsBy(m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "pine_ink", 0)}), "松烟墨")
	if len(st) != 1 || st[0].Phase != 1 || st[0].Kind != StepPoints || st[0].Value != 15 || st[0].CardUID != board[0].UID {
		t.Fatalf("松烟墨 steps %+v", st)
	}
}

func TestShrineDoublesGodCards(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	clerk, soldier := coloredCard(m, t, "clerk", 1, Red), coloredCard(m, t, "soldier", 1, Red)
	res := m.ScoreWith(p, []Card{clerk, soldier}, nil, []RelicInst{relic(t, m, "shrine", 0)})
	st := stepsBy(res, "神龛")
	if len(st) != 1 || st[0].Phase != 1 || st[0].Kind != StepPoints || st[0].Value != float64(m.Points(clerk)) || st[0].CardUID != clerk.UID {
		t.Fatalf("神龛 steps %+v", st)
	}
	if want := float64(2*m.Points(clerk) + m.Points(soldier)); res.Points != want {
		t.Fatalf("points %v want %v", res.Points, want)
	}
}

func TestVanguardUsesFactoredPoints(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	clerk := coloredCard(m, t, "clerk", 1, Red)
	// 先锋旗在神龛左边也一样：神龛属于更早的逐卡倍率段
	res := m.ScoreWith(p, []Card{clerk}, nil, []RelicInst{relic(t, m, "vanguard", 0), relic(t, m, "shrine", 0)})
	st := stepsBy(res, "先锋旗")
	if want := float64(2 * (2 * m.Points(clerk))); len(st) != 1 || st[0].Value != want {
		t.Fatalf("先锋旗 steps %+v, want value %v", st, want)
	}
	if s := stepsBy(res, "神龛"); len(s) != 1 || res.Steps[1].Source != "神龛" {
		t.Fatalf("shrine must come right after the card's own step: %+v", res.Steps)
	}
}

func suitBoard(m *Match, t *testing.T, specs ...string) []Card {
	var out []Card
	for i := 0; i < len(specs); i += 2 {
		c := Red
		if specs[i+1] == "b" {
			c = Black
		}
		out = append(out, coloredCard(m, t, specs[i], 1, c))
	}
	return out
}

func TestNightParade(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	rel := []RelicInst{relic(t, m, "night_parade", 0)}
	b := suitBoard(m, t, "carpenter", "r", "farmer", "b", "scholar", "r", "artist", "b", "clerk", "r")
	st := stepsBy(m.ScoreWith(p, b, nil, rel), "百鬼夜行")
	if len(st) != 1 || st[0].Phase != 5 || st[0].Kind != StepTimesMult || st[0].Value != 2 {
		t.Fatalf("4 ghosts: %+v", st)
	}
	if st := stepsBy(m.ScoreWith(p, b[1:], nil, rel), "百鬼夜行"); len(st) != 0 {
		t.Fatalf("3 ghosts: %+v", st)
	}
}

func TestOneColor(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	rel := []RelicInst{relic(t, m, "one_color", 0)}
	b := suitBoard(m, t, "carpenter", "r", "clerk", "r", "monk", "r")
	st := stepsBy(m.ScoreWith(p, b, nil, rel), "一色")
	if len(st) != 1 || st[0].Kind != StepTimesMult || st[0].Value != 3 {
		t.Fatalf("3 red: %+v", st)
	}
	if st := stepsBy(m.ScoreWith(p, b[:2], nil, rel), "一色"); len(st) != 0 {
		t.Fatalf("2 red: %+v", st)
	}
	b[2].Color = Black
	if st := stepsBy(m.ScoreWith(p, b, nil, rel), "一色"); len(st) != 0 {
		t.Fatalf("2 red 1 black: %+v", st)
	}
	bb := suitBoard(m, t, "carpenter", "b", "clerk", "b", "monk", "b")
	if st := stepsBy(m.ScoreWith(p, bb, nil, rel), "一色"); len(st) != 1 {
		t.Fatalf("3 black: %+v", st)
	}
}

func TestYinYang(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	rel := []RelicInst{relic(t, m, "yin_yang", 0)}
	cases := []struct {
		b    []Card
		want int
	}{
		{suitBoard(m, t, "carpenter", "r", "clerk", "r", "monk", "b", "farmer", "b"), 1},
		{suitBoard(m, t, "carpenter", "r", "clerk", "b"), 0},
		{suitBoard(m, t, "carpenter", "r", "clerk", "r", "monk", "r", "farmer", "b"), 0},
	}
	for i, c := range cases {
		st := stepsBy(m.ScoreWith(p, c.b, nil, rel), "阴阳鱼")
		if len(st) != c.want || (c.want == 1 && (st[0].Kind != StepTimesMult || st[0].Value != 2)) {
			t.Fatalf("case %d: %+v", i, st)
		}
	}
}

func TestFourSymbols(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	rel := []RelicInst{relic(t, m, "four_symbols", 0)}
	b := suitBoard(m, t, "clerk", "r", "monk", "b", "carpenter", "r", "farmer", "b")
	st := stepsBy(m.ScoreWith(p, b, nil, rel), "四象")
	if len(st) != 1 || st[0].Phase != 5 || st[0].Kind != StepAddMult || st[0].Value != 20 {
		t.Fatalf("all four: %+v", st)
	}
	b[3].Color = Red // 黑鬼没了
	if st := stepsBy(m.ScoreWith(p, b, nil, rel), "四象"); len(st) != 0 {
		t.Fatalf("missing black ghost: %+v", st)
	}
}

func TestVermilionMergeIsRed(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Relics = []RelicInst{relic(t, m, "vermilion", 0)}
	forceShopColored(t, m, p, []string{"soldier", "soldier", "soldier"}, []Color{Black, Black, Black})
	for i := 0; i < 3; i++ {
		if _, err := m.BuyCard(p, i); err != nil {
			t.Fatal(err)
		}
	}
	var got *Card
	for _, cs := range [][]Card{p.Board, p.Bench} {
		for i := range cs {
			if cs[i].Star == 2 {
				got = &cs[i]
			}
		}
	}
	if got == nil {
		t.Fatalf("no 2★: board=%+v bench=%+v", p.Board, p.Bench)
	}
	if got.Color != Red || got.Red != 0 {
		t.Fatalf("merged %+v, want Color Red with Red 0", *got)
	}
	if err := m.SellCard(p, got.UID); err != nil {
		t.Fatal(err)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved after selling the vermilion card")
	}
}

func TestSecretArtRepeatsFactors(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	clerk := coloredCard(m, t, "clerk", 1, Red)
	art := relic(t, m, "secret_art", 0)
	art.Choice = m.cfg.TradeOf(clerk.Line)
	if m.cfg.Trades[art.Choice].Key != "court" {
		t.Fatalf("clerk trade %q", m.cfg.Trades[art.Choice].Key)
	}
	st := stepsBy(m.ScoreWith(p, []Card{clerk}, nil, []RelicInst{art, relic(t, m, "shrine", 0)}), "神龛")
	if len(st) != 2 || st[0].Trigger != 1 || st[0].Triggers != 2 || st[1].Trigger != 2 || st[1].Triggers != 2 {
		t.Fatalf("神龛 steps %+v", st)
	}
	if st[0].Value != float64(m.Points(clerk)) || st[1].Value != float64(m.Points(clerk)) {
		t.Fatalf("each retrigger doubles fresh card points: %+v", st)
	}
}

func TestOmenTableComplete(t *testing.T) {
	cfg := loadTestConfig(t)
	if cfg.MaxOmenPoint() != 12 || len(cfg.Omens) != 12 {
		t.Fatalf("omens %d max %d", len(cfg.Omens), cfg.MaxOmenPoint())
	}
	for pt := 1; pt <= 12; pt++ {
		if cfg.OmenByPoint(pt) < 0 {
			t.Fatalf("no omen at %d", pt)
		}
	}
	if cfg.Omens[cfg.OmenByPoint(1)].Key != "quake" || cfg.Omens[cfg.OmenByPoint(12)].Key != "upheaval" {
		t.Fatal("point table order")
	}
	for _, r := range []int{6, 11, 16, 21, 26} {
		if !cfg.IsOmenRound(r) {
			t.Fatalf("round %d", r)
		}
	}
	if cfg.IsOmenRound(7) || cfg.Rules.Omen.Duration != 6 {
		t.Fatal("rounds/duration")
	}
}

func TestOmenTableRejectsDuplicatePoint(t *testing.T) {
	tt := loadTestTables(t)
	tt.O.Omens[1].Point = 1
	_, err := tt.build()
	if err == nil || !strings.Contains(err.Error(), "chess_omens: point 1 used twice") {
		t.Fatalf("err = %v", err)
	}
}

// ---------- omens ----------

// toPrep advances until the given round's prep phase (HP is raised so nobody dies on the way).
func toPrep(t *testing.T, m *Match, round int) {
	t.Helper()
	for _, p := range m.P {
		p.HP = 10000
	}
	for !(m.Round == round && m.Phase == PhasePrep) {
		if m.Round > round {
			t.Fatalf("passed round %d (now %d, phase %d)", round, m.Round, m.Phase)
		}
		advance(t, m)
	}
}

func TestOmenRoundEntersOmenPhase(t *testing.T) {
	m := newPrepMatch(t, 1)
	toPrep(t, m, 5)
	if m.Phase != PhasePrep {
		t.Fatalf("round 5 phase = %d, want prep", m.Phase)
	}
	for _, p := range m.P {
		p.Relics = nil
	}
	gold := [2]int{m.P[0].Gold, m.P[1].Gold}
	advance(t, m)
	if m.Round != 6 || m.Phase != PhaseOmen {
		t.Fatalf("round %d phase %d, want round 6 omen", m.Round, m.Phase)
	}
	for i, p := range m.P {
		if p.Gold != gold[i] {
			t.Fatalf("seat %d gold %d -> %d on entering omen; income must wait for ResolveOmen", i, gold[i], p.Gold)
		}
	}
	if err := m.ResolveOmen(); err != nil {
		t.Fatal(err)
	}
	if m.P[0].Gold <= gold[0] {
		t.Fatal("income not paid after ResolveOmen")
	}
}

func TestChooseOmenRules(t *testing.T) {
	m := newPrepMatch(t, 1)
	if err := m.ChooseOmen(m.P[0], true); err != ErrWrongPhase {
		t.Fatalf("prep ChooseOmen err = %v, want ErrWrongPhase", err)
	}
	toPrep(t, m, 5)
	advance(t, m)
	p := m.P[0]
	if p.Ready || p.OmenChoice != OmenUndecided {
		t.Fatal("omen phase must start undecided and unready")
	}
	if err := m.ChooseOmen(p, false); err != nil {
		t.Fatal(err)
	}
	if !p.Ready || p.OmenChoice != OmenSkips {
		t.Fatalf("after choose: ready=%v choice=%d", p.Ready, p.OmenChoice)
	}
	if err := m.ChooseOmen(p, true); err != ErrAlreadyChosen {
		t.Fatalf("second ChooseOmen err = %v, want ErrAlreadyChosen", err)
	}
	if err := m.ResolveOmen(); err != nil {
		t.Fatal(err)
	}
	if err := m.ResolveOmen(); err != ErrWrongPhase {
		t.Fatalf("ResolveOmen in prep err = %v, want ErrWrongPhase", err)
	}
}

func TestResolveOmenDiceSum(t *testing.T) {
	m := newPrepMatch(t, 1)
	toPrep(t, m, 5)
	advance(t, m)
	if err := m.ChooseOmen(m.P[0], true); err != nil {
		t.Fatal(err)
	}
	// seat 1 never chooses: counts as a skip
	if err := m.ResolveOmen(); err != nil {
		t.Fatal(err)
	}
	r := m.LastRoll
	if r.Dice[0] < 1 || r.Dice[0] > 6 || r.Dice[1] != 0 {
		t.Fatalf("dice = %v, want [1..6, 0]", r.Dice)
	}
	if r.Forced || r.Round != 6 || m.Phase != PhasePrep {
		t.Fatalf("roll %+v phase %d", r, m.Phase)
	}
	if !r.Shifted && r.Point != r.Dice[0] {
		t.Fatalf("point %d != dice sum %d", r.Point, r.Dice[0])
	}
	if r.Omen != m.cfg.OmenByPoint(r.Point) || r.Omen < 0 {
		t.Fatalf("omen %d does not match point %d", r.Omen, r.Point)
	}
}

func TestBothSkipForcesRoll(t *testing.T) {
	m := newPrepMatch(t, 2)
	toPrep(t, m, 5)
	advance(t, m)
	for _, p := range m.P {
		if err := m.ChooseOmen(p, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.ResolveOmen(); err != nil {
		t.Fatal(err)
	}
	r := m.LastRoll
	if !r.Forced || r.Dice[0] < 1 || r.Dice[0] > 6 || r.Dice[1] < 1 || r.Dice[1] > 6 {
		t.Fatalf("roll %+v, want forced with two dice", r)
	}
	if !r.Shifted && r.Point != r.Dice[0]+r.Dice[1] {
		t.Fatalf("point %d != %d", r.Point, r.Dice[0]+r.Dice[1])
	}
}

func TestOmenLastsSixRounds(t *testing.T) {
	m := newPrepMatch(t, 3)
	if m.LastRoll.Omen != -1 {
		t.Fatalf("initial LastRoll.Omen = %d, want -1", m.LastRoll.Omen)
	}
	has6 := func() bool {
		for _, o := range m.ActiveOmens() {
			if o.From == 6 {
				return true
			}
		}
		return false
	}
	toPrep(t, m, 6)
	if len(m.Omens) != 1 || m.Omens[0].From != 6 || m.Omens[0].Until != 11 {
		t.Fatalf("omens = %+v", m.Omens)
	}
	key := m.cfg.Omens[m.Omens[0].Omen].Key
	for r := 6; r <= 11; r++ {
		toPrep(t, m, r)
		if !has6() || !m.OmenActive(key) {
			t.Fatalf("round %d: omen from round 6 not active", r)
		}
	}
	if n := len(m.ActiveOmens()); n != 2 {
		t.Fatalf("round 11 active omens = %d, want 2", n)
	}
	if a := m.ActiveOmens(); a[0].From != 6 || a[1].From != 11 {
		t.Fatalf("active omens out of order: %+v", a)
	}
	toPrep(t, m, 12)
	if has6() {
		t.Fatal("round 12: omen from round 6 still active")
	}
	if len(m.Omens) != 2 {
		t.Fatal("omen history must be kept")
	}
}

func TestOmenDuplicateShiftsAndWraps(t *testing.T) {
	m := newPrepMatch(t, 1)
	forceOmen(t, m, "upheaval")
	forceOmen(t, m, "quake")
	m.resolveOmenDice([2]int{6, 6}, false)
	r := m.LastRoll
	eclipse, _ := m.cfg.OmenIndex("solar_eclipse")
	if r.Point != 2 || r.Omen != eclipse || !r.Shifted {
		t.Fatalf("roll %+v, want point 2 solar_eclipse shifted", r)
	}
	if last := m.Omens[len(m.Omens)-1]; last.Omen != eclipse || last.From != m.Round {
		t.Fatalf("appended omen %+v", last)
	}

	// every omen active: the shift must stop and stay in range
	m2 := newPrepMatch(t, 1)
	for _, o := range m2.cfg.Omens {
		forceOmen(t, m2, o.Key)
	}
	m2.resolveOmenDice([2]int{3, 4}, false)
	if p := m2.LastRoll.Point; p < 1 || p > m2.cfg.MaxOmenPoint() || m2.LastRoll.Omen < 0 {
		t.Fatalf("all-active roll %+v out of range", m2.LastRoll)
	}
}

func TestCloseVaultCanEnterOmen(t *testing.T) {
	m := newPrepMatch(t, 4)
	if !m.cfg.IsVaultRound(10) || !m.cfg.IsOmenRound(11) {
		t.Fatal("default config must have vault round 10 and omen round 11")
	}
	toPrep(t, m, 10)
	advance(t, m)
	if m.Phase != PhaseVault {
		t.Fatalf("after round 10 settle phase = %d, want vault", m.Phase)
	}
	if err := m.CloseVault(); err != nil {
		t.Fatal(err)
	}
	if m.Phase != PhaseOmen || m.Round != 11 {
		t.Fatalf("phase %d round %d, want omen round 11", m.Phase, m.Round)
	}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// foeSetup gives the opponent (P[1]) the relics; P[0] fields the board.
func foeSetup(t *testing.T, m *Match, board []Card, foeRelics ...string) *Player {
	p, o := m.P[0], m.P[1]
	p.Board = board
	o.Relics = nil
	for _, k := range foeRelics {
		o.Relics = append(o.Relics, relic(t, m, k, 0))
	}
	return p
}

func TestDemonMirrorHitsFoeGhosts(t *testing.T) {
	m := newPrepMatch(t, 1)
	soldier, clerk := coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "clerk", 1, Red)
	p := foeSetup(t, m, []Card{soldier, clerk}, "demon_mirror")
	st := stepsBy(m.Score(p), "照妖镜")
	if len(st) != 1 || st[0].Origin != OriginFoeRelic || st[0].Relic != 0 || st[0].CardUID != soldier.UID ||
		st[0].Kind != StepPoints || st[0].Phase != 1 || !approx(st[0].Value, -0.3*float64(m.Points(soldier))) {
		t.Fatalf("照妖镜 steps %+v", st)
	}
}

func TestGodBreakerHitsFoeGods(t *testing.T) {
	m := newPrepMatch(t, 1)
	soldier, clerk := coloredCard(m, t, "soldier", 1, Red), coloredCard(m, t, "clerk", 1, Red)
	p := foeSetup(t, m, []Card{soldier, clerk}, "god_breaker")
	st := stepsBy(m.Score(p), "破神符")
	if len(st) != 1 || st[0].Origin != OriginFoeRelic || st[0].Relic != 0 || st[0].CardUID != clerk.UID ||
		!approx(st[0].Value, -0.3*float64(m.Points(clerk))) {
		t.Fatalf("破神符 steps %+v", st)
	}
}

func TestBeheadZeroesLeftmost(t *testing.T) {
	m := newPrepMatch(t, 1)
	clerk, soldier := coloredCard(m, t, "clerk", 1, Red), coloredCard(m, t, "soldier", 1, Red)
	p := foeSetup(t, m, []Card{clerk, soldier}, "behead")
	p.Relics = []RelicInst{relic(t, m, "vanguard", 0)}
	res := m.Score(p)
	st := stepsBy(res, "斩首")
	if len(st) != 1 || st[0].Origin != OriginFoeRelic || st[0].CardUID != clerk.UID || st[0].Value != -float64(m.Points(clerk)) {
		t.Fatalf("斩首 steps %+v", st)
	}
	if s := stepsBy(res, "先锋旗"); len(s) != 0 {
		t.Fatalf("先锋旗 should not fire on a beheaded card: %+v", s)
	}
	net := 0.0
	for _, s := range res.Steps {
		if s.CardUID == clerk.UID && s.Kind == StepPoints {
			net += s.Value
		}
	}
	if net != 0 {
		t.Fatalf("beheaded card net points %v", net)
	}
}

func TestMirroredDemonMirrorStacks(t *testing.T) {
	m := newPrepMatch(t, 1)
	soldier := coloredCard(m, t, "soldier", 1, Red)
	p := foeSetup(t, m, []Card{soldier}, "mirror", "demon_mirror")
	st := stepsBy(m.Score(p), "照妖镜")
	pts := float64(m.Points(soldier))
	if len(st) != 2 || st[0].Relic != 0 || st[1].Relic != 1 ||
		!approx(st[0].Value, -0.3*pts) || !approx(st[1].Value, -0.3*0.7*pts) {
		t.Fatalf("照妖镜 steps %+v", st)
	}
}

func TestColorCounter(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := foeSetup(t, m, suitBoard(m, t, "soldier", "r", "clerk", "r", "monk", "b"))
	p.Relics = []RelicInst{relic(t, m, "color_counter", 0)}
	m.P[1].Board = suitBoard(m, t, "farmer", "r", "smith", "r", "taoist", "b")
	st := stepsBy(m.Score(p), "克色")
	if len(st) != 1 || st[0].Phase != 5 || st[0].Kind != StepAddMult || st[0].Value != 8 || st[0].Origin != OriginOwn {
		t.Fatalf("克色 steps %+v", st)
	}
	m.P[1].Board = suitBoard(m, t, "farmer", "r", "taoist", "b")
	if st := stepsBy(m.Score(p), "克色"); len(st) != 0 {
		t.Fatalf("tied foe colors must not fire: %+v", st)
	}
}

func TestUsurpCountsFoeSuitRelics(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := foeSetup(t, m, suitBoard(m, t, "soldier", "r"), "cinnabar", "shrine", "vanguard")
	p.Relics = []RelicInst{relic(t, m, "usurp", 0)}
	st := stepsBy(m.Score(p), "反客为主")
	if len(st) != 1 || st[0].Phase != 5 || st[0].Kind != StepAddMult || st[0].Value != 10 {
		t.Fatalf("反客为主 steps %+v", st)
	}
}

func TestPreviewUsesOpponentLastBoard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := foeSetup(t, m, suitBoard(m, t, "soldier", "r", "clerk", "r"))
	p.Relics = []RelicInst{relic(t, m, "color_counter", 0)}
	o := m.P[1]
	o.LastBoard = suitBoard(m, t, "farmer", "r", "smith", "r")
	o.Board = suitBoard(m, t, "taoist", "b", "monk", "b", "artist", "b")
	if st := stepsBy(m.Preview(p), "克色"); len(st) != 1 || st[0].Value != 8 {
		t.Fatalf("preview 克色 steps %+v, want +8 from the foe's last board", st)
	}
	if st := stepsBy(m.Score(p), "克色"); len(st) != 0 {
		t.Fatalf("settlement 克色 steps %+v, want none against the locked all-black board", st)
	}
	// bots arrange against the same view as the preview
	if st := stepsBy(m.ScoreWith(p, p.Board, p.Bench, p.Relics), "克色"); len(st) != 1 {
		t.Fatalf("ScoreWith 克色 steps %+v", st)
	}
}

// ---------- omen effects (Task 14) ----------

// ownCard builds a 1★ card of the given color and takes its copy out of the pool.
func ownCard(m *Match, t *testing.T, key string, c Color) Card {
	cd := coloredCard(m, t, key, 1, c)
	takeFromPool(m, cd.Line, c)
	return cd
}

func omenSteps(r ScoreResult, name string) []Step {
	var out []Step
	for _, s := range stepsBy(r, name) {
		if s.Origin == OriginOmen {
			out = append(out, s)
		}
	}
	return out
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestQuakeSellsBenchOverflow(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	keys := []string{"carpenter", "smith", "farmer", "taoist", "monk", "scholar", "doctor", "artist"}
	for _, k := range keys {
		p.Bench = append(p.Bench, ownCard(m, t, k, Red))
	}
	if len(p.Bench) != 8 || !m.CheckPool() {
		t.Fatal("setup")
	}
	g := p.Gold
	sold := []Card{p.Bench[6], p.Bench[7]}
	kept := append([]Card{}, p.Bench[:6]...)
	forceOmen(t, m, "quake")
	if m.BenchSize(p) != 6 {
		t.Fatalf("bench size %d, want 6", m.BenchSize(p))
	}
	m.fitBoard(p)
	if len(p.Bench) != 6 {
		t.Fatalf("bench %d, want 6", len(p.Bench))
	}
	if want := g + m.SellPrice(sold[0]) + m.SellPrice(sold[1]); p.Gold != want {
		t.Fatalf("gold %d, want %d", p.Gold, want)
	}
	if !reflect.DeepEqual(p.Bench, kept) {
		t.Fatalf("the two rightmost cards must be sold: bench %+v", p.Bench)
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestEclipseHalvesGods(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	forceOmen(t, m, "solar_eclipse")
	clerk, soldier := coloredCard(m, t, "clerk", 1, Red), coloredCard(m, t, "soldier", 1, Red)
	res := m.ScoreWith(p, []Card{clerk, soldier}, nil, nil)
	st := stepsBy(res, "天狗食日")
	if len(st) != 1 || st[0].Origin != OriginOmen || st[0].Relic != -1 || st[0].Kind != StepPoints || st[0].Phase != 1 ||
		st[0].CardUID != clerk.UID || !near(st[0].Value, -0.6*float64(m.Points(clerk))) {
		t.Fatalf("天狗食日 steps %+v", st)
	}
}

func TestColorOmens(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	red, black := coloredCard(m, t, "clerk", 1, Red), coloredCard(m, t, "soldier", 1, Black)
	forceOmen(t, m, "red_moon")
	st := omenSteps(m.ScoreWith(p, []Card{red, black}, nil, nil), "赤月")
	if len(st) != 1 || st[0].CardUID != red.UID || !near(st[0].Value, -0.4*float64(m.Points(red))) {
		t.Fatalf("赤月 steps %+v", st)
	}
	m.Omens = nil
	forceOmen(t, m, "ink_rain")
	st = omenSteps(m.ScoreWith(p, []Card{red, black}, nil, nil), "墨雨")
	if len(st) != 1 || st[0].CardUID != black.UID || !near(st[0].Value, -0.4*float64(m.Points(black))) {
		t.Fatalf("墨雨 steps %+v", st)
	}
}

func TestUpheavalHitsMajoritySuit(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	forceOmen(t, m, "upheaval")
	board := suitBoard(m, t, "clerk", "r", "smith", "r", "soldier", "b")
	st := omenSteps(m.ScoreWith(p, board, nil, nil), "天地异变")
	if len(st) != 2 || st[0].CardUID != board[0].UID || st[1].CardUID != board[1].UID ||
		!near(st[0].Value, -0.5*float64(m.Points(board[0]))) {
		t.Fatalf("天地异变 steps %+v", st)
	}
	board = suitBoard(m, t, "clerk", "r", "soldier", "b")
	if st := omenSteps(m.ScoreWith(p, board, nil, nil), "天地异变"); len(st) != 0 {
		t.Fatalf("no suit has 2 cards: %+v", st)
	}
}

func TestFamineAndHeavyTax(t *testing.T) {
	var gain [2]int
	for i := range gain {
		m := newPrepMatch(t, 7)
		p := m.P[0]
		for _, q := range m.P {
			q.Gold = 30
		}
		if i == 1 {
			forceOmen(t, m, "famine")
		}
		advance(t, m)
		if m.Phase != PhasePrep {
			t.Fatalf("phase %d", m.Phase)
		}
		gain[i] = p.Gold - 30
	}
	if gain[0]-gain[1] != 2 {
		t.Fatalf("famine gains %v, want a difference of 2", gain)
	}
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Gold = 50
	if m.Interest(p) == 0 {
		t.Fatal("interest expected without heavy_tax")
	}
	forceOmen(t, m, "heavy_tax")
	if m.Interest(p) != 0 {
		t.Fatalf("heavy_tax interest %d", m.Interest(p))
	}
}

func TestMarketClosedAndShortage(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Gold = 10
	forceOmen(t, m, "market_closed")
	if m.RefreshCost(p) != 3 {
		t.Fatalf("refresh cost %d", m.RefreshCost(p))
	}
	if err := m.Refresh(p); err != nil || p.Gold != 7 {
		t.Fatalf("refresh: %v gold %d", err, p.Gold)
	}
	m.Omens = nil
	forceOmen(t, m, "shortage")
	if err := m.Refresh(p); err != nil {
		t.Fatal(err)
	}
	if len(p.Shop) != 4 || m.ShopSize() != 4 {
		t.Fatalf("shop %d size %d, want 4", len(p.Shop), m.ShopSize())
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestNarrowPathFitsBoard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Level = 3
	if m.Slots(p) != 5 {
		t.Fatalf("slots %d, want 5", m.Slots(p))
	}
	for _, k := range []string{"carpenter", "smith", "farmer", "taoist", "monk"} {
		p.Board = append(p.Board, ownCard(m, t, k, Red))
	}
	forceOmen(t, m, "narrow_path")
	m.fitBoard(p)
	if len(p.Board) != 4 || len(p.Bench) != 1 || m.Slots(p) != 4 {
		t.Fatalf("board %d bench %d slots %d", len(p.Board), len(p.Bench), m.Slots(p))
	}
	if !m.CheckPool() {
		t.Fatal("pool not conserved")
	}
}

func TestDiscordRaisesThresholds(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Board = []Card{card(m, t, "carpenter", 1), card(m, t, "smith", 1)}
	craft := m.cfg.tradeIdx["craft"]
	if tier := m.Bonds(p).Trade[craft].Tier; tier != 0 {
		t.Fatalf("craft tier %d, want 0", tier)
	}
	forceOmen(t, m, "discord")
	if tier := m.Bonds(p).Trade[craft].Tier; tier != -1 {
		t.Fatalf("craft tier under discord %d, want -1", tier)
	}
	if m.BondAtDelta() != 1 {
		t.Fatalf("delta %d", m.BondAtDelta())
	}
}

func TestDefyHeaven(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	forceOmen(t, m, "famine")
	forceOmen(t, m, "heavy_tax")
	res := m.ScoreWith(p, []Card{card(m, t, "clerk", 1)}, nil, []RelicInst{relic(t, m, "defy_heaven", 0)})
	st := stepsBy(res, "逆天")
	if len(st) != 2 {
		t.Fatalf("逆天 steps %+v", st)
	}
	for _, s := range st {
		if s.Kind != StepTimesMult || s.Value != 1.5 || s.Phase != 5 || s.Origin != OriginOwn {
			t.Fatalf("逆天 step %+v", s)
		}
	}
}

func TestExtraOmensDoNotPersist(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	rm, _ := m.cfg.OmenIndex("red_moon")
	red := coloredCard(m, t, "clerk", 1, Red)
	o := m.Opponent(p)
	res := m.ScoreVs(p, []Card{red}, nil, nil, Foe{o.LastBoard, o.Relics}, []int{rm})
	if st := omenSteps(res, "赤月"); len(st) != 1 {
		t.Fatalf("赤月 steps from extraOmens %+v", st)
	}
	if n := len(m.ActiveOmens()); n != 0 {
		t.Fatalf("active omens %d after ScoreVs", n)
	}
}
