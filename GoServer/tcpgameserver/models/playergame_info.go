package models

// ResponseInfo 响应信息结构体
type PlayerGameInfo struct {
	RoomId         string      `json:"Room_Id"`
	Username       string      `json:"Username"`
	Round          string      `json:"Round"`
	Health         float64     `json:"Health"`
	DamageDealt    float64     `json:"DamageDealt"`
	DamageReceived float64     `json:"DamageReceived"`
	TriggeredBonds []BondModel `json:"TriggeredBonds"`
	SelfCards      []Card      `json:"SelfCards"`
	OtherCards     []Card      `json:"OtherCards"`
}
