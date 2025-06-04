package tools

import (
	"GoServer/tcpgameserver/models"
	"GoServer/tcpgameserver/service"
	"fmt"
	"log"
	"sync"
)

// ResponseCodeManager 响应码管理器
type ResponseCodeManager struct {
	codes map[int]models.ResponseInfo
	mutex sync.RWMutex
}

var (
	responseManager *ResponseCodeManager
	once            sync.Once
)

// GetResponseCodeManager 获取响应码管理器单例
func GetResponseCodeManager() *ResponseCodeManager {
	once.Do(func() {
		responseManager = &ResponseCodeManager{
			codes: make(map[int]models.ResponseInfo),
		}
	})
	return responseManager
}

// LoadResponseCodes 从数据库加载响应码
func LoadResponseCodes() error {
	manager := GetResponseCodeManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// 从数据库获取所有响应码
	responseInfos, err := service.GetAllResponseInfo()
	if err != nil {
		return fmt.Errorf("failed to load response codes from database: %v", err)
	}

	// 清空现有数据
	manager.codes = make(map[int]models.ResponseInfo)

	// 将数据存储到内存中
	for _, info := range responseInfos {
		manager.codes[info.ID] = info
	}

	return nil
}

// GetResponseByID 根据ID获取响应码信息
func GetResponseByID(id int) (*models.ResponseInfo, bool) {
	manager := GetResponseCodeManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	response, exists := manager.codes[id]
	if !exists {
		return nil, false
	}
	return &response, true
}

// GetResponseByCode 根据代码获取响应码信息
func GetResponseByCode(code string) (*models.ResponseInfo, bool) {
	manager := GetResponseCodeManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for _, response := range manager.codes {
		if response.Code == code {
			return &response, true
		}
	}
	return nil, false
}

// GetAllResponseCodes 获取所有响应码信息
func GetAllResponseCodes() []models.ResponseInfo {
	manager := GetResponseCodeManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	responses := make([]models.ResponseInfo, 0, len(manager.codes))
	for _, response := range manager.codes {
		responses = append(responses, response)
	}
	return responses
}

// ReloadResponseCodes 重新加载响应码（用于动态更新）
func ReloadResponseCodes() error {
	log.Println("Reloading response codes from database...")
	return LoadResponseCodes()
}

// GetResponseCodeCount 获取响应码数量
func GetResponseCodeCount() int {
	manager := GetResponseCodeManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return len(manager.codes)
}
