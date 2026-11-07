package rules

import (
	"math/rand"
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
	board := []Card{card(m, t, "taoist", 1), card(m, t, "monk", 1), card(m, t, "doctor", 1)} // 方外 3
	if got := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "tailwind", 0)}).Score; got != 33 {
		t.Fatalf("乘风 = %v, want 33", got)
	}
	if got := m.ScoreWith(p, board, nil, []RelicInst{relic(t, m, "furnace", 4)}).Score; got != 77 {
		t.Fatalf("熔炉 = %v, want 77", got)
	}
	if got := m.ScoreWith(p, board[:2], nil, []RelicInst{relic(t, m, "furnace", 4)}).Score; got != 32 {
		t.Fatalf("方外 2 = %v, want 32", got)
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

func TestOpsBondsUseLastBoard(t *testing.T) {
	m := newPrepMatch(t, 1)
	p := m.P[0]
	p.Board = []Card{card(m, t, "smith", 1), card(m, t, "doctor", 1)} // 市井 2
	if m.FreeRefreshesLeft(p) != 0 {
		t.Fatal("ops bonds must use the last settled board")
	}
	p.LastBoard = p.Board
	if m.FreeRefreshesLeft(p) != 1 {
		t.Fatal("expected a free refresh")
	}
	gold := p.Gold
	if err := m.Refresh(p); err != nil || p.Gold != gold {
		t.Fatalf("free refresh should cost nothing: %v, gold %d→%d", err, gold, p.Gold)
	}
	if err := m.Refresh(p); err != nil || p.Gold != gold-m.cfg.Rules.Economy.RefreshCost {
		t.Fatalf("second refresh should cost gold: %v", err)
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
	bc := m.bondCounts([]Card{card(m, t, "farmer", 3), card(m, t, "monk", 1), card(m, t, "soldier", 1)}, nil)
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
	advance := func() {
		if _, err := m.Settle(); err != nil {
			t.Fatal(err)
		}
		if m.Phase == PhaseVault {
			if err := m.CloseVault(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for r := 0; r < m.cfg.Rules.Board.MaxLockRounds; r++ {
		if err := m.SetLock(p, true); err != nil {
			t.Fatal(err)
		}
		advance()
		for i := range first {
			if p.Shop[i] != first[i] {
				t.Fatalf("locked shop changed after lock %d", r+1)
			}
		}
	}
	_ = m.SetLock(p, true)
	advance()
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
	if _, err := m.Settle(); err != nil {
		t.Fatal(err)
	}
	if m.Phase == PhaseVault {
		if err := m.CloseVault(); err != nil {
			t.Fatal(err)
		}
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
