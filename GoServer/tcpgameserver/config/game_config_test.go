package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedTablesLoadAndValidate(t *testing.T) {
	tables, err := Load("")
	if err != nil {
		t.Fatalf("embedded config tables are invalid: %v", err)
	}
	if tables.Rules.Player.InitialHealth <= 0 || tables.Bot.MatchWaitSeconds != 3 {
		t.Fatalf("unexpected values: %+v %+v", tables.Rules.Player, tables.Bot)
	}
	if got := tables.Messages.Text(MsgComposeCountInvalid, 3); !strings.Contains(got, "3") {
		t.Fatalf("formatting failed: %q", got)
	}
}

func TestOverrideDirTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	override := "enabled: false\nmatch_wait_seconds: 9\nname_prefix: \"Bot-\"\nthink_delay_ms: 0\nmax_cards_per_play: 2\n"
	if err := os.WriteFile(filepath.Join(dir, "bot.yaml"), []byte(override), 0o644); err != nil {
		t.Fatal(err)
	}
	tables, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if tables.Bot.Enabled || tables.Bot.MatchWaitSeconds != 9 {
		t.Fatalf("override not applied: %+v", tables.Bot)
	}
	// 未覆盖的表仍使用内置配置
	if tables.Rules.Player.InitialHandSize == 0 {
		t.Fatal("embedded game_rules should still load")
	}
}

func TestUnknownFieldIsRejected(t *testing.T) {
	dir := t.TempDir()
	bad := "enabled: true\nmatch_wait_secs: 3\nname_prefix: \"Bot-\"\nthink_delay_ms: 0\nmax_cards_per_play: 2\n"
	_ = os.WriteFile(filepath.Join(dir, "bot.yaml"), []byte(bad), 0o644)
	if _, err := Load(dir); err == nil {
		t.Fatal("misspelled field must be rejected")
	}
}

func TestInvalidValueIsRejected(t *testing.T) {
	dir := t.TempDir()
	bad := "enabled: true\nmatch_wait_seconds: 3\nname_prefix: \"\"\nthink_delay_ms: 0\nmax_cards_per_play: 0\n"
	_ = os.WriteFile(filepath.Join(dir, "bot.yaml"), []byte(bad), 0o644)
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "name_prefix") || !strings.Contains(err.Error(), "max_cards_per_play") {
		t.Fatalf("expected validation errors, got %v", err)
	}
}
