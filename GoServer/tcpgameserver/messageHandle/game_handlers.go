package tcpserver

import (
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/logic"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/types"
)

// HandlePing 心跳。
func HandlePing(ctx *Context, msg *pb.ClientMessage) {
	ctx.Reply(&pb.ServerMessage{Body: &pb.ServerMessage_Pong{Pong: &pb.Pong{
		ClientTimeMs: msg.GetPing().GetClientTimeMs(),
		ServerTimeMs: time.Now().UnixMilli(),
	}}})
}

// HandleReady 玩家开始匹配（含对局结束后再来一局）：尝试匹配真人对手；
// 等待超时仍未匹配成功时由机器人补位（见 bot.yaml）。
func HandleReady(ctx *Context, _ *pb.ClientMessage) {
	ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusReady)
	ctx.Reply(matchmakingState(true))
	startMatchmaking(ctx.ClientID, ctx.ConnManager.GetConnectionStats()["ready"])
}

// HandleCancelMatch 取消匹配：离开匹配队列并停止机器人补位计时。
// 已经开局的玩家不受影响（状态不是 Ready 时什么都不做）。
func HandleCancelMatch(ctx *Context, _ *pb.ClientMessage) {
	if info, ok := ctx.ConnManager.GetConnectionByClientID(ctx.ClientID); ok && info.GetStatus() == types.StatusReady {
		ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusLoggedIn)
	}
	logic.CancelBotMatch(ctx.ClientID)
	ctx.Reply(matchmakingState(false))
}

func matchmakingState(searching bool) *pb.ServerMessage {
	return &pb.ServerMessage{Body: &pb.ServerMessage_Matchmaking{Matchmaking: &pb.MatchmakingState{Searching: searching}}}
}

// startMatchmaking 人数足够时立即匹配真人；同时为该玩家开始机器人补位倒计时
func startMatchmaking(clientID string, readyCount int) {
	if readyCount >= config.Game().Rules.Match.PlayersPerRoom {
		// 异步执行（与原事件总线行为一致）：开局会给房间内所有玩家推送消息，不阻塞当前连接的读循环
		go logic.TryStartMatch()
	}
	logic.ScheduleBotMatch(clientID)
}

// HandleGameAction 对局内的所有操作（买卖、刷新、摆阵、就绪、选择……）。
// 只转发意图；校验、判定与结算全部在 logic / rules 中完成。
func HandleGameAction(ctx *Context, msg *pb.ClientMessage) {
	action := msg.GetGameAction()
	if action == nil || action.GetAction() == nil {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_REQUEST, config.Msg(config.MsgEmptyMessageBody))
		return
	}
	logic.HandleGameAction(logic.GameActionRequest{
		Player:   ctx.Username(),
		ClientID: ctx.ClientID,
		Seq:      ctx.Seq,
		Action:   action,
	})
}
