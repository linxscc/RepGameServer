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

func TestCardAndBondTables(t *testing.T) {
	tables, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(tables.Cards.Cards) != 31 || len(tables.Bonds.Bonds) != 12 {
		t.Fatalf("unexpected table sizes: %d cards, %d bonds", len(tables.Cards.Cards), len(tables.Bonds.Bonds))
	}
	for _, c := range tables.Cards.Cards {
		if c.Name == "Fire" || c.Name == "Water" || c.Name == "Wind" {
			t.Fatalf("placeholder card %q must not be in the card table", c.Name)
		}
	}
}

func TestCardTableRejectsBrokenReferences(t *testing.T) {
	cards := CardTable{Cards: []CardDef{
		{ID: 1, Name: "A", Count: 3, Level: 1, Target: "Missing"},
		{ID: 2, Name: "B", Count: 1, Level: 2},
		{ID: 3, Name: "C", Count: 1, Level: 3},
	}}
	bonds := BondTable{Bonds: []BondDef{{ID: 1, Name: "X", Cards: []string{"A", "Nope"}}}}
	err := validateCardData(cards, bonds)
	if err == nil || !strings.Contains(err.Error(), "Missing") || !strings.Contains(err.Error(), "Nope") {
		t.Fatalf("expected broken target and bond reference errors, got %v", err)
	}

	cards.Cards[0].Target = "C" // 等级跳级（1 -> 3）
	if err := validateCardData(cards, BondTable{}); err == nil || !strings.Contains(err.Error(), "level 2") {
		t.Fatalf("expected level-skip error, got %v", err)
	}
}
