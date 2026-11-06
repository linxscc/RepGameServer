package logic

import (
	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/config"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/pb"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/types"
	"fmt"
	"log"
	"sync"
	"time"
)

// GameStartProcessor 游戏开始处理器
type GameStartProcessor struct{}

// NewGameStartProcessor 创建游戏开始处理器
func NewGameStartProcessor() *GameStartProcessor {
	return &GameStartProcessor{}
}

// ProcessGameStart 处理游戏开始事件：尝试用已准备的真人玩家凑满一个房间
func (g *GameStartProcessor) ProcessGameStart(eventData interface{}) error {
	if _, ok := eventData.(*events.EventData); !ok {
		return fmt.Errorf("invalid event data type")
	}
	return g.performMatchmaking()
}

// CreateGameRoom 创建游戏房间
func (g *GameStartProcessor) CreateGameRoom(roomName string, maxPlayers int) (*types.RoomInfo, error) {
	roomManager := service.GetRoomManager()

	room, err := roomManager.CreateRoom(roomName, maxPlayers)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %v", err)
	}

	return room, nil
}

// CleanupRoom 清理房间资源
func (g *GameStartProcessor) CleanupRoom(roomID string) {
	roomManager := service.GetRoomManager()
	roomManager.RemoveRoom(roomID)
}

// InitializeRoomCardPools 初始化房间卡牌池
func (g *GameStartProcessor) InitializeRoomCardPools(room *types.RoomInfo) error {
	// 获取卡牌池
	level1Cards := cards.GetLevel1Cards()
	level2Cards := cards.GetLevel2Cards()
	level3Cards := cards.GetLevel3Cards()

	roomManager := service.GetRoomManager()
	err := roomManager.InitCardPool(room.RoomID, level1Cards, level2Cards, level3Cards)
	if err != nil {
		return fmt.Errorf("failed to initialize card pools: %v", err)
	}

	return nil
}

// AddPlayersToRoom 添加玩家到房间
func (g *GameStartProcessor) AddPlayersToRoom(room *types.RoomInfo, players []*types.ClientInfo, connManager *service.ConnectionManager) error {
	roomManager := service.GetRoomManager()

	for _, player := range players {
		if player.Username == "" {
			continue // 跳过未登录的玩家
		}
		err := roomManager.JoinRoom(room.RoomID, player.Username)
		if err != nil {
			return fmt.Errorf("failed to add player %s to room: %v", player.Username, err)
		}

		// 设置玩家状态为游戏中
		connManager.SetPlayerStatus(player.ClientID, types.StatusInGame)
		connManager.SetPlayerGameRoom(player.ClientID, room.RoomID)
	}

	return nil
}

// DealInitialCardsToAllPlayers 为房间内所有玩家分发初始手牌
func (g *GameStartProcessor) DealInitialCardsToAllPlayers(room *types.RoomInfo) error {
	for _, username := range room.PlayerNames() {
		err := g.DealInitialCardsToPlayer(room, username)
		if err != nil {
			return fmt.Errorf("failed to deal cards to player %s: %v", username, err)
		}
	}

	return nil
}

// DealInitialCardsToPlayer 为指定玩家分发初始手牌
func (g *GameStartProcessor) DealInitialCardsToPlayer(room *types.RoomInfo, username string) error {
	// 从房间的1级卡牌池中抽取开局手牌（数量见 game_rules.yaml）
	initCards, err := room.DrawRandomCardsFromLevel1Pool(config.Game().Rules.Player.InitialHandSize)
	if err != nil {
		return err
	}
	return room.SetPlayerHand(username, initCards)
}

// InitializePlayersHealthAndNotify 初始化玩家血量与先后手，并向真人玩家发送开局状态。
// participants 的顺序决定先后手：第一个为先手。机器人没有网络连接，只设置状态不发送消息。
func (g *GameStartProcessor) InitializePlayersHealthAndNotify(room *types.RoomInfo, participants []string) error {
	room.UpdateRoomStatus(types.RoomStatusPlaying)

	for i, username := range participants {
		if err := room.SetPlayerHealth(username, room.InitialHealth); err != nil {
			return fmt.Errorf("failed to set initial health for player %s: %v", username, err)
		}
		round := types.RoundWaiting
		if i == 0 {
			round = types.RoundCurrent
		}
		if err := room.SetPlayerRound(username, round); err != nil {
			return err
		}
	}

	for _, username := range participants {
		sendRoomState(room, username, pb.GameStateReason_GAME_STATE_REASON_GAME_STARTED)
	}

	// 开始第一回合（真人计时 / 机器人出牌）
	return GlobalRoomTimerProcessor.StartRoomTimer(room.RoomID)
}

// matchmakingMu 串行化所有匹配操作，避免同一名玩家同时被真人匹配和机器人匹配选中
var matchmakingMu sync.Mutex

// performMatchmaking 已准备的真人玩家凑满一个房间时开局
func (g *GameStartProcessor) performMatchmaking() error {
	matchmakingMu.Lock()
	defer matchmakingMu.Unlock()

	needed := config.Game().Rules.Match.PlayersPerRoom
	readyPlayers := service.GetConnectionManager().GetConnectionsByStatus(types.StatusReady)
	if len(readyPlayers) < needed {
		return nil // 人数不足，等待更多玩家或机器人匹配
	}
	return g.startMatch(readyPlayers[:needed], nil)
}

// StartBotMatch 玩家等待超时后，为其安排机器人对手。
// 如果玩家已经匹配到真人或已离开，则什么都不做。
func (g *GameStartProcessor) StartBotMatch(clientID string) error {
	matchmakingMu.Lock()
	defer matchmakingMu.Unlock()

	player, exists := service.GetConnectionManager().GetConnectionByClientID(clientID)
	if !exists || player == nil || player.GetStatus() != types.StatusReady || player.Username == "" {
		return nil
	}

	bots := make([]string, 0, config.Game().Rules.Match.PlayersPerRoom-1)
	for len(bots) < cap(bots) {
		bots = append(bots, NewBotName())
	}
	return g.startMatch([]*types.ClientInfo{player}, bots)
}

// startMatch 创建房间、发牌并开局。真人玩家在前（先手），机器人在后。
// 调用方必须持有 matchmakingMu。
func (g *GameStartProcessor) startMatch(humans []*types.ClientInfo, bots []string) error {
	connManager := service.GetConnectionManager()

	room, err := g.CreateGameRoom(fmt.Sprintf("Game Room %d", time.Now().Unix()), config.Game().Rules.Match.PlayersPerRoom)
	if err != nil {
		return fmt.Errorf("failed to create room: %v", err)
	}

	if err := g.InitializeRoomCardPools(room); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to initialize room card pools: %v", err)
	}

	if err := g.AddPlayersToRoom(room, humans, connManager); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to add players to room: %v", err)
	}
	for _, bot := range bots {
		if err := service.GetRoomManager().JoinRoom(room.RoomID, bot); err != nil {
			g.CleanupRoom(room.RoomID)
			return fmt.Errorf("failed to add bot %s to room: %v", bot, err)
		}
	}

	if err := g.DealInitialCardsToAllPlayers(room); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to deal initial cards: %v", err)
	}

	participants := make([]string, 0, len(humans)+len(bots))
	for _, h := range humans {
		participants = append(participants, h.Username)
	}
	participants = append(participants, bots...)

	// 开局后客户端即可对该房间发起操作，从这里开始与其他房间操作串行
	err = withRoom(room.RoomID, func(room *types.RoomInfo) error {
		return g.InitializePlayersHealthAndNotify(room, participants)
	})
	if err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to initialize players and send notifications: %v", err)
	}
	if len(bots) > 0 {
		log.Printf("bot match started: room=%s players=%v", room.RoomID, participants)
	}
	return nil
}
