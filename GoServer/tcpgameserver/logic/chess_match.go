package logic

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"GoServer/tcpgameserver/chessbot"
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/rules"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
)

// chess_match.go runs a rules.Match inside a room: phase timers, player actions,
// the settlement pause, bots and disconnects. Every mutation happens under
// withRoom(roomID), so a chessRoom is never touched concurrently.

type chessRoom struct {
	roomID    string
	match     *rules.Match
	phase     pb.MatchPhase // what clients see; SETTLE is a pause the engine does not have
	deadline  time.Time
	bots      [2]*chessbot.Bot // nil for humans
	timer     *time.Timer
	token     int // invalidates stale timers
	lastScore [2]float64
	settled   int // round shown during the SETTLE pause (the engine has already moved on)
	// settleDone: the seat's client finished the settle cutscene (incl. the damage animation).
	// When every connected human is done the pause ends early; settle_seconds is only a cap.
	settleDone [2]bool
}

var chessRooms sync.Map // roomID -> *chessRoom

func getChessRoom(roomID string) *chessRoom {
	if v, ok := chessRooms.Load(roomID); ok {
		return v.(*chessRoom)
	}
	return nil
}

func phaseSeconds(n int) time.Duration { return time.Duration(n) * time.Second }

// startChessMatch creates the match for a freshly filled room. Call inside withRoom.
func startChessMatch(room *types.RoomInfo, participants []string) error {
	if len(participants) != 2 {
		return errors.New("chess match needs exactly 2 players")
	}
	cfg := config.Game().Chess
	seed := time.Now().UnixNano()
	m := rules.NewMatch(cfg, rand.New(rand.NewSource(seed)), [2]string{participants[0], participants[1]})
	cr := &chessRoom{roomID: room.RoomID, match: m}
	botRng := rand.New(rand.NewSource(seed + 1))
	for i, name := range participants {
		if IsBot(name) {
			cr.bots[i] = chessbot.NewRandom(m, botRng)
			cr.bots[i].PickFamily(m, m.P[i])
		}
	}
	chessRooms.Store(room.RoomID, cr)
	room.UpdateRoomStatus(types.RoomStatusPlaying)

	cr.phase = pb.MatchPhase_MATCH_PHASE_FAMILY
	cr.schedule(phaseSeconds(cfg.Rules.Timers.FamilySeconds))
	for seat := range m.P {
		cr.sendTo(seat, matchConfigMessage(cr.roomID, m))
	}
	cr.broadcast()
	cr.advanceIfReady()
	return nil
}

// ---------- sending ----------

func (cr *chessRoom) conn(seat int) *types.ClientInfo {
	name := cr.match.P[seat].ID
	if IsBot(name) {
		return nil
	}
	info, ok := service.GetConnectionManager().GetConnectionByUsername(name)
	if !ok || info.GetConn() == nil || info.GetStatus() == types.StatusWaitingReconnect {
		return nil
	}
	return info
}

// isPresent: a bot, or a human with a live connection.
func (cr *chessRoom) isPresent(seat int) bool {
	return cr.bots[seat] != nil || cr.conn(seat) != nil
}

func (cr *chessRoom) sendTo(seat int, msg *pb.ServerMessage) {
	if info := cr.conn(seat); info != nil {
		_ = sendTo(info.GetConn(), msg)
	}
}

func (cr *chessRoom) broadcast() {
	for seat := range cr.match.P {
		cr.sendTo(seat, matchStateMessage(cr, seat, 0))
	}
}

// ---------- timers ----------

func (cr *chessRoom) schedule(d time.Duration) {
	if cr.timer != nil {
		cr.timer.Stop()
	}
	cr.token++
	token, roomID := cr.token, cr.roomID
	cr.deadline = time.Now().Add(d)
	cr.timer = time.AfterFunc(d, func() {
		_ = withRoom(roomID, func(room *types.RoomInfo) error {
			if c := getChessRoom(roomID); c != nil && c.token == token {
				c.advance(room)
			}
			return nil
		})
	})
}

func (cr *chessRoom) stop() {
	if cr.timer != nil {
		cr.timer.Stop()
	}
	cr.token++
}

// ---------- phase machine ----------

// allReady: every seat has finished the current phase, or cannot act (bot / disconnected).
func (cr *chessRoom) allReady() bool {
	for seat, p := range cr.match.P {
		if !p.Ready && cr.conn(seat) != nil {
			return false
		}
	}
	return true
}

// allSettleDone: every connected human has finished watching the settlement (bots never block).
func (cr *chessRoom) allSettleDone() bool {
	for seat := range cr.match.P {
		if !cr.settleDone[seat] && cr.conn(seat) != nil {
			return false
		}
	}
	return true
}

func (cr *chessRoom) advanceIfReady() {
	switch cr.phase {
	case pb.MatchPhase_MATCH_PHASE_SETTLE:
		if cr.allSettleDone() {
			if room, err := service.GetRoomManager().GetRoom(cr.roomID); err == nil {
				cr.advance(room)
			}
		}
	case pb.MatchPhase_MATCH_PHASE_FAMILY, pb.MatchPhase_MATCH_PHASE_PREP, pb.MatchPhase_MATCH_PHASE_VAULT:
		if cr.allReady() {
			if room, err := service.GetRoomManager().GetRoom(cr.roomID); err == nil {
				cr.advance(room)
			}
		}
	}
}

// advance leaves the current phase (on timeout or when everyone is ready).
func (cr *chessRoom) advance(room *types.RoomInfo) {
	m := cr.match
	timers := m.Config().Rules.Timers
	switch cr.phase {
	case pb.MatchPhase_MATCH_PHASE_FAMILY:
		if err := m.Begin(); err != nil {
			log.Printf("room %s: begin match: %v", cr.roomID, err)
			return
		}
		cr.enterRound()
	case pb.MatchPhase_MATCH_PHASE_PREP:
		res, err := m.Settle()
		if err != nil {
			log.Printf("room %s: settle: %v", cr.roomID, err)
			return
		}
		for seat := range m.P {
			cr.lastScore[seat] = res.Scores[seat].Score
		}
		cr.settled = res.Round
		cr.settleDone = [2]bool{}
		cr.phase = pb.MatchPhase_MATCH_PHASE_SETTLE
		cr.schedule(phaseSeconds(timers.SettleSeconds))
		for seat := range m.P {
			cr.sendTo(seat, roundResultMessage(cr, res, seat))
		}
		cr.broadcast()
	case pb.MatchPhase_MATCH_PHASE_SETTLE:
		switch m.Phase {
		case rules.PhaseOver:
			cr.finish(room)
		case rules.PhaseVault:
			cr.enterVault()
		default:
			cr.enterRound()
		}
	case pb.MatchPhase_MATCH_PHASE_VAULT:
		if err := m.CloseVault(); err != nil {
			log.Printf("room %s: close vault: %v", cr.roomID, err)
			return
		}
		cr.enterRound()
	}
}

// enterRound starts the round the engine just opened.
// TEMPORARY: the omen phase is resolved at once (no rolls) until the omen phase is wired up.
func (cr *chessRoom) enterRound() {
	if cr.match.Phase == rules.PhaseOmen {
		_ = cr.match.ResolveOmen()
	}
	cr.enterPrep()
}

func (cr *chessRoom) enterPrep() {
	cr.phase = pb.MatchPhase_MATCH_PHASE_PREP
	cr.schedule(phaseSeconds(cr.match.Config().Rules.Timers.PrepSeconds))
	for seat, b := range cr.bots {
		if b != nil {
			b.Prep(cr.match, cr.match.P[seat])
		}
	}
	cr.broadcast()
	cr.advanceIfReady()
}

func (cr *chessRoom) enterVault() {
	cr.phase = pb.MatchPhase_MATCH_PHASE_VAULT
	cr.schedule(phaseSeconds(cr.match.Config().Rules.Timers.VaultSeconds))
	for seat, b := range cr.bots {
		if b != nil {
			b.PickVault(cr.match, cr.match.P[seat])
		}
	}
	cr.broadcast()
	cr.advanceIfReady()
}

func (cr *chessRoom) finish(room *types.RoomInfo) {
	cr.stop()
	cr.phase = pb.MatchPhase_MATCH_PHASE_OVER
	cr.deadline = time.Time{}
	for seat := range cr.match.P {
		cr.sendTo(seat, matchOverMessage(cr, seat))
	}
	chessRooms.Delete(cr.roomID)
	finishGame(room)
}

// ---------- player actions ----------

// GameActionRequest is one client action inside a match.
type GameActionRequest struct {
	Player   string
	ClientID string
	Seq      uint32
	Action   *pb.GameAction
}

var errNotInMatch = errors.New("player is not in a match")

// HandleGameAction validates and applies an action, then pushes fresh state to both players.
func HandleGameAction(req GameActionRequest) {
	info, ok := service.GetConnectionManager().GetConnectionByClientID(req.ClientID)
	roomID := ""
	if ok {
		roomID = info.GetGameRoom()
	}
	if roomID == "" {
		notifyClientError(req.ClientID, req.Seq, pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE, config.Msg(config.MsgNotInGame))
		return
	}
	err := withRoom(roomID, func(room *types.RoomInfo) error {
		cr := getChessRoom(roomID)
		if cr == nil {
			return errNotInMatch
		}
		seat := -1
		for i, p := range cr.match.P {
			if p.ID == req.Player {
				seat = i
			}
		}
		if seat < 0 {
			return errNotInMatch
		}
		if err := cr.apply(seat, req.Action); err != nil {
			return err
		}
		if cr.match.Phase == rules.PhaseOver {
			cr.finish(room) // surrender: MatchOver goes to both players
			return nil
		}
		cr.sendTo(seat, matchStateMessage(cr, seat, req.Seq))
		cr.sendTo(1-seat, matchStateMessage(cr, 1-seat, 0))
		cr.advanceIfReady()
		return nil
	})
	if err != nil {
		code, text := actionError(err)
		notifyClientError(req.ClientID, req.Seq, code, text)
	}
}

func (cr *chessRoom) apply(seat int, a *pb.GameAction) error {
	m := cr.match
	p := m.P[seat]
	if _, ok := a.GetAction().(*pb.GameAction_SettleDone); ok {
		// a late ack (the cap already ended the pause) is harmless, so never an error
		if cr.phase == pb.MatchPhase_MATCH_PHASE_SETTLE {
			cr.settleDone[seat] = true
		}
		return nil
	}
	allowed := map[pb.MatchPhase]bool{}
	switch a.GetAction().(type) {
	case *pb.GameAction_PickFamily:
		allowed[pb.MatchPhase_MATCH_PHASE_FAMILY] = true
	case *pb.GameAction_PickVault, *pb.GameAction_SkipVault:
		allowed[pb.MatchPhase_MATCH_PHASE_VAULT] = true
	case *pb.GameAction_SellRelic:
		allowed[pb.MatchPhase_MATCH_PHASE_PREP] = true
		allowed[pb.MatchPhase_MATCH_PHASE_VAULT] = true
	case *pb.GameAction_SetReady:
		allowed[pb.MatchPhase_MATCH_PHASE_FAMILY] = true
		allowed[pb.MatchPhase_MATCH_PHASE_PREP] = true
	case *pb.GameAction_Surrender:
		for _, ph := range []pb.MatchPhase{pb.MatchPhase_MATCH_PHASE_FAMILY, pb.MatchPhase_MATCH_PHASE_PREP,
			pb.MatchPhase_MATCH_PHASE_SETTLE, pb.MatchPhase_MATCH_PHASE_VAULT} {
			allowed[ph] = true
		}
	default:
		allowed[pb.MatchPhase_MATCH_PHASE_PREP] = true
	}
	if !allowed[cr.phase] {
		return rules.ErrWrongPhase
	}
	switch act := a.GetAction().(type) {
	case *pb.GameAction_PickFamily:
		return m.PickFamily(p, int(act.PickFamily.GetOfferIndex()))
	case *pb.GameAction_BuyCard:
		_, err := m.BuyCard(p, int(act.BuyCard.GetShopSlot()))
		return err
	case *pb.GameAction_SellCard:
		return m.SellCard(p, act.SellCard.GetCardUid())
	case *pb.GameAction_RefreshShop:
		return m.Refresh(p)
	case *pb.GameAction_BuyXp:
		return m.BuyXP(p)
	case *pb.GameAction_BuyRelic:
		return m.BuyRelic(p, int(act.BuyRelic.GetTradeChoice()))
	case *pb.GameAction_SellRelic:
		return m.SellRelic(p, int(act.SellRelic.GetRelicSlot()))
	case *pb.GameAction_Arrange:
		var order []int
		if ro := act.Arrange.GetRelicOrder(); len(ro) > 0 {
			order = make([]int, len(ro))
			for i, v := range ro {
				order[i] = int(v)
			}
		}
		return m.Arrange(p, act.Arrange.GetBoardUids(), order)
	case *pb.GameAction_LockShop:
		return m.SetLock(p, act.LockShop.GetLocked())
	case *pb.GameAction_SetReady:
		if cr.phase == pb.MatchPhase_MATCH_PHASE_FAMILY && p.Family < 0 && act.SetReady.GetReady() {
			return rules.ErrBadChoice // pick a family instead
		}
		return m.SetReady(p, act.SetReady.GetReady())
	case *pb.GameAction_PickVault:
		pv := act.PickVault
		return m.PickVault(p, int(pv.GetOfferIndex()), int(pv.GetSellRelicSlot()), int(pv.GetTradeChoice()))
	case *pb.GameAction_SkipVault:
		return m.SkipVault(p)
	case *pb.GameAction_Surrender:
		return m.Surrender(p)
	}
	return rules.ErrBadChoice
}

func actionError(err error) (pb.ErrorCode, string) {
	switch {
	case errors.Is(err, rules.ErrNotEnoughGold):
		return pb.ErrorCode_ERROR_CODE_NOT_ENOUGH_GOLD, config.Msg(config.MsgNotEnoughGold)
	case errors.Is(err, rules.ErrBenchFull):
		return pb.ErrorCode_ERROR_CODE_BENCH_FULL, config.Msg(config.MsgBenchFull)
	case errors.Is(err, rules.ErrRelicSlotsFull):
		return pb.ErrorCode_ERROR_CODE_RELIC_SLOTS_FULL, config.Msg(config.MsgRelicSlotsFull)
	case errors.Is(err, rules.ErrMaxLevel):
		return pb.ErrorCode_ERROR_CODE_MAX_LEVEL, config.Msg(config.MsgMaxLevel)
	case errors.Is(err, rules.ErrWrongPhase), errors.Is(err, errNotInMatch), errors.Is(err, rules.ErrAlreadyChosen):
		return pb.ErrorCode_ERROR_CODE_INVALID_GAME_STATE, config.Msg(config.MsgWrongPhase)
	case errors.Is(err, rules.ErrNoSuchItem), errors.Is(err, rules.ErrBadLayout), errors.Is(err, rules.ErrBadChoice):
		return pb.ErrorCode_ERROR_CODE_INVALID_CHOICE, config.Msg(config.MsgInvalidChoice)
	}
	log.Printf("game action failed: %v", err)
	return pb.ErrorCode_ERROR_CODE_INTERNAL, config.Msg(config.MsgInternalError)
}

// ---------- presence hooks ----------

// chessResend sends the full match view to a player who (re)joined. Call inside withRoom.
func chessResend(roomID, username string) bool {
	cr := getChessRoom(roomID)
	if cr == nil {
		return false
	}
	for seat, p := range cr.match.P {
		if p.ID == username {
			cr.sendTo(seat, matchConfigMessage(cr.roomID, cr.match))
			cr.sendTo(seat, matchStateMessage(cr, seat, 0))
			cr.sendTo(1-seat, matchStateMessage(cr, 1-seat, 0))
			return true
		}
	}
	return false
}

// chessPresenceChanged lets a disconnected player stop blocking the phase.
func chessPresenceChanged(roomID string) {
	_ = withRoom(roomID, func(room *types.RoomInfo) error {
		if cr := getChessRoom(roomID); cr != nil {
			cr.broadcast()
			cr.advanceIfReady()
		}
		return nil
	})
}
