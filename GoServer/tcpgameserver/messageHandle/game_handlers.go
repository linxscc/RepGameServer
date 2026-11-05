package tcpserver

import (
	"time"

	"GoServer/tcpgameserver/events"
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

// HandleReady 玩家准备：先下发羁绊表，人数足够时开始匹配。
func HandleReady(ctx *Context, _ *pb.ClientMessage) {
	ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusReady)
	stats := ctx.ConnManager.GetConnectionStats()

	bondsData := events.CreateRoomEventData(events.EventCardBonds, "", int(stats["ready"]))
	bondsData.AddData("client_id", ctx.ClientID)
	events.Publish(events.EventCardBonds, bondsData)

	tryStartMatchmaking(stats["ready"], "user_ready_handler")
}

// HandleRestart 对局结束后再来一局。
func HandleRestart(ctx *Context, _ *pb.ClientMessage) {
	ctx.ConnManager.SetPlayerStatus(ctx.ClientID, types.StatusReady)
	tryStartMatchmaking(ctx.ConnManager.GetConnectionStats()["ready"], "user_restart_handler")
}

func tryStartMatchmaking(readyCount int, source string) {
	if readyCount < 2 {
		return
	}
	gameStartData := events.CreateRoomEventData(events.EventGameStart, "new_room", readyCount)
	gameStartData.AddData("message", "Players ready, starting matchmaking")
	gameStartData.AddData("trigger_source", source)
	events.Publish(events.EventGameStart, gameStartData)
}

// HandlePlayCards 出牌：只接收牌的 UID，校验与结算交给业务层。
func HandlePlayCards(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetPlayCards()
	if req.GetRoomId() == "" || len(req.GetCardUids()) == 0 {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CARDS, "no cards to play")
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

// HandleComposeCards 合成：数量须为 3 的倍数，其余校验交给业务层。
func HandleComposeCards(ctx *Context, msg *pb.ClientMessage) {
	req := msg.GetComposeCards()
	if req.GetRoomId() == "" || len(req.GetCardUids()) == 0 || len(req.GetCardUids())%3 != 0 {
		ctx.Fail(pb.ErrorCode_ERROR_CODE_INVALID_CARDS, "card count must be a multiple of 3")
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
