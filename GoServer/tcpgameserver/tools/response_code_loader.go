package tools

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"GoServer/tcpgameserver/models"
)

type ResponseCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var (
	responseCodeMap  map[string]ResponseCode
	responseCodeOnce sync.Once
)

// LoadResponseCodes 读取 response_codes.json 并缓存到内存
func LoadResponseCodes(path string) {
	responseCodeOnce.Do(func() {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read response_codes.json: %v", err)
		}
		responseCodeMap = make(map[string]ResponseCode)
		if err := json.Unmarshal(data, &responseCodeMap); err != nil {
			log.Fatalf("Failed to parse response_codes.json: %v", err)
		}
		log.Println("Response codes loaded.")
	})
}

// GetResponseCodeByID 根据编号获取响应内容
func GetResponseCodeByID(id string, data interface{}) models.TcpResponse {
	code, ok := responseCodeMap[id]
	if !ok {
		return models.TcpResponse{Code: 400, Message: "unknown error type", Data: data}
	}
	return models.TcpResponse{Code: code.Code, Message: code.Message, Data: data}
}
