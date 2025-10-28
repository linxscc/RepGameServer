package logic

import (
	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"GoServer/tcpgameserver/tools"
	"GoServer/tcpgameserver/types"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// GameStartProcessor 游戏开始处理器
type GameStartProcessor struct{}

// NewGameStartProcessor 创建游戏开始处理器
func NewGameStartProcessor() *GameStartProcessor {
	return &GameStartProcessor{}
}

// ProcessGameStart 处理游戏开始事件
func (g *GameStartProcessor) ProcessGameStart(eventData interface{}) error {
	// 类型断言获取事件数据
	if data, ok := eventData.(*events.EventData); ok {
		// 检查是否是匹配触发的游戏开始事件
		if triggerSource, exists := data.GetString("trigger_source"); exists && triggerSource == "user_ready_handler" {

			// 执行匹配逻辑
			err := g.performMatchmaking()
			if err != nil {
				log.Printf("匹配失败: %v", err)
				return err
			} else {
				log.Printf("匹配成功，游戏已开始")
			}
			return nil
		}

		return nil
	}
	return fmt.Errorf("invalid event data type")
}

// CreateGameRoom 创建游戏房间
func (g *GameStartProcessor) CreateGameRoom(roomName string, maxPlayers int) (*types.RoomInfo, error) {
	roomManager := service.GetRoomManager()

	room, err := roomManager.CreateRoom(roomName, maxPlayers)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %v", err)
	}

	log.Printf("Created new game room: %s (max players: %d)", room.RoomID, maxPlayers)
	return room, nil
}

// CleanupRoom 清理房间资源
func (g *GameStartProcessor) CleanupRoom(roomID string) {
	roomManager := service.GetRoomManager()
	roomManager.RemoveRoom(roomID)
	log.Printf("Cleaned up room: %s", roomID)
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
	for username := range room.Players {
		err := g.DealInitialCardsToPlayer(room, username)
		if err != nil {
			return fmt.Errorf("failed to deal cards to player %s: %v", username, err)
		}
	}

	return nil
}

// DealInitialCardsToPlayer 为指定玩家分发初始手牌
func (g *GameStartProcessor) DealInitialCardsToPlayer(room *types.RoomInfo, username string) error {
	roomManager := service.GetRoomManager()
	// 获取玩家信息
	player, exists := room.Players[username]
	if !exists {
		return fmt.Errorf("player %s not found in room", username)
	}

	// 初始化玩家手牌
	// 从房间的1级卡牌池中抽取6张卡牌
	var initCards []models.Card
	initCards, _ = roomManager.InitPlayerHandCard(room.RoomID, 6)
	player.HandCards = initCards

	log.Printf("Dealt %d cards to player %s", len(player.HandCards), username)
	return nil
}

// InitializePlayersHealthAndNotify 初始化玩家信息并发送游戏开始通知
func (g *GameStartProcessor) InitializePlayersHealthAndNotify(room *types.RoomInfo, players []*types.ClientInfo, connManager *service.ConnectionManager) error {
	// 收集所有玩家信息用于构建游戏信息
	playerUsernames := make([]string, 0, len(players))
	for _, player := range players {
		if player.Username != "" {
			playerUsernames = append(playerUsernames, player.Username)
		}
	}

	// 设置房间状态为进行中
	room.Status = "playing"

	// 获取所有羁绊数据
	// bondPoolManager := cards.GetBondPoolManager()
	roomManager := service.GetRoomManager()
	// allBonds := bondPoolManager.GetAllBonds()
	// 为每个玩家设置初始信息并发送游戏开始消息
	for i, player := range players {
		if player.Username == "" {
			continue
		}

		// 设置玩家初始血量
		err := roomManager.SetPlayerHealth(room.RoomID, player.Username, room.Players[player.Username].MaxHealth)
		if err != nil {
			return fmt.Errorf("failed to set initial health for player %s: %v", player.Username, err)
		}
		// 设置先手回合
		if isFirstPlayer := (i == 0); isFirstPlayer {
			roomManager.SetPlayerRound(room.RoomID, player.Username, "current")
		} else {
			roomManager.SetPlayerRound(room.RoomID, player.Username, "waiting")
		}

		// 再发送游戏开始通知 (5001)
		err = g.SendGameStartNotification(room, player, connManager)
		if err != nil {
			return fmt.Errorf("failed to send game start notification to player %s: %v", player.Username, err)
		}
	}

	GlobalRoomTimerProcessor.StartRoomTimer(room.RoomID)

	return nil
}

// SendGameStartNotification 发送游戏开始通知
func (g *GameStartProcessor) SendGameStartNotification(room *types.RoomInfo, player *types.ClientInfo, connManager *service.ConnectionManager) error {
	// 获取玩家游戏信息
	roomManager := service.GetRoomManager()
	playerInfo, err := roomManager.GetPlayerGameInfo(room.RoomID, player.Username)
	if err != nil {
		return fmt.Errorf("failed to get player game info for %s: %v", player.Username, err)
	}

	// 发送游戏开始消息
	response := tools.GlobalResponseHelper.CreateSuccessTcpResponse(5001, playerInfo)

	// 获取玩家连接并发送消息
	clientInfo, exists := connManager.GetConnectionByClientID(player.ClientID)
	if !exists || clientInfo == nil || clientInfo.Conn == nil {
		return fmt.Errorf("player connection not found or invalid")
	}

	g.sendTCPResponse(clientInfo.Conn, response)

	return nil
}

// performMatchmaking 执行匹配逻辑
func (g *GameStartProcessor) performMatchmaking() error {
	// 获取连接管理器
	connManager := service.GetConnectionManager()

	// 获取准备就绪的玩家
	readyPlayers := connManager.GetConnectionsByStatus(types.StatusReady)
	if len(readyPlayers) < 2 {
		return fmt.Errorf("insufficient ready players: %d", len(readyPlayers))
	}

	log.Printf("Found %d ready players, starting matchmaking", len(readyPlayers))

	// 选择前两位玩家进行匹配
	selectedPlayers := readyPlayers[:2]

	// 创建新房间
	room, err := g.CreateGameRoom(fmt.Sprintf("Game Room %d", time.Now().Unix()), 2)
	if err != nil {
		return fmt.Errorf("failed to create room: %v", err)
	}

	// 初始化房间卡牌池
	if err := g.InitializeRoomCardPools(room); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to initialize room card pools: %v", err)
	}

	// 添加玩家到房间
	if err := g.AddPlayersToRoom(room, selectedPlayers, connManager); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to add players to room: %v", err)
	}

	// 为所有玩家分发初始手牌
	if err := g.DealInitialCardsToAllPlayers(room); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to deal initial cards: %v", err)
	}

	// 设置玩家初始信息并发送游戏开始通知
	if err := g.InitializePlayersHealthAndNotify(room, selectedPlayers, connManager); err != nil {
		g.CleanupRoom(room.RoomID)
		return fmt.Errorf("failed to initialize players and send notifications: %v", err)
	}

	log.Printf("Successfully matched players and started game in room: %s", room.RoomID)
	return nil
}

// sendTCPResponse 发送TCP响应消息
func (g *GameStartProcessor) sendTCPResponse(conn net.Conn, resp *models.TcpResponse) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	_, err = conn.Write(jsonBytes)
	if err != nil {
		log.Printf("Failed to write response to connection: %v", err)
	}
}
