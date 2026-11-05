package cards

import (
	"testing"

	"GoServer/tcpgameserver/config"
)

// 卡牌池完全由配置表生成，不依赖数据库
func TestCardPoolBuiltFromConfig(t *testing.T) {
	if err := InitCardPool(); err != nil {
		t.Fatal(err)
	}
	want := map[int]int{}
	for _, c := range config.Game().Cards.Cards {
		want[c.Level] += c.Count
	}
	got := map[int]int{1: len(GetLevel1Cards()), 2: len(GetLevel2Cards()), 3: len(GetLevel3Cards())}
	for level, n := range want {
		if got[level] != n {
			t.Fatalf("level %d pool: got %d cards, want %d", level, got[level], n)
		}
	}
	if c := GetLevel1Cards()[0]; c.Name == "" || c.UID == "" {
		t.Fatalf("pool card missing data: %+v", c)
	}
}

func TestBondPoolBuiltFromConfig(t *testing.T) {
	if err := InitBondPool(); err != nil {
		t.Fatal(err)
	}
	bonds := GetBondPoolManager().GetAllBonds()
	if len(bonds) != len(config.Game().Bonds.Bonds) {
		t.Fatalf("got %d bonds, want %d", len(bonds), len(config.Game().Bonds.Bonds))
	}
	for _, b := range bonds {
		if len(b.CardNames) == 0 {
			t.Fatalf("bond %s has no cards", b.Name)
		}
	}
}
