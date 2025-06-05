package models

// BondModel 羁绊模型，存储羁绊信息和关联的卡牌
type BondModel struct {
	ID          int      `json:"id"`          // 羁绊ID
	Name        string   `json:"name"`        // 羁绊名称
	Level       int      `json:"level"`       // 羁绊等级
	CardNames   []string `json:"cardnames"`   // 关联的卡牌列表
	Damage      float64  `json:"damage"`      // 羁绊伤害
	Description string   `json:"description"` // 羁绊描述
	Skill       string   `json:"skill"`       // 羁绊技能
}
