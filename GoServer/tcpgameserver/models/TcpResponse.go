package models

// TCP响应结构体
type TcpResponse struct {
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}
