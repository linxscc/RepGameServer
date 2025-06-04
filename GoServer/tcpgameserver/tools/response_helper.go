package tools

import (
	"GoServer/tcpgameserver/models"
)

// ResponseHelper 响应助手，专门用于创建返回消息
type ResponseHelper struct{}

// CreateTcpResponse 根据ID创建TcpResponse，合并data参数
func (h *ResponseHelper) CreateSuccessTcpResponse(id int, data interface{}) *models.TcpResponse {
	response, exists := GetResponseByID(id)
	if !exists {
		defaultResponse, _ := GetResponseByID(5001)
		return &models.TcpResponse{
			Code:        defaultResponse.Code,
			Message:     defaultResponse.Message,
			ResponseKey: defaultResponse.ResponseKey,
			Data:        data,
		}
	}

	return &models.TcpResponse{
		Code:        response.Code,
		Message:     response.Message,
		ResponseKey: response.ResponseKey,
		Data:        data,
	}
}

func (h *ResponseHelper) CreateErrorTcpResponse(id int) *models.TcpResponse {
	response, exists := GetResponseByID(id)
	if !exists {
		defaultResponse, _ := GetResponseByID(5001)
		return &models.TcpResponse{
			Code:        defaultResponse.Code,
			Message:     defaultResponse.Message,
			ResponseKey: defaultResponse.ResponseKey,
			Data:        "",
		}
	}

	return &models.TcpResponse{
		Code:        response.Code,
		Message:     response.Message,
		ResponseKey: response.ResponseKey,
		Data:        "",
	}
}

// NewResponseHelper 创建响应助手实例
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// 全局响应助手实例
var GlobalResponseHelper = NewResponseHelper()
