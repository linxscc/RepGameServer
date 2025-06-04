package events

// 预定义的事件类型常量
const (
	// 游戏相关事件
	EventGameStart  = "game.start"  // 游戏开始
	EventGameEnd    = "game.end"    // 游戏结束
	EventGamePause  = "game.pause"  // 游戏暂停
	EventGameResume = "game.resume" // 游戏恢复
	EventGameReset  = "game.reset"  // 游戏重置

	// 玩家相关事件
	EventPlayerJoin   = "player.join"   // 玩家加入
	EventPlayerLeave  = "player.leave"  // 玩家离开
	EventPlayerMove   = "player.move"   // 玩家移动
	EventPlayerAction = "player.action" // 玩家行动
	EventPlayerDeath  = "player.death"  // 玩家死亡
	EventPlayerRevive = "player.revive" // 玩家复活

	// 卡牌相关事件
	EventCardDraw    = "card.draw"    // 抽卡
	EventCardPlay    = "card.play"    // 出牌
	EventCardDiscard = "card.discard" // 弃牌
	EventCardShuffle = "card.shuffle" // 洗牌
	EventDeckEmpty   = "deck.empty"   // 牌库为空

	// 战斗相关事件
	EventBattleStart = "battle.start"   // 战斗开始
	EventBattleEnd   = "battle.end"     // 战斗结束
	EventAttack      = "battle.attack"  // 攻击
	EventDefense     = "battle.defense" // 防御
	EventDamage      = "battle.damage"  // 造成伤害
	EventHeal        = "battle.heal"    // 治疗

	// 房间相关事件
	EventRoomCreate  = "room.create"  // 房间创建
	EventRoomDestroy = "room.destroy" // 房间销毁
	EventRoomFull    = "room.full"    // 房间已满
	EventRoomEmpty   = "room.empty"   // 房间为空

	// 连接相关事件
	EventClientConnect    = "client.connect"    // 客户端连接
	EventClientDisconnect = "client.disconnect" // 客户端断开
	EventClientTimeout    = "client.timeout"    // 客户端超时

	// 系统相关事件
	EventSystemStart       = "system.start"       // 系统启动
	EventSystemShutdown    = "system.shutdown"    // 系统关闭
	EventSystemError       = "system.error"       // 系统错误
	EventServerMaintenance = "server.maintenance" // 服务器维护

	// 数据相关事件
	EventDataSave    = "data.save"    // 数据保存
	EventDataLoad    = "data.load"    // 数据加载
	EventDataBackup  = "data.backup"  // 数据备份
	EventDataRestore = "data.restore" // 数据恢复
)

// EventData 通用事件数据结构
type EventData struct {
	Type      string                 `json:"type"`      // 事件类型
	Source    string                 `json:"source"`    // 事件源
	Timestamp int64                  `json:"timestamp"` // 时间戳
	Data      map[string]interface{} `json:"data"`      // 事件数据
	UserID    string                 `json:"user_id"`   // 用户ID（可选）
	RoomID    string                 `json:"room_id"`   // 房间ID（可选）
}

// NewEventData 创建新的事件数据
func NewEventData(eventType, source string, data map[string]interface{}) *EventData {
	return &EventData{
		Type:      eventType,
		Source:    source,
		Timestamp: getCurrentTimestamp(),
		Data:      data,
	}
}

// SetUser 设置用户ID
func (ed *EventData) SetUser(userID string) *EventData {
	ed.UserID = userID
	return ed
}

// SetRoom 设置房间ID
func (ed *EventData) SetRoom(roomID string) *EventData {
	ed.RoomID = roomID
	return ed
}

// AddData 添加数据
func (ed *EventData) AddData(key string, value interface{}) *EventData {
	if ed.Data == nil {
		ed.Data = make(map[string]interface{})
	}
	ed.Data[key] = value
	return ed
}

// GetData 获取数据
func (ed *EventData) GetData(key string) (interface{}, bool) {
	if ed.Data == nil {
		return nil, false
	}
	value, exists := ed.Data[key]
	return value, exists
}

// GetString 获取字符串数据
func (ed *EventData) GetString(key string) (string, bool) {
	if value, exists := ed.GetData(key); exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt 获取整数数据
func (ed *EventData) GetInt(key string) (int, bool) {
	if value, exists := ed.GetData(key); exists {
		if num, ok := value.(int); ok {
			return num, true
		}
	}
	return 0, false
}

// GetFloat64 获取浮点数数据
func (ed *EventData) GetFloat64(key string) (float64, bool) {
	if value, exists := ed.GetData(key); exists {
		if num, ok := value.(float64); ok {
			return num, true
		}
	}
	return 0.0, false
}

// GetBool 获取布尔数据
func (ed *EventData) GetBool(key string) (bool, bool) {
	if value, exists := ed.GetData(key); exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}
