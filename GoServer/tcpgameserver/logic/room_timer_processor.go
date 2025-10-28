package logic

import (
	"log"
	"sync"
	"time"

	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// 计时器配置常量
const (
	DefaultTimerDuration = 30 * time.Second // 默认计时时间：30秒
)

// RoomTimerProcessor 房间计时处理器
type RoomTimerProcessor struct {
	Name       string
	timers     map[string]*time.Timer // 房间ID -> 计时器
	timerMutex sync.RWMutex           // 计时器操作的互斥锁
	Duration   time.Duration          // 计时时长，可配置
}

// 全局处理器实例
var GlobalRoomTimerProcessor *RoomTimerProcessor

// init 初始化全局处理器实例
func init() {
	GlobalRoomTimerProcessor = &RoomTimerProcessor{
		Name:     "RoomTimerProcessor",
		timers:   make(map[string]*time.Timer),
		Duration: DefaultTimerDuration,
	}
	log.Printf("RoomTimerProcessor: Global instance initialized with %v duration", DefaultTimerDuration)
}

// NewRoomTimerProcessor 创建新的房间计时处理器
func NewRoomTimerProcessor() *RoomTimerProcessor {
	return &RoomTimerProcessor{
		Name:     "RoomTimerProcessor",
		timers:   make(map[string]*time.Timer),
		Duration: DefaultTimerDuration,
	}
}

// StartRoomTimer 开始房间计时
func (rtp *RoomTimerProcessor) StartRoomTimer(roomID string) error {

	// 获取房间信息
	roomManager := service.GetRoomManager()
	room, err := roomManager.GetRoom(roomID)
	if err != nil {
		log.Printf("RoomTimerProcessor: Failed to get room %s: %v", roomID, err)
		return err
	}

	// 启动对Round为current的玩家的计时
	err = rtp.startPlayerTimer(room)
	if err != nil {
		log.Printf("RoomTimerProcessor: Failed to start player timer for room %s: %v", roomID, err)
		return err
	}

	log.Printf("RoomTimerProcessor: Successfully started timer for room %s", roomID)
	return nil
}

// startPlayerTimer 对房间内Round为current的玩家进行计时
func (rtp *RoomTimerProcessor) startPlayerTimer(room *types.RoomInfo) error {
	log.Printf("RoomTimerProcessor: Starting %v timer for room %s", rtp.Duration, room.RoomID)

	// 停止现有计时器（如果存在）
	rtp.stopRoomTimer(room.RoomID)

	// 检查是否有Round为current的玩家
	hasCurrentPlayer := false
	for _, player := range room.Players {
		if player.Round == "current" {
			hasCurrentPlayer = true
			log.Printf("RoomTimerProcessor: Found current player %s in room %s", player.Username, room.RoomID)
			break
		}
	}

	if !hasCurrentPlayer {
		log.Printf("RoomTimerProcessor: No current player found in room %s", room.RoomID)
		return nil
	}

	// 创建计时器，使用可配置的时长
	timer := time.AfterFunc(rtp.Duration, func() {
		log.Printf("RoomTimerProcessor: Timer expired for room %s after %v, forcing card play", room.RoomID, rtp.Duration)

		// 调用强制出牌请求
		err := rtp.forceCardPlay(room.RoomID)
		if err != nil {
			log.Printf("RoomTimerProcessor: Failed to force card play for room %s: %v", room.RoomID, err)
		}

		// 清理计时器记录
		rtp.timerMutex.Lock()
		delete(rtp.timers, room.RoomID)
		rtp.timerMutex.Unlock()
	})

	// 保存计时器引用
	rtp.timerMutex.Lock()
	rtp.timers[room.RoomID] = timer
	rtp.timerMutex.Unlock()

	log.Printf("RoomTimerProcessor: Successfully started %v timer for room %s", rtp.Duration, room.RoomID)
	return nil
}

// forceCardPlay 发送强制出牌请求
func (rtp *RoomTimerProcessor) forceCardPlay(roomID string) error {
	// 发布游戏状态更新事件
	stateUpdateData := events.NewEventData(events.EventGameStateUpdate, "force_cardplay_processor", map[string]interface{}{})
	stateUpdateData.SetRoom(roomID)
	events.Publish(events.EventGameStateUpdate, stateUpdateData)
	return nil
}

// stopRoomTimer 结束房间计时
func (rtp *RoomTimerProcessor) stopRoomTimer(roomID string) error {
	log.Printf("RoomTimerProcessor: Stopping timer for room %s", roomID)

	rtp.timerMutex.Lock()
	defer rtp.timerMutex.Unlock()

	// 查找并停止计时器
	if timer, exists := rtp.timers[roomID]; exists {
		timer.Stop()
		delete(rtp.timers, roomID)
		log.Printf("RoomTimerProcessor: Successfully stopped timer for room %s", roomID)
	} else {
		log.Printf("RoomTimerProcessor: No timer found for room %s", roomID)
	}

	return nil
}

// StopRoomTimer 外部调用接口，用于结束房间计时
func (rtp *RoomTimerProcessor) StopRoomTimer(data interface{}) error {
	eventData, ok := data.(*events.EventData)
	if !ok {
		log.Printf("RoomTimerProcessor: Invalid event data type for stop timer")
		return nil
	}

	return rtp.stopRoomTimer(eventData.RoomID)
}

// CleanupAllTimers 清理所有计时器（用于系统关闭时）
func (rtp *RoomTimerProcessor) CleanupAllTimers() {
	log.Printf("RoomTimerProcessor: Cleaning up all timers")

	rtp.timerMutex.Lock()
	defer rtp.timerMutex.Unlock()

	for roomID, timer := range rtp.timers {
		timer.Stop()
		log.Printf("RoomTimerProcessor: Stopped timer for room %s", roomID)
	}

	// 清空计时器映射
	rtp.timers = make(map[string]*time.Timer)
	log.Printf("RoomTimerProcessor: All timers cleaned up")
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
