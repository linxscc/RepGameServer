package models

// BondModel 羁绊模型，存储羁绊信息和关联的卡牌
type BondModel struct {
	ID          int      `json:"ID"`          // 羁绊ID
	Name        string   `json:"Name"`        // 羁绊名称
	Level       int      `json:"level"`       // 羁绊等级
	CardNames   []string `json:"CardNames"`   // 关联的卡牌列表
	Damage      float64  `json:"Damage"`      // 羁绊伤害
	Description string   `json:"Description"` // 羁绊描述
	Skill       string   `json:"Skill"`       // 羁绊技能
}
