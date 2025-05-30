package models

type HealthUpdate struct {
	AttackerHealth int `json:"AttackerHealth"`
	ReceiverHealth int `json:"ReceiverHealth"`
}
