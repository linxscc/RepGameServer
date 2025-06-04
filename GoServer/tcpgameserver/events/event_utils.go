package events

import (
	"log"
	"time"
)

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// LogEventData 记录事件数据到日志
func LogEventData(eventType string, data interface{}) {
	log.Printf("Event [%s]: %v", eventType, data)
}

// CreatePlayerEventData 创建玩家相关事件数据
func CreatePlayerEventData(eventType, playerID, playerName string) *EventData {
	data := make(map[string]interface{})
	data["player_id"] = playerID
	data["player_name"] = playerName

	return NewEventData(eventType, "player_manager", data).SetUser(playerID)
}

// CreateCardEventData 创建卡牌相关事件数据
func CreateCardEventData(eventType, cardName string, damage float64, level int) *EventData {
	data := make(map[string]interface{})
	data["card_name"] = cardName
	data["damage"] = damage
	data["level"] = level

	return NewEventData(eventType, "card_manager", data)
}

// CreateRoomEventData 创建房间相关事件数据
func CreateRoomEventData(eventType, roomID string, playerCount int) *EventData {
	data := make(map[string]interface{})
	data["room_id"] = roomID
	data["player_count"] = playerCount

	return NewEventData(eventType, "room_manager", data).SetRoom(roomID)
}

// CreateSystemEventData 创建系统相关事件数据
func CreateSystemEventData(eventType, message string) *EventData {
	data := make(map[string]interface{})
	data["message"] = message
	data["server_time"] = time.Now().Format("2006-01-02 15:04:05")

	return NewEventData(eventType, "system", data)
}
