package tools

import (
	"GoServer/tcpgameserver/models"
	"encoding/json"
	"log"
)

func ConvertJsontoList_Health(healthJson interface{}) ([]models.HealthUpdate, error) {
	var healthUpdates []models.HealthUpdate
	dataBytes, _ := json.Marshal(healthJson)
	err := json.Unmarshal([]byte(dataBytes), &healthUpdates)
	if err != nil {
		log.Printf("Failed to parse health updates: %v", err)
		return nil, err
	}
	return healthUpdates, nil
}
