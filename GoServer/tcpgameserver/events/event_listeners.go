package events

import (
	"log"
	"time"
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
				EventGameStart,
				EventGameEnd,
				EventGamePause,
				EventGameResume,
			},
			Priority: 10, // 高优先级
		},
	}
}

func (g *GameEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case EventGameStart:
		g.handleGameStart(data)
	case EventGameEnd:
		g.handleGameEnd(data)
	case EventGamePause:
		g.handleGamePause(data)
	case EventGameResume:
		g.handleGameResume(data)
	default:
		log.Printf("GameEventListener: Unknown event type: %s", eventType)
	}
}

func (g *GameEventListener) handleGameStart(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("🎮 Game Started - Room: %s, Time: %s",
			eventData.RoomID,
			time.Unix(eventData.Timestamp, 0).Format("15:04:05"))

		// 执行游戏开始逻辑
		if roomID := eventData.RoomID; roomID != "" {
			// 初始化游戏状态
			// 发送游戏开始消息给所有玩家
			log.Printf("Initializing game state for room: %s", roomID)
		}
	}
}

func (g *GameEventListener) handleGameEnd(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
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
	if eventData, ok := data.(*EventData); ok {
		log.Printf("⏸️ Game Paused - Room: %s", eventData.RoomID)
		// 暂停游戏逻辑
	}
}

func (g *GameEventListener) handleGameResume(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("▶️ Game Resumed - Room: %s", eventData.RoomID)
		// 恢复游戏逻辑
	}
}

// PlayerEventListener 玩家事件监听器
type PlayerEventListener struct {
	BaseEventListener
}

func NewPlayerEventListener() *PlayerEventListener {
	return &PlayerEventListener{
		BaseEventListener: BaseEventListener{
			Name: "PlayerEventListener",
			EventTypes: []string{
				EventPlayerJoin,
				EventPlayerLeave,
				EventPlayerAction,
				EventPlayerDeath,
			},
			Priority: 20,
		},
	}
}

func (p *PlayerEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case EventPlayerJoin:
		p.handlePlayerJoin(data)
	case EventPlayerLeave:
		p.handlePlayerLeave(data)
	case EventPlayerAction:
		p.handlePlayerAction(data)
	case EventPlayerDeath:
		p.handlePlayerDeath(data)
	default:
		log.Printf("PlayerEventListener: Unknown event type: %s", eventType)
	}
}

func (p *PlayerEventListener) handlePlayerJoin(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		playerName, _ := eventData.GetString("player_name")
		log.Printf("👤 Player Joined - %s in room %s", playerName, eventData.RoomID)

		// 执行玩家加入逻辑
		// 更新房间玩家列表
		// 发送欢迎消息
		// 同步游戏状态给新玩家

		// 触发相关事件
		if playerCount, exists := eventData.GetInt("player_count"); exists && playerCount >= 2 {
			// 如果房间人数满足条件，可以触发游戏开始事件
			gameStartData := CreateRoomEventData(EventGameStart, eventData.RoomID, playerCount)
			Publish(EventGameStart, gameStartData)
		}
	}
}

func (p *PlayerEventListener) handlePlayerLeave(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		playerName, _ := eventData.GetString("player_name")
		log.Printf("👋 Player Left - %s from room %s", playerName, eventData.RoomID)

		// 执行玩家离开逻辑
		// 更新房间玩家列表
		// 检查是否需要暂停游戏
		// 如果房间为空，触发房间销毁事件

		if playerCount, exists := eventData.GetInt("remaining_players"); exists && playerCount == 0 {
			roomDestroyData := CreateRoomEventData(EventRoomEmpty, eventData.RoomID, 0)
			Publish(EventRoomEmpty, roomDestroyData)
		}
	}
}

func (p *PlayerEventListener) handlePlayerAction(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		action, _ := eventData.GetString("action")
		playerName, _ := eventData.GetString("player_name")
		log.Printf("⚡ Player Action - %s performed %s", playerName, action)

		// 处理玩家行动
		// 验证行动合法性
		// 更新游戏状态
		// 广播行动结果
	}
}

func (p *PlayerEventListener) handlePlayerDeath(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		playerName, _ := eventData.GetString("player_name")
		log.Printf("💀 Player Death - %s died", playerName)

		// 处理玩家死亡
		// 移除玩家
		// 检查游戏结束条件
		// 触发复活机制（如果有）
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
				EventCardDraw,
				EventCardPlay,
				EventCardDiscard,
				EventCardShuffle,
				EventDeckEmpty,
			},
			Priority: 30,
		},
	}
}

func (c *CardEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case EventCardDraw:
		c.handleCardDraw(data)
	case EventCardPlay:
		c.handleCardPlay(data)
	case EventCardDiscard:
		c.handleCardDiscard(data)
	case EventCardShuffle:
		c.handleCardShuffle(data)
	case EventDeckEmpty:
		c.handleDeckEmpty(data)
	default:
		log.Printf("CardEventListener: Unknown event type: %s", eventType)
	}
}

func (c *CardEventListener) handleCardDraw(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
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
	if eventData, ok := data.(*EventData); ok {
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
			damageData := NewEventData(EventDamage, "card_system", map[string]interface{}{
				"target":   target,
				"damage":   damage,
				"source":   cardName,
				"attacker": playerName,
			})
			Publish(EventDamage, damageData)
		}
	}
}

func (c *CardEventListener) handleCardDiscard(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		cardName, _ := eventData.GetString("card_name")
		playerName, _ := eventData.GetString("player_name")
		log.Printf("🗑️ Card Discard - %s discarded %s", playerName, cardName)

		// 处理弃牌逻辑
		// 移动卡牌到弃牌堆
		// 触发弃牌效果
	}
}

func (c *CardEventListener) handleCardShuffle(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("🔀 Card Shuffle - Deck shuffled in room %s", eventData.RoomID)

		// 处理洗牌逻辑
		// 重新排列牌库
		// 通知所有玩家
	}
}

func (c *CardEventListener) handleDeckEmpty(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("📭 Deck Empty - No more cards in room %s", eventData.RoomID)

		// 处理牌库为空
		// 将弃牌堆洗入牌库
		// 或者触发特殊规则

		// 自动洗牌
		shuffleData := CreateRoomEventData(EventCardShuffle, eventData.RoomID, 0)
		Publish(EventCardShuffle, shuffleData)
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
				EventBattleStart,
				EventBattleEnd,
				EventAttack,
				EventDamage,
				EventHeal,
			},
			Priority: 25,
		},
	}
}

func (b *BattleEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case EventBattleStart:
		b.handleBattleStart(data)
	case EventBattleEnd:
		b.handleBattleEnd(data)
	case EventAttack:
		b.handleAttack(data)
	case EventDamage:
		b.handleDamage(data)
	case EventHeal:
		b.handleHeal(data)
	default:
		log.Printf("BattleEventListener: Unknown event type: %s", eventType)
	}
}

func (b *BattleEventListener) handleBattleStart(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("⚔️ Battle Started in room %s", eventData.RoomID)

		// 初始化战斗状态
		// 设置回合顺序
		// 发送战斗开始消息
	}
}

func (b *BattleEventListener) handleBattleEnd(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		winner, _ := eventData.GetString("winner")
		log.Printf("🏆 Battle Ended - Winner: %s", winner)

		// 结算战斗结果
		// 发放奖励
		// 重置战斗状态
	}
}

func (b *BattleEventListener) handleAttack(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		attacker, _ := eventData.GetString("attacker")
		target, _ := eventData.GetString("target")
		log.Printf("⚔️ Attack - %s attacks %s", attacker, target)

		// 处理攻击逻辑
		// 计算伤害
		// 检查防御
	}
}

func (b *BattleEventListener) handleDamage(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		target, _ := eventData.GetString("target")
		damage, _ := eventData.GetFloat64("damage")
		source, _ := eventData.GetString("source")

		log.Printf("💥 Damage - %s takes %.1f damage from %s", target, damage, source)

		// 应用伤害
		// 检查死亡条件
		// 触发伤害效果

		// 检查是否死亡
		if currentHP, exists := eventData.GetFloat64("current_hp"); exists && currentHP <= 0 {
			deathData := NewEventData(EventPlayerDeath, "battle_system", map[string]interface{}{
				"player_name": target,
				"killer":      eventData.Data["attacker"],
			})
			deathData.SetRoom(eventData.RoomID).SetUser(target)
			Publish(EventPlayerDeath, deathData)
		}
	}
}

func (b *BattleEventListener) handleHeal(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
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
				EventSystemStart,
				EventSystemShutdown,
				EventSystemError,
				EventServerMaintenance,
			},
			Priority: 5, // 最高优先级
		},
	}
}

func (s *SystemEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case EventSystemStart:
		s.handleSystemStart(data)
	case EventSystemShutdown:
		s.handleSystemShutdown(data)
	case EventSystemError:
		s.handleSystemError(data)
	case EventServerMaintenance:
		s.handleServerMaintenance(data)
	default:
		log.Printf("SystemEventListener: Unknown event type: %s", eventType)
	}
}

func (s *SystemEventListener) handleSystemStart(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("🚀 System Start - %s", eventData.Data["message"])

		// 系统启动逻辑
		// 初始化全局状态
		// 加载配置
		// 启动服务
	}
}

func (s *SystemEventListener) handleSystemShutdown(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
		log.Printf("🔴 System Shutdown - %s", eventData.Data["message"])

		// 系统关闭逻辑
		// 保存数据
		// 断开连接
		// 清理资源
	}
}

func (s *SystemEventListener) handleSystemError(data interface{}) {
	if eventData, ok := data.(*EventData); ok {
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
	if eventData, ok := data.(*EventData); ok {
		maintenanceType, _ := eventData.GetString("type")
		log.Printf("🔧 Server Maintenance - %s", maintenanceType)

		// 维护逻辑
		// 通知玩家
		// 暂停新连接
		// 执行维护任务
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
		subscriptionID := Subscribe(eventType, func(data interface{}) {
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
			Unsubscribe(subscriptionID)
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
	lm.RegisterListener(NewPlayerEventListener())
	lm.RegisterListener(NewBattleEventListener())
	lm.RegisterListener(NewCardEventListener())

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
	systemStartData := CreateSystemEventData(EventSystemStart, "Event system initialized successfully")
	Publish(EventSystemStart, systemStartData)

	log.Printf("Event system initialized with %d listeners", listenerManager.GetListenerCount())
}

// ShutdownEventSystem 关闭事件系统
func ShutdownEventSystem() {
	log.Println("Shutting down event system...")

	// 发布系统关闭事件
	systemShutdownData := CreateSystemEventData(EventSystemShutdown, "Event system shutting down")
	PublishSync(EventSystemShutdown, systemShutdownData) // 同步发布，确保处理完成

	// 清空所有订阅
	Clear()

	log.Println("Event system shutdown complete")
}
