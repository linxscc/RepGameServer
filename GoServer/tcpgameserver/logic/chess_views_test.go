package logic

import (
	"math/rand"
	"testing"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/rules"
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
