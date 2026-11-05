package tcpserver

import (
	"time"

	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/logic"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
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

// HandlePlayCards 出牌：只接收牌的 UID，校验与结算交给业务层。
func HandlePlayCards(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetPlayCards()
	if req.GetRoomId() == "" || len(req.GetCardUids()) == 0 {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CARDS, config.Msg(config.MsgNoCardsToPlay))
		return
	}

	events.Publish(events.EventCardPlay, events.NewEventData(events.EventCardPlay, "user_play_card_handler", map[string]interface{}{
		"player":     ctx.Username(),
		"self_cards": protocol.CardsFromUIDs(req.GetCardUids()),
		"room_id":    req.GetRoomId(),
		"client_id":  ctx.ClientID,
		"seq":        ctx.Seq,
	}))
}

// HandleComposeCards 合成：数量须为 group_size（game_rules.yaml）的倍数，其余校验交给业务层。
func HandleComposeCards(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetComposeCards()
	need := config.Game().Rules.Compose.GroupSize
	if req.GetRoomId() == "" || len(req.GetCardUids()) == 0 || len(req.GetCardUids())%need != 0 {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CARDS, config.Msg(config.MsgComposeCountInvalid, need))
		return
	}

	events.Publish(events.EventCardCompose, events.NewEventData(events.EventCardCompose, "user_compose_card_handler", map[string]interface{}{
		"room_id":   req.GetRoomId(),
		"player":    ctx.Username(),
		"cards":     protocol.CardsFromUIDs(req.GetCardUids()),
		"client_id": ctx.ClientID,
		"seq":       ctx.Seq,
	}))
}
