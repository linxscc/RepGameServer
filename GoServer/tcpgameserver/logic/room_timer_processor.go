package logic

import (
	"sync"
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// 回合调度器：每个回合开始时调用 StartRoomTimer。
//   - 当前回合是机器人：bot.yaml 的 think_delay_ms 后由机器人出牌；
//   - 当前回合是真人：game_rules.yaml 的 turn.timeout_seconds 后由服务器替他打出第一张手牌。

// RoomTimerProcessor 房间计时处理器
type RoomTimerProcessor struct {
	Name       string
	timers     map[string]*time.Timer // 房间ID -> 计时器
	timerMutex sync.RWMutex           // 计时器操作的互斥锁
}

// 全局处理器实例
var GlobalRoomTimerProcessor *RoomTimerProcessor

// init 初始化全局处理器实例
func init() {
	GlobalRoomTimerProcessor = &RoomTimerProcessor{
		Name:   "RoomTimerProcessor",
		timers: make(map[string]*time.Timer),
	}
}

// NewRoomTimerProcessor 创建新的房间计时处理器
func NewRoomTimerProcessor() *RoomTimerProcessor {
	return &RoomTimerProcessor{
		Name:   "RoomTimerProcessor",
		timers: make(map[string]*time.Timer),
	}
}

// StartRoomTimer 开始房间计时
func (rtp *RoomTimerProcessor) StartRoomTimer(roomID string) error {

	// 获取房间信息
	roomManager := service.GetRoomManager()
	room, err := roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 启动对Round为current的玩家的计时
	err = rtp.startPlayerTimer(room)
	if err != nil {
		return err
	}

	return nil
}

// startPlayerTimer 为房间内当前回合的玩家安排计时：机器人到点出牌，真人超时被自动出牌
func (rtp *RoomTimerProcessor) startPlayerTimer(room *types.RoomInfo) error {
	// 停止上一回合的计时器（如果存在）
	rtp.stopRoomTimer(room.RoomID)

	currentPlayer := room.CurrentPlayer()
	if currentPlayer == "" {
		return nil
	}

	roomID := room.RoomID
	delay := config.Game().Rules.TurnTimeout()
	action := func() { rtp.forceCardPlay(roomID, currentPlayer) }
	if IsBot(currentPlayer) {
		delay = config.Game().Bot.ThinkDelay()
		action = func() { PlayBotTurn(roomID, currentPlayer) }
	}

	var timer *time.Timer
	timer = time.AfterFunc(delay, func() {
		// 只清理自己这一个计时器，避免误删下一回合新建的计时器
		rtp.timerMutex.Lock()
		if rtp.timers[roomID] == timer {
			delete(rtp.timers, roomID)
		}
		rtp.timerMutex.Unlock()

		action()
	})

	rtp.timerMutex.Lock()
	rtp.timers[roomID] = timer
	rtp.timerMutex.Unlock()

	return nil
}

// forceCardPlay 真人玩家回合超时：由服务器替其打出手牌中的第一张。
// 走正常出牌流程（会再次校验是否仍是该玩家的回合），结算后广播的状态原因为 FORCED_PLAY。
func (rtp *RoomTimerProcessor) forceCardPlay(roomID, username string) {
	PlayCards(PlayRequest{RoomID: roomID, Player: username, Forced: true})
}

// stopRoomTimer 结束房间计时
func (rtp *RoomTimerProcessor) stopRoomTimer(roomID string) error {

	rtp.timerMutex.Lock()
	defer rtp.timerMutex.Unlock()

	// 查找并停止计时器
	if timer, exists := rtp.timers[roomID]; exists {
		timer.Stop()
		delete(rtp.timers, roomID)
	} else {
	}

	return nil
}

// StopRoomTimer 外部调用接口，用于结束房间计时
func (rtp *RoomTimerProcessor) StopRoomTimer(data interface{}) error {
	eventData, ok := data.(*events.EventData)
	if !ok {
		return nil
	}

	return rtp.stopRoomTimer(eventData.RoomID)
}

// CleanupAllTimers 清理所有计时器（用于系统关闭时）
func (rtp *RoomTimerProcessor) CleanupAllTimers() {

	rtp.timerMutex.Lock()
	defer rtp.timerMutex.Unlock()

	for _, timer := range rtp.timers {
		timer.Stop()
	}

	// 清空计时器映射
	rtp.timers = make(map[string]*time.Timer)
}

// 全局便捷函数，供外部直接调用

// StartTimer 开始房间计时 - 全局函数
func StartTimer(roomID string) error {
	return GlobalRoomTimerProcessor.StartRoomTimer(roomID)
}

// StopTimer 停止房间计时 - 全局函数
func StopTimer(roomID string) error {
	return GlobalRoomTimerProcessor.stopRoomTimer(roomID)
}

// CleanupTimers 清理所有计时器 - 全局函数
func CleanupTimers() {
	GlobalRoomTimerProcessor.CleanupAllTimers()
}
