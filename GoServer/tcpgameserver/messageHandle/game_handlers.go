package tcpserver

import (
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
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

// HandleReady 玩家开始匹配：先下发羁绊表，然后尝试匹配真人对手；
// 等待超时仍未匹配成功时由机器人补位（见 bot.yaml）。
func HandleReady(ctx *Context, _ *pb.ClientMessage) {
	ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusReady)
	stats := ctx.ConnManager.GetConnectionStats()

	bondsData := events.CreateRoomEventData(events.EventCardBonds, "", int(stats["ready"]))
	bondsData.AddData("client_id", ctx.ClientID)
	events.Publish(events.EventCardBonds, bondsData)

	startMatchmaking(ctx.ClientID, stats["ready"])
}

// HandleRestart 对局结束后再来一局（与开始匹配流程相同，只是不再重复下发羁绊表）
func HandleRestart(ctx *Context, _ *pb.ClientMessage) {
	ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusReady)
	startMatchmaking(ctx.ClientID, ctx.ConnManager.GetConnectionStats()["ready"])
}

// startMatchmaking 人数足够时立即匹配真人；同时为该玩家开始机器人补位倒计时
func startMatchmaking(clientID string, readyCount int) {
	if readyCount >= config.Game().Rules.Match.PlayersPerRoom {
		gameStartData := events.CreateRoomEventData(events.EventGameStart, "new_room", readyCount)
		gameStartData.AddData("trigger_source", "matchmaking")
		events.Publish(events.EventGameStart, gameStartData)
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
