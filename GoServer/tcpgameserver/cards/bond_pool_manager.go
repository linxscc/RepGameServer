package cards

import (
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"fmt"
	"log"
	"sync"
)

// BondPoolManager 羁绊池管理器
type BondPoolManager struct {
	bonds map[int]*models.BondModel // 羁绊ID -> 羁绊模型映射
	mutex sync.RWMutex              // 读写锁
}

var (
	bondPool     *BondPoolManager
	bondPoolOnce sync.Once
)

// GetBondPoolManager 获取羁绊池管理器单例
func GetBondPoolManager() *BondPoolManager {
	bondPoolOnce.Do(func() {
		bondPool = &BondPoolManager{
			bonds: make(map[int]*models.BondModel),
		}
	})
	return bondPool
}

// InitBondPool 初始化羁绊池
func InitBondPool() error {
	manager := GetBondPoolManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	log.Println("Initializing bond pool...")

	// 从数据库获取所有羁绊数据
	bonds, err := service.GetAllBonds()
	if err != nil {
		return fmt.Errorf("failed to load bonds from database: %v", err)
	}

	// 清空现有羁绊池
	manager.bonds = make(map[int]*models.BondModel)

	// 加载羁绊数据到池中
	for _, bond := range bonds {
		manager.bonds[bond.ID] = &bond
	}
	log.Printf("Bond pool initialized with %d bonds", len(manager.bonds))
	return nil
}

// GetBondByID 根据ID获取羁绊
func (bpm *BondPoolManager) GetBondByID(bondID int) (*models.BondModel, bool) {
	bpm.mutex.RLock()
	defer bpm.mutex.RUnlock()

	bond, exists := bpm.bonds[bondID]
	return bond, exists
}

// GetAllBonds 获取所有羁绊
func (bpm *BondPoolManager) GetAllBonds() map[int]*models.BondModel {
	bpm.mutex.RLock()
	defer bpm.mutex.RUnlock()

	// 创建副本避免外部修改
	result := make(map[int]*models.BondModel)
	for id, bond := range bpm.bonds {
		result[id] = bond
	}
	return result
}

// GetBondsByLevel 根据等级获取羁绊列表
func (bpm *BondPoolManager) GetBondsByLevel(level int) []*models.BondModel {
	bpm.mutex.RLock()
	defer bpm.mutex.RUnlock()

	var result []*models.BondModel
	for _, bond := range bpm.bonds {
		if bond.Level == level {
			result = append(result, bond)
		}
	}
	return result
}

// GetBondsByCardName 根据卡牌名称获取相关的羁绊列表
func (bpm *BondPoolManager) GetBondsByCardName(cardName string) []*models.BondModel {
	bpm.mutex.RLock()
	defer bpm.mutex.RUnlock()

	var result []*models.BondModel
	for _, bond := range bpm.bonds {
		for _, card := range bond.CardNames {
			if card == cardName {
				result = append(result, bond)
				break
			}
		}
	}
	return result
}

// ReloadBonds 重新加载羁绊数据
func (bpm *BondPoolManager) ReloadBonds() error {
	return InitBondPool()
}

// GetBondStats 获取羁绊池统计信息
func (bpm *BondPoolManager) GetBondStats() map[string]int {
	bpm.mutex.RLock()
	defer bpm.mutex.RUnlock()

	stats := map[string]int{
		"total":   len(bpm.bonds),
		"level_1": 0,
		"level_2": 0,
		"level_3": 0,
	}

	for _, bond := range bpm.bonds {
		switch bond.Level {
		case 1:
			stats["level_1"]++
		case 2:
			stats["level_2"]++
		case 3:
			stats["level_3"]++
		}
	}

	return stats
}
