package match

import (
	"math/rand"
	"testing"

	"GoServer/tcpgameserver/data/config"
	"GoServer/tcpgameserver/logic/rules"
	"GoServer/tcpgameserver/pb"
)

func TestViewsCarryColorAndKind(t *testing.T) {
	cfg := config.Game().Chess
	m := rules.NewMatch(cfg, rand.New(rand.NewSource(1)), [2]string{"a", "b"})
	p := m.Player("a")
	p.Board = []rules.Card{{UID: "u1", Line: 0, Star: 1, Color: rules.Red, Red: 1}}
	p.Shop = []int{0, -1, 0}
	p.ShopColor = []rules.Color{rules.Black, rules.Red, rules.Red}

	sv := selfView(m, p)
	if got := sv.GetBoard()[0].GetColor(); got != pb.CardColor_CARD_COLOR_RED {
		t.Fatalf("board card color = %v", got)
	}
	if got := sv.GetShop()[0].GetColor(); got != pb.CardColor_CARD_COLOR_BLACK {
		t.Fatalf("shop[0] color = %v", got)
	}
	if got := sv.GetShop()[1].GetColor(); got != pb.CardColor_CARD_COLOR_UNSPECIFIED {
		t.Fatalf("empty shop slot color = %v", got)
	}
	if got := sv.GetShop()[2].GetColor(); got != pb.CardColor_CARD_COLOR_RED {
		t.Fatalf("shop[2] color = %v", got)
	}

	mc := matchConfigMessage("r", m).GetMatchConfig()
	if len(mc.GetLines()) != len(cfg.Lines) {
		t.Fatal("line count mismatch")
	}
	for i, li := range mc.GetLines() {
		want := pb.CardKind_CARD_KIND_GOD
		if cfg.Lines[i].Kind == rules.KindGhost {
			want = pb.CardKind_CARD_KIND_GHOST
		}
		if li.GetKind() != want {
			t.Fatalf("line %s kind = %v, want %v", li.GetKey(), li.GetKind(), want)
		}
	}
}

func TestViewsCarryRelicCategory(t *testing.T) {
	cfg := config.Game().Chess
	m := rules.NewMatch(cfg, rand.New(rand.NewSource(1)), [2]string{"a", "b"})
	mc := matchConfigMessage("r", m).GetMatchConfig()
	if len(mc.GetRelics()) != len(cfg.Relics) {
		t.Fatal("relic count mismatch")
	}
	for i, ri := range mc.GetRelics() {
		if ri.GetCategory() != cfg.Relics[i].Category || ri.GetCategoryName() == "" {
			t.Fatalf("%s category %q name %q", ri.GetKey(), ri.GetCategory(), ri.GetCategoryName())
		}
	}
}

func TestScoreStepOriginInView(t *testing.T) {
	cfg := config.Game().Chess
	m := rules.NewMatch(cfg, rand.New(rand.NewSource(1)), [2]string{"a", "b"})
	p, o := m.P[0], m.P[1]
	soldier, _ := cfg.LineIndex("soldier")
	clerk, _ := cfg.LineIndex("clerk")
	dm, _ := cfg.RelicIndex("demon_mirror")
	p.Board = []rules.Card{{UID: "s", Line: soldier, Star: 1, Color: rules.Red, Red: 1}, {UID: "c", Line: clerk, Star: 1, Color: rules.Red, Red: 1}}
	o.Relics = []rules.RelicInst{{Relic: dm, Choice: -1}}
	var found bool
	for _, s := range scoreBreakdown(m.Score(p)).GetSteps() {
		if s.GetSource() == "照妖镜" {
			found = true
			if s.GetOrigin() != pb.StepOrigin_STEP_ORIGIN_FOE_RELIC || s.GetRelicSlot() != 0 || s.GetCardUid() != "s" {
				t.Fatalf("step %+v", s)
			}
		} else if s.GetOrigin() != pb.StepOrigin_STEP_ORIGIN_OWN {
			t.Fatalf("own step with origin %v: %+v", s.GetOrigin(), s)
		}
	}
	if !found {
		t.Fatal("no 照妖镜 step")
	}
}

func TestViewsCarryOmenFields(t *testing.T) {
	cfg := config.Game().Chess
	m := rules.NewMatch(cfg, rand.New(rand.NewSource(1)), [2]string{"a", "b"})
	m.Round = 3
	for _, key := range []string{"quake", "discord"} {
		i, ok := cfg.OmenIndex(key)
		if !ok {
			t.Fatalf("no omen %q", key)
		}
		m.Omens = append(m.Omens, rules.ActiveOmen{Omen: i, From: m.Round, Until: m.Round + cfg.Rules.Omen.Duration - 1})
	}
	sv := selfView(m, m.P[0])
	if sv.GetBenchSize() != 6 {
		t.Fatalf("bench size under 地动 = %d", sv.GetBenchSize())
	}
	if sv.GetRefreshCost() != int32(m.RefreshCost(m.P[0])) {
		t.Fatalf("refresh cost %d", sv.GetRefreshCost())
	}
	for _, b := range append(sv.GetTradeBonds(), sv.GetOriginBonds()...) {
		if b.GetAtDelta() != 1 {
			t.Fatalf("bond %s at_delta = %d under 失和", b.GetKey(), b.GetAtDelta())
		}
	}
	mc := matchConfigMessage("r", m).GetMatchConfig()
	if len(mc.GetOmens()) != 12 {
		t.Fatalf("omens in config = %d", len(mc.GetOmens()))
	}
	seen := map[int32]bool{}
	for _, o := range mc.GetOmens() {
		if o.GetPoint() < 1 || o.GetPoint() > 12 || seen[o.GetPoint()] || o.GetKey() == "" || o.GetName() == "" || o.GetDesc() == "" {
			t.Fatalf("bad omen info %v", o)
		}
		seen[o.GetPoint()] = true
	}
}
