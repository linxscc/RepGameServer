package models

// ResponseInfo 响应信息结构体
type PlayerGameInfo struct {
	RoomId       string                `json:"Room_Id"`
	Username     string                `json:"Username"`
	Round        string                `json:"Round"`
	Health       float64               `json:"Health"`
	SelfCards    []Card                `json:"SelfCards"`
	OtherPlayers []OtherPlayerGameInfo `json:"OtherPlayers"`
	DamageInfo   []DamageInfo          `json:"DamageInfo"`
}

type OtherPlayerGameInfo struct {
	Username   string  `json:"Username"`
	Round      string  `json:"Round"`
	Health     float64 `json:"Health"`
	CardsCount int     `json:"CardsCount"`
}

type DamageInfo struct {
	DamageSource   string      `json:"DamageSource"`
	DamageTarget   string      `json:"DamageTarget"`
	DamageType     string      `json:"DamageType"`
	DamageValue    float64     `json:"DamageValue"`
	TriggeredBonds []BondModel `json:"TriggeredBonds"`
}
