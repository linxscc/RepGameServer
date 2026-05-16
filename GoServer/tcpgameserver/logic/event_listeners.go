package logic

import (
	"encoding/json"
	"net"
	"time"

	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
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
				events.EventGameStateUpdate,
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
	case events.EventGameStateUpdate:
		g.handleGameStateUpdate(data)
	default:
	}
}

func (g *GameEventListener) handleGameStart(data interface{}) {

	// 直接创建并使用GameStartProcessor处理游戏开始逻辑
	processor := &GameStartProcessor{}
	err := processor.ProcessGameStart(data)
	if err != nil {
	} else {
	}
}

func (g *GameEventListener) handleGameEnd(data interface{}) {

	// 直接创建并使用GameEndProcessor处理游戏结束逻辑
	processor := NewGameEndProcessor()
	processor.ProcessGameEnd(data)
}

func (g *GameEventListener) handleGamePause(data interface{}) {
	if _, ok := data.(*events.EventData); ok {
		// 暂停游戏逻辑
	}
}

func (g *GameEventListener) handleGameResume(data interface{}) {
	if _, ok := data.(*events.EventData); ok {
		// 恢复游戏逻辑
	}
}

func (g *GameEventListener) handleGameStateUpdate(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {

		// 向房间内所有玩家发送游戏状态更新
		broadcaster := NewGameStateBroadcaster()
		broadcaster.BroadcastGameStateToRoom(eventData)
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
				events.EventCardCompose,
				events.EventDeckEmpty,
				events.EventCardBonds,
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
	case events.EventCardCompose:
		c.handleCardCompose(data)
	case events.EventDeckEmpty:
		c.handleDeckEmpty(data)
	case events.EventCardBonds:
		c.handleCardBonds(data)
	default:
	}
}

func (c *CardEventListener) handleCardDraw(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("card_name")
		_, _ = eventData.GetString("player_name")

		// 处理抽卡逻辑
		// 更新玩家手牌
		// 检查手牌上限
		// 触发抽卡特效
	}
}

func (c *CardEventListener) handleCardPlay(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		// 使用PlayCardProcessor处理出牌逻辑（包含所有验证）
		processor := NewPlayCardProcessor()
		processor.ProcessPlayCard(eventData)
	}
}

func (c *CardEventListener) handleCardCompose(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		// 使用CardComposeProcessor处理合成逻辑
		processor := NewCardComposeProcessor()
		processor.ProcessCardCompose(eventData)

	}
}

func (c *CardEventListener) handleDeckEmpty(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 处理牌库为空
		// 将弃牌堆洗入牌库
		// 或者触发特殊规则

	}
}

func (c *CardEventListener) handleCardBonds(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {

		connManager := service.GetConnectionManager()
		ClientID, _ := eventData.GetString("client_id")
		// 将羁绊数据转换为切片格式，便于序列化
		bondPoolManager := cards.GetBondPoolManager()
		allBonds := bondPoolManager.GetAllBonds()
		bondList := make([]*models.BondModel, 0, len(allBonds))
		for _, bond := range allBonds {
			bondList = append(bondList, bond)
		}

		// 发送羁绊数据消息 (5002)
		response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(5002, bondList)

		// 获取玩家连接并发送消息
		clientInfo, exists := connManager.GetConnectionByClientID(ClientID)
		if !exists || clientInfo == nil || clientInfo.Conn == nil {
			return
		}

		sendTCPResponse(clientInfo.Conn, response)

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
	}
}

func (b *BattleEventListener) handleBattleStart(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 初始化战斗状态
		// 设置回合顺序
		// 发送战斗开始消息
	}
}

func (b *BattleEventListener) handleBattleEnd(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("winner")

		// 结算战斗结果
		// 发放奖励
		// 重置战斗状态
	}
}

func (b *BattleEventListener) handleAttack(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("attacker")
		_, _ = eventData.GetString("target")

		// 处理攻击逻辑
		// 计算伤害
		// 检查防御
	}
}

func (b *BattleEventListener) handleDamage(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		target, _ := eventData.GetString("target")
		_, _ = eventData.GetFloat64("damage")
		_, _ = eventData.GetString("source")


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
		_, _ = eventData.GetString("target")
		_, _ = eventData.GetFloat64("heal_amount")
		_, _ = eventData.GetString("source")


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
	}
}

func (s *SystemEventListener) handleSystemStart(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 系统启动逻辑
		// 初始化全局状态
		// 加载配置
		// 启动服务

		// 加载响应码配置文件
		if err := tools.LoadResponseCodes(); err != nil {
		} else {
		}

		// 初始化卡牌池
		if err := cards.InitCardPool(); err != nil {
		} else {
		}

		// 初始化羁绊池
		if err := cards.InitBondPool(); err != nil {
		} else {
		}

	}
}

func (s *SystemEventListener) handleSystemShutdown(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 系统关闭逻辑
		// 保存数据
		// 断开连接
		// 清理资源
	}
}

func (s *SystemEventListener) handleSystemError(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("error")
		_, _ = eventData.GetString("severity")


		// 错误处理逻辑
		// 记录错误日志
		// 发送告警
		// 尝试恢复
	}
}

func (s *SystemEventListener) handleServerMaintenance(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("type")

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
	}
}

func (r *RoomEventListener) handleRoomCreate(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

	}
}

func (r *RoomEventListener) handleRoomDestroy(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 处理房间销毁逻辑
		// 清理房间资源
		// 通知相关玩家
		// 保存房间数据
	}
}

func (r *RoomEventListener) handleRoomFull(data interface{}) {
	if _, ok := data.(*events.EventData); ok {

		// 处理房间已满逻辑
		// 拒绝新玩家加入
		// 可能触发游戏开始
	}
}

func (r *RoomEventListener) handleRoomEmpty(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {

		// 处理房间为空逻辑
		// 准备销毁房间
		destroyData := events.CreateRoomEventData(events.EventRoomDestroy, eventData.RoomID, 0)
		events.Publish(events.EventRoomDestroy, destroyData)
	}
}

// ConnectionEventListener 连接事件监听器
type ConnectionEventListener struct {
	BaseEventListener
}

func NewConnectionEventListener() *ConnectionEventListener {
	return &ConnectionEventListener{
		BaseEventListener: BaseEventListener{
			Name: "ConnectionEventListener",
			EventTypes: []string{
				events.EventClientConnect,
				events.EventClientDisconnect,
				events.EventClientTimeout,
				events.EventClientBind,
				events.EventClientUnbind,
				events.EventClientKicked,
				events.EventClientReconnect,
				events.EventConnectionCleanup,
			},
			Priority: 20, // 中等优先级
		},
	}
}

func (c *ConnectionEventListener) HandleEvent(eventType string, data interface{}) {
	switch eventType {
	case events.EventClientConnect:
		c.handleClientConnect(data)
	case events.EventClientDisconnect:
		c.handleClientDisconnect(data)
	case events.EventClientTimeout:
		c.handleClientTimeout(data)
	case events.EventClientBind:
		c.handleClientBind(data)
	case events.EventClientUnbind:
		c.handleClientUnbind(data)
	case events.EventClientKicked:
		c.handleClientKicked(data)
	case events.EventClientReconnect:
		c.handleClientReconnect(data)
	case events.EventConnectionCleanup:
		c.handleConnectionCleanup(data)
	default:
	}
}

func (c *ConnectionEventListener) handleClientConnect(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		clientID, _ := eventData.GetString("client_id")
		connectionType, _ := eventData.GetString("connection_type")
		userAgent, _ := eventData.GetString("user_agent")
		version, _ := eventData.GetString("version")

		// 获取连接管理器来设置连接元数据
		connManager := service.GetConnectionManager()
		if clientInfo, exists := connManager.GetConnectionByClientID(clientID); exists {
			// 设置连接元数据
			clientInfo.SetMetadata("connection_type", connectionType)
			clientInfo.SetMetadata("user_agent", userAgent)
			clientInfo.SetMetadata("version", version)
			if firstConnectTime, exists := eventData.GetData("first_connect_time"); exists {
				clientInfo.SetMetadata("first_connect_time", firstConnectTime)
			}

			// 发送欢迎消息
			welcomeResponse := tools.GlobalResponseHelper.CreateSuccessTcpResponse(1001, map[string]interface{}{
				"client_id":   clientID,
				"server_time": time.Now().Unix(),
				"status":      "connected",
				"message":     "Welcome to the game server!",
			})

			// 通过连接发送欢迎消息
			if welcomeData, err := json.Marshal(welcomeResponse); err == nil {
				welcomeData = append(welcomeData, '\n')
				if _, writeErr := clientInfo.Conn.Write(welcomeData); writeErr != nil {
				}
			}
		}

		// 处理客户端连接逻辑
		// 初始化连接状态
		// 记录连接统计
		// 发送欢迎消息已完成
	}
}

func (c *ConnectionEventListener) handleClientDisconnect(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		clientID, _ := eventData.GetString("client_id")
		username, _ := eventData.GetString("username")
		reason, _ := eventData.GetString("reason")

		handler := NewDisconnectHandler()
		err := handler.HandlePlayerDisconnect(clientID, username, reason)
		if err != nil {
		} else {
		}
	}
}

func (c *ConnectionEventListener) handleClientTimeout(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		clientID, _ := eventData.GetString("client_id")
		username, _ := eventData.GetString("username")
		lastActivity, _ := eventData.GetString("last_activity")


		// 处理客户端超时逻辑
		// 标记为超时状态
		// 给予短暂重连时间
		// 或直接断开连接

		// 触发断开连接事件
		disconnectData := events.CreateUserConnectionEventData(
			events.EventClientDisconnect, clientID, username, "")
		disconnectData.AddData("reason", "timeout")
		disconnectData.AddData("last_activity", lastActivity)
		events.Publish(events.EventClientDisconnect, disconnectData)
	}
}

func (c *ConnectionEventListener) handleClientBind(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("client_id")
		_, _ = eventData.GetString("username")
		_, _ = eventData.GetString("remote_addr")


		// 处理用户绑定逻辑
		// 加载用户数据
		// 设置在线状态
		// 发送登录成功消息
		// 同步游戏状态
	}
}

func (c *ConnectionEventListener) handleClientUnbind(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetString("client_id")
		_, _ = eventData.GetString("username")
		_, _ = eventData.GetString("reason")


		// 处理用户解绑逻辑
		// 保存用户数据
		// 设置离线状态
		// 清理用户相关状态
	}
}

func (c *ConnectionEventListener) handleClientKicked(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		clientID, _ := eventData.GetString("client_id")
		username, _ := eventData.GetString("username")
		kickReason, _ := eventData.GetString("kick_reason")
		kickedBy, _ := eventData.GetString("kicked_by")
		newClientID, _ := eventData.GetString("new_client_id")

		// 获取连接管理器
		connManager := service.GetConnectionManager()

		// 先触发断开连接事件处理原客户端
		disconnectData := events.CreateUserConnectionEventData(
			events.EventClientDisconnect, clientID, username, "")
		disconnectData.AddData("reason", "kicked")
		disconnectData.AddData("kick_reason", kickReason)
		disconnectData.AddData("kicked_by", kickedBy)
		events.Publish(events.EventClientDisconnect, disconnectData) // 处理新客户端绑定
		if newClientID != "" {
			// 获取新客户端连接
			newClient, exists := connManager.GetConnectionByClientID(newClientID)
			if exists && newClient != nil {
				// 绑定用户到新连接
				err := connManager.BindUser(newClientID, username)
				if err != nil {
					return
				}

				// 设置新客户端状态为已登录
				connManager.SetPlayerStatus(newClientID, types.StatusLoggedIn)

				// 发送登录成功响应给新客户端
				if newClient.Conn != nil {
					response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(2001, map[string]interface{}{
						"username": username,
					})
					sendTCPResponse(newClient.Conn, response)
				}
			} else {
			}
		}
	}
}

func (c *ConnectionEventListener) handleClientReconnect(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		clientID, _ := eventData.GetString("client_id")
		username, _ := eventData.GetString("username")

		handler := NewReconnectionHandler()
		err := handler.HandlePlayerReconnection(clientID, username)
		if err != nil {
		} else {
		}
	}
}

func (c *ConnectionEventListener) handleConnectionCleanup(data interface{}) {
	if eventData, ok := data.(*events.EventData); ok {
		_, _ = eventData.GetInt("cleaned_count")
		_, _ = eventData.GetInt("total_connections")
		_, _ = eventData.GetString("cleanup_duration")


		// 处理连接清理逻辑
		// 记录清理统计
		// 优化内存使用
		// 更新连接监控数据
	}
}

// sendTCPResponse 发送TCP响应消息
func sendTCPResponse(conn net.Conn, resp *models.TcpResponse) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	_, err = conn.Write(jsonBytes)
	if err != nil {
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
	lm.RegisterListener(NewConnectionEventListener())

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

	// 获取监听器管理器并注册默认监听器
	listenerManager := GetListenerManager()
	listenerManager.RegisterAllDefaultListeners()

	// 发布系统启动事件
	systemStartData := events.CreateSystemEventData(events.EventSystemStart, "Event system initialized successfully")
	events.Publish(events.EventSystemStart, systemStartData)
}

// ShutdownEventSystem 关闭事件系统
func ShutdownEventSystem() {

	// 发布系统关闭事件
	systemShutdownData := events.CreateSystemEventData(events.EventSystemShutdown, "Event system shutting down")
	events.PublishSync(events.EventSystemShutdown, systemShutdownData) // 同步发布，确保处理完成

	// 清空所有订阅
	events.Clear()
}
