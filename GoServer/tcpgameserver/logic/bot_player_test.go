package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
)

// 测试使用较短的机器人思考时间，其余配置沿用内置配置表
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "repgame-config")
	if err != nil {
		panic(err)
	}
	bot := "enabled: true\nmatch_wait_seconds: 3\nname_prefix: \"机器人-\"\nthink_delay_ms: 20\nmax_cards_per_play: 3\n"
	if err := os.WriteFile(filepath.Join(dir, "bot.yaml"), []byte(bot), 0o644); err != nil {
		panic(err)
	}
	os.Setenv("GAME_CONFIG_DIR", dir)
	code := m.Run()
	CleanupTimers()
	os.RemoveAll(dir)
	os.Exit(code)
}

func card(uid, name string, dmg float64) models.Card {
	return models.Card{UID: uid, Name: name, Damage: dmg, Level: 1}
}

// 不考虑羁绊时，伤害就是卡牌伤害之和
func sumScore(cards []models.Card) BondCalculationResult {
	total := 0.0
	for _, c := range cards {
		total += c.Damage
	}
	return BondCalculationResult{TotalDamage: total}
}

func TestChooseBotPlayPrefersHighestDamageWithinLimit(t *testing.T) {
	hand := []models.Card{card("a", "Fire", 1), card("b", "Water", 5), card("c", "Fire", 1), card("d", "Wind", 4), card("e", "Wind", 3)}

	got := ChooseBotPlay(hand, 3, sumScore)
	if len(got) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(got))
	}
	total := sumScore(got).TotalDamage
	if total != 12 { // 5 + 4 + 3
		t.Fatalf("expected best total 12, got %v (%v)", total, got)
	}
}

func TestChooseBotPlayRespectsMaxCards(t *testing.T) {
	hand := []models.Card{card("a", "Fire", 2), card("b", "Fire", 2), card("c", "Fire", 2), card("d", "Fire", 2)}
	if got := ChooseBotPlay(hand, 1, sumScore); len(got) != 1 {
		t.Fatalf("max_cards_per_play=1 must play a single card, got %d", len(got))
	}
}

func TestChooseBotPlayUsesBondScore(t *testing.T) {
	// 羁绊让两张“木匠”的组合比单张高伤害卡更强
	hand := []models.Card{card("a", "木匠", 1), card("b", "木匠", 1), card("c", "铁匠", 3)}
	bondScore := func(cards []models.Card) BondCalculationResult {
		n := 0
		for _, c := range cards {
			if c.Name == "木匠" {
				n++
			}
		}
		if n == 2 && len(cards) == 2 {
			return BondCalculationResult{TotalDamage: 10}
		}
		return sumScore(cards)
	}
	got := ChooseBotPlay(hand, 3, bondScore)
	if len(got) != 2 || got[0].Name != "木匠" {
		t.Fatalf("expected the bonded pair, got %v", got)
	}
}

func TestBotNamesAreReservedAndUnique(t *testing.T) {
	a, b := NewBotName(), NewBotName()
	if a == b {
		t.Fatal("bot names must be unique")
	}
	if !IsBot(a) || !service.IsReservedUsername(a) || IsBot("alice") {
		t.Fatal("bot detection / reserved prefix mismatch")
	}
}

// 集成测试：机器人按回合自动出牌，真人超时由服务器自动出牌，回合正确轮换
func TestTurnLoopWithBotAndForcedPlay(t *testing.T) {
	events.Subscribe(events.EventCardPlay, func(data interface{}) {
		if ed, ok := data.(*events.EventData); ok {
			NewPlayCardProcessor().ProcessPlayCard(ed)
		}
	})

	rm := service.GetRoomManager()
	room, err := rm.CreateRoom("test", config.Game().Rules.Match.PlayersPerRoom)
	if err != nil {
		t.Fatal(err)
	}
	pool := make([]models.Card, 0, 40)
	for i := 0; i < 40; i++ {
		pool = append(pool, card(fmt.Sprintf("card_%d", i), "Fire", 1))
	}
	if err := rm.InitCardPool(room.RoomID, pool, nil, nil); err != nil {
		t.Fatal(err)
	}

	human, bot := "alice", NewBotName()
	for _, u := range []string{human, bot} {
		if err := rm.JoinRoom(room.RoomID, u); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewGameStartProcessor().DealInitialCardsToAllPlayers(room); err != nil {
		t.Fatal(err)
	}
	room.Status = "playing"
	rm.SetPlayerRound(room.RoomID, bot, "current")
	rm.SetPlayerRound(room.RoomID, human, "waiting")

	startHealth := room.Players[human].CurrentHealth
	if err := GlobalRoomTimerProcessor.StartRoomTimer(room.RoomID); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "bot plays and passes the turn", func() bool {
		return room.Players[human].Round == "current" && room.Players[bot].Round == "waiting"
	})
	if room.Players[human].CurrentHealth >= startHealth {
		t.Fatalf("bot play should damage the human: %v -> %v", startHealth, room.Players[human].CurrentHealth)
	}

	// 真人回合超时：服务器替他打出第一张牌，随后又轮到机器人并自动出牌
	healthBeforeForced := room.Players[bot].CurrentHealth
	GlobalRoomTimerProcessor.forceCardPlay(room.RoomID, human)
	waitFor(t, "forced play damages the bot", func() bool {
		return room.Players[bot].CurrentHealth < healthBeforeForced
	})
	waitFor(t, "bot takes its next turn automatically", func() bool {
		return room.Players[human].Round == "current"
	})

	GlobalRoomTimerProcessor.stopRoomTimer(room.RoomID)
}

// 不是自己的回合时出牌会被拒绝，房间状态不变
func TestPlayOutOfTurnIsRejected(t *testing.T) {
	rm := service.GetRoomManager()
	room, _ := rm.CreateRoom("turn-check", 2)
	_ = rm.InitCardPool(room.RoomID, []models.Card{card("x1", "Fire", 1), card("x2", "Fire", 1)}, nil, nil)
	_ = rm.JoinRoom(room.RoomID, "p1")
	_ = rm.JoinRoom(room.RoomID, "p2")
	room.Status = "playing"
	room.Players["p1"].HandCards = []models.Card{card("x1", "Fire", 1)}
	rm.SetPlayerRound(room.RoomID, "p1", "waiting")
	rm.SetPlayerRound(room.RoomID, "p2", "current")

	ed := events.NewEventData(events.EventCardPlay, "test", map[string]interface{}{
		"player": "p1", "room_id": room.RoomID, "self_cards": []models.Card{{UID: "x1"}},
	})
	NewPlayCardProcessor().ProcessPlayCard(ed)

	if len(room.Players["p1"].HandCards) != 1 || room.Players["p2"].Round != "current" {
		t.Fatal("out-of-turn play must not change room state")
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", what)
}
