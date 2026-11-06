package logic

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
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
	if err := cards.InitCardPool(); err != nil {
		panic(err)
	}
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

// newTestRoom 创建一个对局中的双人房间：first 先手，两人手牌由 hands 指定
func newTestRoom(t *testing.T, pool []models.Card, first, second string, hands map[string][]models.Card) *types.RoomInfo {
	t.Helper()
	rm := service.GetRoomManager()
	room, err := rm.CreateRoom("test", 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := rm.InitCardPool(room.RoomID, pool, nil, nil); err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{first, second} {
		if err := rm.JoinRoom(room.RoomID, u); err != nil {
			t.Fatal(err)
		}
		if h, ok := hands[u]; ok {
			_ = room.SetPlayerHand(u, h)
		}
	}
	room.UpdateRoomStatus(types.RoomStatusPlaying)
	_ = room.SetPlayerRound(first, types.RoundCurrent)
	_ = room.SetPlayerRound(second, types.RoundWaiting)
	return room
}

func playerOf(t *testing.T, room *types.RoomInfo, name string) *types.PlayerInfo {
	t.Helper()
	p, err := room.GetPlayerInfo(name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// 集成测试：机器人按回合自动出牌，真人超时由服务器自动出牌，回合正确轮换
func TestTurnLoopWithBotAndForcedPlay(t *testing.T) {
	pool := make([]models.Card, 0, 40)
	for i := 0; i < 40; i++ {
		pool = append(pool, card(fmt.Sprintf("card_%d", i), "Fire", 1))
	}
	human, bot := "alice", NewBotName()
	room := newTestRoom(t, pool, bot, human, nil)
	if err := NewGameStartProcessor().DealInitialCardsToAllPlayers(room); err != nil {
		t.Fatal(err)
	}

	startHealth := playerOf(t, room, human).CurrentHealth
	if err := GlobalRoomTimerProcessor.StartRoomTimer(room.RoomID); err != nil {
		t.Fatal(err)
	}

	waitFor(t, "bot plays and passes the turn", func() bool {
		return room.CurrentPlayer() == human
	})
	if playerOf(t, room, human).CurrentHealth >= startHealth {
		t.Fatalf("bot play should damage the human")
	}

	// 真人回合超时：服务器替他打出第一张牌，随后又轮到机器人并自动出牌
	healthBeforeForced := playerOf(t, room, bot).CurrentHealth
	GlobalRoomTimerProcessor.forceCardPlay(room.RoomID, human)
	if playerOf(t, room, bot).CurrentHealth >= healthBeforeForced {
		t.Fatal("forced play should damage the bot")
	}
	waitFor(t, "bot takes its next turn automatically", func() bool {
		return room.CurrentPlayer() == human
	})

	GlobalRoomTimerProcessor.stopRoomTimer(room.RoomID)
}

// 不是自己的回合时出牌会被拒绝，房间状态不变
func TestPlayOutOfTurnIsRejected(t *testing.T) {
	room := newTestRoom(t, []models.Card{card("x2", "Fire", 1)}, "p2", "p1",
		map[string][]models.Card{"p1": {card("x1", "Fire", 1)}})

	PlayCards(PlayRequest{RoomID: room.RoomID, Player: "p1", CardUIDs: []string{"x1"}})

	if len(playerOf(t, room, "p1").HandCards) != 1 || room.CurrentPlayer() != "p2" {
		t.Fatal("out-of-turn play must not change room state")
	}
}

// 合成的目标卡池为空时，用于合成的牌必须留在手里
func TestComposeKeepsCardsWhenTargetPoolEmpty(t *testing.T) {
	target := "木匠" // cards.yaml 中真实存在的 2 级卡
	mk := func(uid string) models.Card {
		c := card(uid, "木匠学徒", 1)
		c.TargetName = &target
		return c
	}
	hand := []models.Card{mk("f1"), mk("f2"), mk("f3")}

	// 2 级卡池为空：合成失败，牌留在手里
	room := newTestRoom(t, nil, "p1", "p2", map[string][]models.Card{"p1": hand})
	ComposeCards(ComposeRequest{RoomID: room.RoomID, Player: "p1", CardUIDs: []string{"f1", "f2", "f3"}})
	if got := len(playerOf(t, room, "p1").HandCards); got != 3 {
		t.Fatalf("hand should be unchanged when the target pool is empty, got %d cards", got)
	}

	// 2 级卡池有目标卡：3 张换 1 张
	room = newTestRoom(t, nil, "p1", "p2", map[string][]models.Card{"p1": hand})
	_ = service.GetRoomManager().InitCardPool(room.RoomID, nil, []models.Card{{UID: "w1", Name: target, Level: 2}}, nil)
	ComposeCards(ComposeRequest{RoomID: room.RoomID, Player: "p1", CardUIDs: []string{"f1", "f2", "f3"}})
	if h := playerOf(t, room, "p1").HandCards; len(h) != 1 || h[0].Name != target {
		t.Fatalf("expected one composed %s, got %v", target, h)
	}
}

// 合成与出牌并发到达时串行结算：每张牌只会被用掉一次，回合只切换一次
func TestConcurrentPlayAndComposeAreSerialized(t *testing.T) {
	for i := 0; i < 50; i++ {
		target := "木匠"
		mk := func(uid string) models.Card {
			c := card(uid, "木匠学徒", 1)
			c.TargetName = &target
			return c
		}
		hand := []models.Card{mk("a"), mk("b"), mk("c")}
		room := newTestRoom(t, nil, "p1", "p2", map[string][]models.Card{"p1": hand})
		_ = service.GetRoomManager().InitCardPool(room.RoomID, nil, []models.Card{{UID: "w1", Name: target, Level: 2}}, nil)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			PlayCards(PlayRequest{RoomID: room.RoomID, Player: "p1", CardUIDs: []string{"a"}})
		}()
		go func() {
			defer wg.Done()
			ComposeCards(ComposeRequest{RoomID: room.RoomID, Player: "p1", CardUIDs: []string{"a", "b", "c"}})
		}()
		wg.Wait()
		GlobalRoomTimerProcessor.stopRoomTimer(room.RoomID)

		// 两种合法结果：先出牌（a 被打出，剩 2 张凑不满合成）或先合成（3 换 1，出牌因 a 已不在手里被拒）
		p1 := playerOf(t, room, "p1")
		played := playerOf(t, room, "p2").CurrentHealth < room.InitialHealth
		switch {
		case played && len(p1.HandCards) == 2 && room.CurrentPlayer() == "p2":
		case !played && len(p1.HandCards) == 1 && p1.HandCards[0].Name == target && room.CurrentPlayer() == "p1":
		default:
			t.Fatalf("iteration %d: inconsistent result: played=%v hand=%v current=%s", i, played, p1.HandCards, room.CurrentPlayer())
		}
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
