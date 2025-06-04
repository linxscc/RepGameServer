package logic

import (
	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/tools"
	"log"
)

// EventListener 事件监听器接口
type EventListener interface {
	GetName() string
	GetEventTypes() []string
	HandleEvent(eventType string, data interface{})
	GetPriority() int
}

// BaseEventListener 基础事件监听器
type BaseEventListener struct {
	Name       string
	EventTypes []string
	Priority   int
}

func (b *BaseEventListener) GetName() string {
	return b.Name
}

func (b *BaseEventListener) GetEventTypes() []string {
	return b.EventTypes
}

func (b *BaseEventListener) GetPriority() int {
	return b.Priority
}

// GameEventListener 游戏事件监听器
type GameEventListener struct {
	BaseEventListener
}

func NewGameEventListener() *GameEventListener {
	return &GameEventListener{
		BaseEventListener: BaseEventListener{
			Name: "GameEventListener",
			EventTypes: []string{
				events.EventGameStart,
				events.EventGameEnd,
				events.EventGamePause,
				events.EventGameResume,
			},
			Priority: 10, // 高优先级
		},
	}
}

func (g *GameEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventGameStart:
		g.handleGameStart(data)
	case events.EventGameEnd:
		g.handleGameEnd(data)
	case events.EventGamePause:
		g.handleGamePause(data)
	case events.EventGameResume:
		g.handleGameResume(data)
	default:
		log.Printf("GameEventListener: Unknown event type: %s", eventType)
	}
}

func (g *GameEventListener) handleGameStart(data interface{}) {
	log.Printf("🎮 Received game start event, processing directly")

	// 直接创建并使用GameStartProcessor处理游戏开始逻辑
	processor := &GameStartProcessor{}
	err := processor.ProcessGameStart(data)
	if err != nil {
		log.Printf("Game start processing failed: %v", err)
	} else {
		log.Printf("Game start processing completed successfully")
	}
}

func (g *GameEventListener) handleGameEnd(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🏁 Game Ended - Room: %s", eventData.RoomID)

		// 执行游戏结束逻辑
		if winner, exists := eventData.GetString("winner"); exists {
			log.Printf("Winner: %s", winner)
		}

		// 清理游戏状态
		// 计算积分
		// 保存游戏记录
	}
}

func (g *GameEventListener) handleGamePause(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("⏸️ Game Paused - Room: %s", eventData.RoomID)
		// 暂停游戏逻辑
	}
}

func (g *GameEventListener) handleGameResume(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("▶️ Game Resumed - Room: %s", eventData.RoomID)
		// 恢复游戏逻辑
	}
}

// CardEventListener 卡牌事件监听器
type CardEventListener struct {
	BaseEventListener
}

func NewCardEventListener() *CardEventListener {
	return &CardEventListener{
		BaseEventListener: BaseEventListener{
			Name: "CardEventListener",
			EventTypes: []string{
				events.EventCardDraw,
				events.EventCardPlay,
				events.EventCardDiscard,
				events.EventCardShuffle,
				events.EventDeckEmpty,
			},
			Priority: 30,
		},
	}
}

func (c *CardEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventCardDraw:
		c.handleCardDraw(data)
	case events.EventCardPlay:
		c.handleCardPlay(data)
	case events.EventDeckEmpty:
		c.handleDeckEmpty(data)
	default:
		log.Printf("CardEventListener: Unknown event type: %s", eventType)
	}
}

func (c *CardEventListener) handleCardDraw(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		cardName, _ := eventData.GetString("card_name")
		playerName, _ := eventData.GetString("player_name")
		log.Printf("🃏 Card Draw - %s drew %s", playerName, cardName)

		// 处理抽卡逻辑
		// 更新玩家手牌
		// 检查手牌上限
		// 触发抽卡特效
	}
}

func (c *CardEventListener) handleCardPlay(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		cardName, _ := eventData.GetString("card_name")
		playerName, _ := eventData.GetString("player_name")
		target, _ := eventData.GetString("target")

		log.Printf("🎯 Card Play - %s played %s on %s", playerName, cardName, target)

		// 处理出牌逻辑
		// 执行卡牌效果
		// 消耗资源
		// 移动卡牌到弃牌堆

		// 如果是攻击卡牌，触发伤害事件
		if damage, exists := eventData.GetFloat64("damage"); exists && damage > 0 {
			damageData := events.NewEventData(events.EventDamage, "card_system", map[string]interface{}{
				"target":   target,
				"damage":   damage,
				"source":   cardName,
				"attacker": playerName,
			})
			events.Publish(events.EventDamage, damageData)
		}
	}
}

func (c *CardEventListener) handleDeckEmpty(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("📭 Deck Empty - No more cards in room %s", eventData.RoomID)

		// 处理牌库为空
		// 将弃牌堆洗入牌库
		// 或者触发特殊规则

	}
}

// BattleEventListener 战斗事件监听器
type BattleEventListener struct {
	BaseEventListener
}

func NewBattleEventListener() *BattleEventListener {
	return &BattleEventListener{
		BaseEventListener: BaseEventListener{
			Name: "BattleEventListener",
			EventTypes: []string{
				events.EventBattleStart,
				events.EventBattleEnd,
				events.EventAttack,
				events.EventDamage,
				events.EventHeal,
			},
			Priority: 25,
		},
	}
}

func (b *BattleEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventBattleStart:
		b.handleBattleStart(data)
	case events.EventBattleEnd:
		b.handleBattleEnd(data)
	case events.EventAttack:
		b.handleAttack(data)
	case events.EventDamage:
		b.handleDamage(data)
	case events.EventHeal:
		b.handleHeal(data)
	default:
		log.Printf("BattleEventListener: Unknown event type: %s", eventType)
	}
}

func (b *BattleEventListener) handleBattleStart(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("⚔️ Battle Started in room %s", eventData.RoomID)

		// 初始化战斗状态
		// 设置回合顺序
		// 发送战斗开始消息
	}
}

func (b *BattleEventListener) handleBattleEnd(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		winner, _ := eventData.GetString("winner")
		log.Printf("🏆 Battle Ended - Winner: %s", winner)

		// 结算战斗结果
		// 发放奖励
		// 重置战斗状态
	}
}

func (b *BattleEventListener) handleAttack(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		attacker, _ := eventData.GetString("attacker")
		target, _ := eventData.GetString("target")
		log.Printf("⚔️ Attack - %s attacks %s", attacker, target)

		// 处理攻击逻辑
		// 计算伤害
		// 检查防御
	}
}

func (b *BattleEventListener) handleDamage(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		target, _ := eventData.GetString("target")
		damage, _ := eventData.GetFloat64("damage")
		source, _ := eventData.GetString("source")

		log.Printf("💥 Damage - %s takes %.1f damage from %s", target, damage, source)

		// 应用伤害
		// 检查死亡条件
		// 触发伤害效果

		// 检查是否死亡
		if currentHP, exists := eventData.GetFloat64("current_hp"); exists && currentHP <= 0 {
			deathData := events.NewEventData(events.EventPlayerDeath, "battle_system", map[string]interface{}{
				"player_name": target,
				"killer":      eventData.Data["attacker"],
			})
			deathData.SetRoom(eventData.RoomID).SetUser(target)
			events.Publish(events.EventPlayerDeath, deathData)
		}
	}
}

func (b *BattleEventListener) handleHeal(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		target, _ := eventData.GetString("target")
		healAmount, _ := eventData.GetFloat64("heal_amount")
		source, _ := eventData.GetString("source")

		log.Printf("💚 Heal - %s heals %.1f HP from %s", target, healAmount, source)

		// 应用治疗
		// 检查HP上限
		// 触发治疗效果
	}
}

// SystemEventListener 系统事件监听器
type SystemEventListener struct {
	BaseEventListener
}

func NewSystemEventListener() *SystemEventListener {
	return &SystemEventListener{
		BaseEventListener: BaseEventListener{
			Name: "SystemEventListener",
			EventTypes: []string{
				events.EventSystemStart,
				events.EventSystemShutdown,
				events.EventSystemError,
				events.EventServerMaintenance,
			},
			Priority: 5, // 最高优先级
		},
	}
}

func (s *SystemEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventSystemStart:
		s.handleSystemStart(data)
	case events.EventSystemShutdown:
		s.handleSystemShutdown(data)
	case events.EventSystemError:
		s.handleSystemError(data)
	case events.EventServerMaintenance:
		s.handleServerMaintenance(data)
	default:
		log.Printf("SystemEventListener: Unknown event type: %s", eventType)
	}
}

func (s *SystemEventListener) handleSystemStart(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🚀 System Start - %s", eventData.Data["message"])

		// 系统启动逻辑
		// 初始化全局状态
		// 加载配置
		// 启动服务

		// 加载响应码配置文件
		if err := tools.LoadResponseCodes(); err != nil {
			log.Printf("Failed to load response codes from database: %v", err)
		} else {
			log.Println("Response codes loaded successfully")
		}

		// 初始化卡牌池
		if err := cards.InitCardPool(); err != nil {
			log.Printf("Failed to initialize card pool: %v", err)
		} else {
			log.Println("Card pool initialized successfully")
		}

	}
}

func (s *SystemEventListener) handleSystemShutdown(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🔴 System Shutdown - %s", eventData.Data["message"])

		// 系统关闭逻辑
		// 保存数据
		// 断开连接
		// 清理资源
	}
}

func (s *SystemEventListener) handleSystemError(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		errorMsg, _ := eventData.GetString("error")
		severity, _ := eventData.GetString("severity")

		log.Printf("❌ System Error [%s] - %s", severity, errorMsg)

		// 错误处理逻辑
		// 记录错误日志
		// 发送告警
		// 尝试恢复
	}
}

func (s *SystemEventListener) handleServerMaintenance(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		maintenanceType, _ := eventData.GetString("type")
		log.Printf("🔧 Server Maintenance - %s", maintenanceType)

		// 维护逻辑
		// 通知玩家
		// 暂停新连接
		// 执行维护任务
	}
}

// RoomEventListener 房间事件监听器
type RoomEventListener struct {
	BaseEventListener
}

func NewRoomEventListener() *RoomEventListener {
	return &RoomEventListener{
		BaseEventListener: BaseEventListener{
			Name: "RoomEventListener",
			EventTypes: []string{
				events.EventRoomCreate,
				events.EventRoomDestroy,
				events.EventRoomFull,
				events.EventRoomEmpty,
			},
			Priority: 15, // 中等优先级
		},
	}
}

func (r *RoomEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventRoomCreate:
		r.handleRoomCreate(data)
	case events.EventRoomDestroy:
		r.handleRoomDestroy(data)
	case events.EventRoomFull:
		r.handleRoomFull(data)
	case events.EventRoomEmpty:
		r.handleRoomEmpty(data)
	default:
		log.Printf("RoomEventListener: Unknown event type: %s", eventType)
	}
}

func (r *RoomEventListener) handleRoomCreate(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🏠 Room Create Event - Room: %s", eventData.RoomID)

	}
}

func (r *RoomEventListener) handleRoomDestroy(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🏚️ Room Destroy - Room: %s", eventData.RoomID)

		// 处理房间销毁逻辑
		// 清理房间资源
		// 通知相关玩家
		// 保存房间数据
	}
}

func (r *RoomEventListener) handleRoomFull(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🔒 Room Full - Room: %s", eventData.RoomID)

		// 处理房间已满逻辑
		// 拒绝新玩家加入
		// 可能触发游戏开始
	}
}

func (r *RoomEventListener) handleRoomEmpty(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		log.Printf("🕳️ Room Empty - Room: %s", eventData.RoomID)

		// 处理房间为空逻辑
		// 准备销毁房间
		destroyData := events.CreateRoomEventData(events.EventRoomDestroy, eventData.RoomID, 0)
		events.Publish(events.EventRoomDestroy, destroyData)
	}
}

// ListenerManager 监听器管理器
type ListenerManager struct {
	listeners       []EventListener
	subscriptionIDs map[string][]string // listener name -> subscription IDs
}

func NewListenerManager() *ListenerManager {
	return &ListenerManager{
		listeners:       make([]EventListener, 0),
		subscriptionIDs: make(map[string][]string),
	}
}

// RegisterListener 注册监听器
func (lm *ListenerManager) RegisterListener(listener EventListener) {
	lm.listeners = append(lm.listeners, listener)

	// 为监听器订阅所有相关事件
	subscriptionIDs := make([]string, 0)
	for _, eventType := range listener.GetEventTypes() {
		subscriptionID := events.Subscribe(eventType, func(data interface{}) {
			listener.HandleEvent(eventType, data)
		}, listener.GetPriority())

		subscriptionIDs = append(subscriptionIDs, subscriptionID)
	}

	lm.subscriptionIDs[listener.GetName()] = subscriptionIDs
	log.Printf("Registered event listener: %s for %d event types",
		listener.GetName(), len(listener.GetEventTypes()))
}

// UnregisterListener 注销监听器
func (lm *ListenerManager) UnregisterListener(listenerName string) bool {
	if subscriptionIDs, exists := lm.subscriptionIDs[listenerName]; exists {
		// 取消所有订阅
		for _, subscriptionID := range subscriptionIDs {
			events.Unsubscribe(subscriptionID)
		}

		// 从列表中移除
		for i, listener := range lm.listeners {
			if listener.GetName() == listenerName {
				lm.listeners = append(lm.listeners[:i], lm.listeners[i+1:]...)
				break
			}
		}

		delete(lm.subscriptionIDs, listenerName)
		log.Printf("Unregistered event listener: %s", listenerName)
		return true
	}
	return false
}

// GetListeners 获取所有监听器
func (lm *ListenerManager) GetListeners() []EventListener {
	return lm.listeners
}

// GetListenerCount 获取监听器数量
func (lm *ListenerManager) GetListenerCount() int {
	return len(lm.listeners)
}

// RegisterAllDefaultListeners 注册所有默认监听器
func (lm *ListenerManager) RegisterAllDefaultListeners() {
	lm.RegisterListener(NewSystemEventListener())
	lm.RegisterListener(NewGameEventListener())
	lm.RegisterListener(NewBattleEventListener())
	lm.RegisterListener(NewCardEventListener())
	lm.RegisterListener(NewRoomEventListener())

	log.Printf("Registered %d default event listeners", lm.GetListenerCount())
}

// 全局监听器管理器实例
var globalListenerManager *ListenerManager

// GetListenerManager 获取全局监听器管理器
func GetListenerManager() *ListenerManager {
	if globalListenerManager == nil {
		globalListenerManager = NewListenerManager()
	}
	return globalListenerManager
}

// InitializeEventSystem 初始化事件系统
func InitializeEventSystem() {
	log.Println("Initializing event system...")

	// 获取监听器管理器并注册默认监听器
	listenerManager := GetListenerManager()
	listenerManager.RegisterAllDefaultListeners()

	// 发布系统启动事件
	systemStartData := events.CreateSystemEventData(events.EventSystemStart, "Event system initialized successfully")
	events.Publish(events.EventSystemStart, systemStartData)

	log.Printf("Event system initialized with %d listeners", listenerManager.GetListenerCount())
}

// ShutdownEventSystem 关闭事件系统
func ShutdownEventSystem() {
	log.Println("Shutting down event system...")

	// 发布系统关闭事件
	systemShutdownData := events.CreateSystemEventData(events.EventSystemShutdown, "Event system shutting down")
	events.PublishSync(events.EventSystemShutdown, systemShutdownData) // 同步发布，确保处理完成

	// 清空所有订阅
	events.Clear()

	log.Println("Event system shutdown complete")
}
