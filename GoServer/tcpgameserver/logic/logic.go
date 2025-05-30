package logic

import (
	"GoServer/tcpgameserver/models"
	"log"
)

// 处理 HealthUpdate 列表的业务逻辑
func HandleHealthUpdateList(healthUpdates []models.HealthUpdate) {
	// 这里可以实现具体的业务逻辑，比如保存、广播、计算等
	log.Printf("[Logic] Handling HealthUpdate list: %+v", healthUpdates)
	// ...业务处理...
}

// 你可以继续添加更多类型的消息处理函数
