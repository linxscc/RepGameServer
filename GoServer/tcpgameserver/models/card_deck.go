package models

// CardDeck 卡组结构体
type CardDeck struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	CardsNum   int     `json:"cards_num"`
	Damage     float64 `json:"damage"`
	TargetName *string `json:"targetname,omitempty"`
	Level      int     `json:"level"`
}
