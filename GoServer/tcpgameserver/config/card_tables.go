package config

import (
	"errors"
	"fmt"
)

// CardDef 卡牌配置（cards.yaml）
type CardDef struct {
	ID     int     `yaml:"id"`     // 卡牌编号
	Name   string  `yaml:"name"`   // 卡牌名称（与客户端预制体同名）
	Count  int     `yaml:"count"`  // 每个房间牌池中的数量
	Damage float64 `yaml:"damage"` // 基础伤害
	Level  int     `yaml:"level"`  // 等级 1~3
	Target string  `yaml:"target"` // 合成目标卡牌名称，空表示不能合成
}

// CardTable cards.yaml 的结构
type CardTable struct {
	Cards []CardDef `yaml:"cards"`
}

// BondDef 羁绊配置（bonds.yaml）
type BondDef struct {
	ID          int      `yaml:"id"`          // 羁绊编号
	Name        string   `yaml:"name"`        // 羁绊名称
	Level       int      `yaml:"level"`       // 羁绊等级
	Damage      float64  `yaml:"damage"`      // 触发后的伤害
	Skill       string   `yaml:"skill"`       // 技能标识（预留）
	Description string   `yaml:"description"` // 说明文字
	Cards       []string `yaml:"cards"`       // 组成羁绊的卡牌名称
}

// BondTable bonds.yaml 的结构
type BondTable struct {
	Bonds []BondDef `yaml:"bonds"`
}

const (
	minCardLevel    = 1
	maxCardLevel    = 3
	maxCardsPerBond = 7
)

// validateCardData 校验卡牌与羁绊表之间的引用关系
func validateCardData(cards CardTable, bonds BondTable) error {
	var errs []error
	add := func(format string, args ...interface{}) { errs = append(errs, fmt.Errorf(format, args...)) }

	byName := make(map[string]CardDef, len(cards.Cards))
	ids := make(map[int]bool, len(cards.Cards))
	poolSize := make(map[int]int)
	for _, c := range cards.Cards {
		if c.Name == "" {
			add("cards.yaml: card id %d has empty name", c.ID)
			continue
		}
		if _, dup := byName[c.Name]; dup {
			add("cards.yaml: duplicate card name %q", c.Name)
		}
		if ids[c.ID] {
			add("cards.yaml: duplicate card id %d", c.ID)
		}
		byName[c.Name], ids[c.ID] = c, true

		if c.Level < minCardLevel || c.Level > maxCardLevel {
			add("cards.yaml: %s level must be %d..%d", c.Name, minCardLevel, maxCardLevel)
		}
		if c.Count <= 0 {
			add("cards.yaml: %s count must be > 0", c.Name)
		}
		if c.Damage < 0 {
			add("cards.yaml: %s damage must be >= 0", c.Name)
		}
		poolSize[c.Level] += c.Count
	}

	for _, c := range cards.Cards {
		if c.Target == "" {
			continue
		}
		target, ok := byName[c.Target]
		if !ok {
			add("cards.yaml: %s compose target %q does not exist", c.Name, c.Target)
		} else if target.Level != c.Level+1 {
			add("cards.yaml: %s (level %d) must compose into a level %d card, %s is level %d",
				c.Name, c.Level, c.Level+1, target.Name, target.Level)
		}
	}
	for level := minCardLevel; level <= maxCardLevel; level++ {
		if poolSize[level] == 0 {
			add("cards.yaml: no cards of level %d", level)
		}
	}

	bondIDs := make(map[int]bool, len(bonds.Bonds))
	for _, b := range bonds.Bonds {
		if bondIDs[b.ID] {
			add("bonds.yaml: duplicate bond id %d", b.ID)
		}
		bondIDs[b.ID] = true
		if len(b.Cards) == 0 || len(b.Cards) > maxCardsPerBond {
			add("bonds.yaml: %s must have 1..%d cards", b.Name, maxCardsPerBond)
		}
		for _, name := range b.Cards {
			if _, ok := byName[name]; !ok {
				add("bonds.yaml: %s references unknown card %q", b.Name, name)
			}
		}
	}
	return errors.Join(errs...)
}
