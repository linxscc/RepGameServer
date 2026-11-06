package logic

import (
	"bufio"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/protocol"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// fakeClient drains a net.Pipe and records every ServerMessage.
type fakeClient struct {
	mu   sync.Mutex
	msgs []*pb.ServerMessage
	conn net.Conn
}

func newFakeClient(t *testing.T, clientID, username string) *fakeClient {
	t.Helper()
	server, client := net.Pipe()
	fc := &fakeClient{conn: client}
	go func() {
		r := bufio.NewReader(client)
		for {
			payload, err := protocol.ReadFrame(r)
			if err != nil {
				return
			}
			msg := &pb.ServerMessage{}
			if proto.Unmarshal(payload, msg) == nil {
				fc.mu.Lock()
				fc.msgs = append(fc.msgs, msg)
				fc.mu.Unlock()
			}
		}
	}()
	cm := service.GetConnectionManager()
	cm.AddConnection(server, clientID)
	if err := cm.BindUser(clientID, username); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return fc
}

// waitFor returns the first recorded message matching pred, waiting up to 2s.
func (fc *fakeClient) waitFor(t *testing.T, what string, pred func(*pb.ServerMessage) bool) *pb.ServerMessage {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		fc.mu.Lock()
		for _, m := range fc.msgs {
			if pred(m) {
				fc.mu.Unlock()
				return m
			}
		}
		fc.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
	return nil
}

func (fc *fakeClient) last(t *testing.T) *pb.MatchState {
	t.Helper()
	time.Sleep(30 * time.Millisecond)
	fc.mu.Lock()
	defer fc.mu.Unlock()
	for i := len(fc.msgs) - 1; i >= 0; i-- {
		if st := fc.msgs[i].GetMatchState(); st != nil {
			return st
		}
	}
	t.Fatal("no MatchState received")
	return nil
}

// newChessTestRoom puts a connected human and a bot into a fresh chess match.
func newChessTestRoom(t *testing.T, human, clientID string) (*types.RoomInfo, *fakeClient) {
	t.Helper()
	fc := newFakeClient(t, clientID, human)
	rm := service.GetRoomManager()
	room, err := rm.CreateRoom("chess-test", 2)
	if err != nil {
		t.Fatal(err)
	}
	bot := NewBotName()
	for _, u := range []string{human, bot} {
		if err := rm.JoinRoom(room.RoomID, u); err != nil {
			t.Fatal(err)
		}
	}
	cm := service.GetConnectionManager()
	_ = cm.SetPlayerStatus(clientID, types.StatusInGame)
	_ = cm.SetPlayerGameRoom(clientID, room.RoomID)
	if err := withRoom(room.RoomID, func(r *types.RoomInfo) error { return startChessMatch(r, []string{human, bot}) }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cr := getChessRoom(room.RoomID); cr != nil {
			cr.stop()
			chessRooms.Delete(room.RoomID)
		}
	})
	return room, fc
}

func act(username, clientID string, seq uint32, a *pb.GameAction) {
	HandleGameAction(GameActionRequest{Player: username, ClientID: clientID, Seq: seq, Action: a})
}

func pickFamily(i int32) *pb.GameAction {
	return &pb.GameAction{Action: &pb.GameAction_PickFamily{PickFamily: &pb.PickFamily{OfferIndex: i}}}
}

func ready() *pb.GameAction {
	return &pb.GameAction{Action: &pb.GameAction_SetReady{SetReady: &pb.SetReady{Ready: true}}}
}

// forceAdvance runs the phase machine as if the current phase timer fired.
func forceAdvance(t *testing.T, roomID string) {
	t.Helper()
	_ = withRoom(roomID, func(room *types.RoomInfo) error {
		if cr := getChessRoom(roomID); cr != nil {
			cr.advance(room)
		}
		return nil
	})
}

func TestChessMatchStartsWithConfigAndFamilyOffer(t *testing.T) {
	_, fc := newChessTestRoom(t, "erin", "c-erin")
	fc.waitFor(t, "MatchConfig", func(m *pb.ServerMessage) bool { return len(m.GetMatchConfig().GetLines()) == 12 })
	st := fc.waitFor(t, "family MatchState", func(m *pb.ServerMessage) bool {
		return m.GetMatchState().GetPhase() == pb.MatchPhase_MATCH_PHASE_FAMILY
	}).GetMatchState()
	if len(st.GetFamilyOffer()) != 3 || st.GetPhaseDeadlineMs() == 0 || !st.GetOpponent().GetReady() {
		t.Fatalf("bad family state: %v", st)
	}
}

// The human picks a family; since the bot already picked, the match begins at once.
func TestPickingFamilyStartsPrepAndPreviewIsServerComputed(t *testing.T) {
	room, fc := newChessTestRoom(t, "frank", "c-frank")
	act("frank", "c-frank", 7, pickFamily(0))
	reply := fc.waitFor(t, "reply to seq 7", func(m *pb.ServerMessage) bool { return m.GetReplyTo() == 7 })
	if reply.GetMatchState() == nil {
		t.Fatalf("expected MatchState reply, got %v", reply)
	}
	st := fc.last(t)
	if st.GetPhase() != pb.MatchPhase_MATCH_PHASE_PREP || st.GetRound() != 1 {
		t.Fatalf("expected round 1 prep, got %v round %d", st.GetPhase(), st.GetRound())
	}
	self := st.GetSelf()
	if len(self.GetBoard()) == 0 || len(self.GetShop()) != 5 || self.GetPreview().GetScore() <= 0 {
		t.Fatalf("prep view incomplete: board=%d shop=%d preview=%v", len(self.GetBoard()), len(self.GetShop()), self.GetPreview())
	}
	if cr := getChessRoom(room.RoomID); cr == nil || cr.match.Score(cr.match.P[0]).Score != self.GetPreview().GetScore() {
		t.Fatal("preview must equal the server's own score")
	}
}

func TestActionErrorsReplyWithCodes(t *testing.T) {
	_, fc := newChessTestRoom(t, "gina", "c-gina")
	// buying before the match begins is the wrong phase
	act("gina", "c-gina", 3, &pb.GameAction{Action: &pb.GameAction_BuyXp{BuyXp: &pb.BuyXp{}}})
	e := fc.waitFor(t, "error 3", func(m *pb.ServerMessage) bool { return m.GetReplyTo() == 3 }).GetError()
	if e.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE {
		t.Fatalf("want INVALID_GAME_STATE, got %v", e)
	}
	act("gina", "c-gina", 4, pickFamily(0))
	fc.waitFor(t, "reply 4", func(m *pb.ServerMessage) bool { return m.GetReplyTo() == 4 })
	// spend everything on xp, then one more must fail with NOT_ENOUGH_GOLD (or MAX_LEVEL)
	for seq := uint32(10); seq < 40; seq++ {
		act("gina", "c-gina", seq, &pb.GameAction{Action: &pb.GameAction_BuyXp{BuyXp: &pb.BuyXp{}}})
	}
	e = fc.waitFor(t, "gold error", func(m *pb.ServerMessage) bool { return m.GetError() != nil && m.GetReplyTo() >= 10 }).GetError()
	if e.GetCode() != pb.ErrorCode_ERROR_CODE_NOT_ENOUGH_GOLD && e.GetCode() != pb.ErrorCode_ERROR_CODE_MAX_LEVEL {
		t.Fatalf("want NOT_ENOUGH_GOLD, got %v", e)
	}
}

// Ready → settle pause (actions rejected) → next round; then play until the end.
func TestFullMatchAgainstBotReachesMatchOver(t *testing.T) {
	room, fc := newChessTestRoom(t, "hank", "c-hank")
	act("hank", "c-hank", 1, pickFamily(0))
	seq := uint32(100)
	sawSettleReject, sawRound := false, false
	for i := 0; i < 200; i++ {
		cr := getChessRoom(room.RoomID)
		if cr == nil {
			break // finished and cleaned up
		}
		switch cr.phase {
		case pb.MatchPhase_MATCH_PHASE_PREP:
			seq++
			act("hank", "c-hank", seq, ready()) // bot is ready too → settles immediately
		case pb.MatchPhase_MATCH_PHASE_SETTLE:
			if !sawSettleReject {
				seq++
				s := seq
				act("hank", "c-hank", s, &pb.GameAction{Action: &pb.GameAction_RefreshShop{RefreshShop: &pb.RefreshShop{}}})
				e := fc.waitFor(t, "settle reject", func(m *pb.ServerMessage) bool { return m.GetReplyTo() == s }).GetError()
				if e.GetCode() != pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE {
					t.Fatalf("actions during settlement must be rejected, got %v", e)
				}
				sawSettleReject = true
			}
			forceAdvance(t, room.RoomID)
		default:
			forceAdvance(t, room.RoomID)
		}
	}
	rr := fc.waitFor(t, "RoundResult", func(m *pb.ServerMessage) bool { return m.GetRoundResult() != nil }).GetRoundResult()
	if rr.GetSelf().GetScore().GetScore() <= 0 || len(rr.GetSelf().GetScore().GetSteps()) == 0 {
		t.Fatalf("round result must carry the full breakdown: %v", rr)
	}
	sawRound = true
	over := fc.waitFor(t, "MatchOver", func(m *pb.ServerMessage) bool { return m.GetMatchOver() != nil }).GetMatchOver()
	if !sawRound || !sawSettleReject || over.GetRounds() < 2 {
		t.Fatalf("unexpected end: %v", over)
	}
	if _, err := service.GetRoomManager().GetRoom(room.RoomID); err == nil {
		t.Fatal("room must be removed after the match")
	}
	info, _ := service.GetConnectionManager().GetConnectionByUsername("hank")
	if info.GetStatus() != types.StatusLoggedIn || info.GetGameRoom() != "" {
		t.Fatalf("player must be back in the lobby, got %v/%q", info.GetStatus(), info.GetGameRoom())
	}
}

// A disconnected player no longer blocks the phase; reconnecting resends config and state.
func TestDisconnectUnblocksPhaseAndResumeResendsState(t *testing.T) {
	room, _ := newChessTestRoom(t, "ivy", "c-ivy")
	act("ivy", "c-ivy", 1, pickFamily(0))
	if cr := getChessRoom(room.RoomID); cr.phase != pb.MatchPhase_MATCH_PHASE_PREP {
		t.Fatalf("expected prep, got %v", cr.phase)
	}
	PlayerDisconnected("c-ivy", "test")
	if cr := getChessRoom(room.RoomID); cr.phase != pb.MatchPhase_MATCH_PHASE_SETTLE {
		t.Fatalf("disconnect should let the round settle, got %v", cr.phase)
	}
	fc := newFakeClient(t, "c-ivy-2", "ivy-tmp")
	_ = service.GetConnectionManager().UnbindUser("c-ivy-2")
	if !ResumeGame("c-ivy-2", "ivy") {
		t.Fatal("resume failed")
	}
	fc.waitFor(t, "config on resume", func(m *pb.ServerMessage) bool { return m.GetMatchConfig() != nil })
	st := fc.waitFor(t, "state on resume", func(m *pb.ServerMessage) bool { return m.GetMatchState() != nil }).GetMatchState()
	if st.GetRound() != 1 || st.GetPhase() != pb.MatchPhase_MATCH_PHASE_SETTLE {
		t.Fatalf("resumed state wrong: round %d phase %v", st.GetRound(), st.GetPhase())
	}
	cancelReconnectGrace("ivy")
}

// The real phase timer advances the match without any player input.
func TestPhaseTimerAdvancesOnItsOwn(t *testing.T) {
	room, fc := newChessTestRoom(t, "jade", "c-jade")
	act("jade", "c-jade", 1, pickFamily(0))
	_ = withRoom(room.RoomID, func(*types.RoomInfo) error {
		getChessRoom(room.RoomID).schedule(20 * time.Millisecond)
		return nil
	})
	fc.waitFor(t, "timer-driven RoundResult", func(m *pb.ServerMessage) bool { return m.GetRoundResult() != nil })
	if cr := getChessRoom(room.RoomID); cr.phase != pb.MatchPhase_MATCH_PHASE_SETTLE {
		t.Fatalf("expected the timer to settle the round, got %v", cr.phase)
	}
}

// Surrender ends the match immediately: the surrendering player loses and the room is freed.
func TestSurrenderEndsMatch(t *testing.T) {
	room, fc := newChessTestRoom(t, "kate", "c-kate")
	act("kate", "c-kate", 1, pickFamily(0))
	act("kate", "c-kate", 2, &pb.GameAction{Action: &pb.GameAction_Surrender{Surrender: &pb.Surrender{}}})
	over := fc.waitFor(t, "MatchOver", func(m *pb.ServerMessage) bool { return m.GetMatchOver() != nil }).GetMatchOver()
	if over.GetResult() != -1 || !over.GetSurrendered() {
		t.Fatalf("surrendering player must lose: %v", over)
	}
	if _, err := service.GetRoomManager().GetRoom(room.RoomID); err == nil {
		t.Fatal("room must be removed after surrender")
	}
}
