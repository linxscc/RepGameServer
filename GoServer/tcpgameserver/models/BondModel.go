package models

// Card 卡牌模型，存储在共享池中的值类型
type BondModel struct {
	Name        string  `json:"Name"`
	Cards       []Card  `json:"Cards"`
	Damage      float64 `json:"Damage"`
	Description string  `json:"Description"`
	Skill       string  `json:"Skill"`
}
