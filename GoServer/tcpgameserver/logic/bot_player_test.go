package logic

import (
	"os"
	"path/filepath"
	"testing"

	"GoServer/tcpgameserver/cards"
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
	if err := cards.InitCardPool(); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
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
