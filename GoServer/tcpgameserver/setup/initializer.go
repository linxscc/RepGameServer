package setup

import (
	"GoServer/tcpgameserver/cards"
	"GoServer/tcpgameserver/events"
	"GoServer/tcpgameserver/tools"
	"log"
)

// InitializeServer 初始化TCP服务器的所有必要组件
func InitializeServer() {
	log.Println("Starting server initialization...")

	// 1. 初始化事件系统
	initializeEventSystem()

	// 2. 加载响应码
	loadResponseCodes()

	// 3. 初始化卡牌池
	initializeCardPool()

	// 4. 发布服务器启动完成事件
	publishServerStartEvent()

	log.Println("Server initialization completed successfully")
}

// 初始化事件系统
func initializeEventSystem() {
	log.Println("Initializing event system...")
	events.InitializeEventSystem()
}

// 加载响应码配置文件
func loadResponseCodes() {
	log.Println("Loading response codes from database...")

	// 从数据库加载响应码到内存
	if err := tools.LoadResponseCodes(); err != nil {
		log.Printf("Failed to load response codes from database: %v", err)
	}
}

// 初始化卡牌池
func initializeCardPool() {
	if err := cards.InitCardPool(); err != nil {
		log.Printf("Failed to initialize card pool: %v", err)
	}
}

// 发布服务器启动完成事件
func publishServerStartEvent() {
	serverStartData := events.CreateSystemEventData(events.EventSystemStart, "TCP Game Server started and ready to accept connections")
	serverStartData.AddData("server_port", 9060)
	serverStartData.AddData("server_type", "tcp_game_server")
	events.Publish(events.EventSystemStart, serverStartData)
}
