package match

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"GoServer/tcpgameserver/data/config"
)

// ============================================================
// 机器人对局
// ------------------------------------------------------------
// 1. 玩家开始匹配后调用 ScheduleBotMatch；若 bot.yaml 的
//    match_wait_seconds 内没有匹配到真人对手，就为其安排机器人。
// 2. 机器人只存在于房间中（没有网络连接）。每个阶段开始时，
//    chess_match.go 调用 chessbot 包为它做决策；机器人只通过规则引擎的
//    公开操作行动，与真人走完全相同的校验与结算。
// ============================================================

var botSeq int64

// NewBotName 生成本次运行内唯一的机器人名称（前缀见 bot.yaml，真人不能注册该前缀）
func NewBotName() string {
	n := atomic.AddInt64(&botSeq, 1)
	return fmt.Sprintf("%s%04d", config.Game().Bot.NamePrefix, 1000+n%9000)
}

// IsBot 判断玩家是否为机器人
func IsBot(username string) bool {
	return strings.HasPrefix(username, config.Game().Bot.NamePrefix)
}

// botMatchTimers 等待机器人匹配的玩家，clientID -> 计时器
var (
	botMatchMu     sync.Mutex
	botMatchTimers = make(map[string]*time.Timer)
)

// ScheduleBotMatch 玩家开始匹配时调用：等待一段时间后若仍未匹配到真人，则安排机器人对局
func ScheduleBotMatch(clientID string) {
	bot := config.Game().Bot
	if !bot.Enabled {
		return
	}

	botMatchMu.Lock()
	defer botMatchMu.Unlock()

	if old, ok := botMatchTimers[clientID]; ok {
		old.Stop()
	}
	botMatchTimers[clientID] = time.AfterFunc(bot.MatchWait(), func() {
		botMatchMu.Lock()
		delete(botMatchTimers, clientID)
		botMatchMu.Unlock()

		// StartBotMatch 内部会再次确认玩家仍在等待匹配
		if err := NewGameStartProcessor().StartBotMatch(clientID); err != nil {
			log.Printf("start bot match for %s: %v", clientID, err)
		}
	})
}

// CancelBotMatch 取消该玩家的机器人补位计时（玩家取消匹配时调用）。
func CancelBotMatch(clientID string) {
	botMatchMu.Lock()
	defer botMatchMu.Unlock()
	if t, ok := botMatchTimers[clientID]; ok {
		t.Stop()
		delete(botMatchTimers, clientID)
	}
}
