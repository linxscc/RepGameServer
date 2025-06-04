package models

// ResponseInfo 响应信息结构体
type ResponseInfo struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	ResponseKey string `json:"responsekey"`
	Message     string `json:"message"`
}

// TcpRequest TCP请求结构体，用于接收客户端数据
type TcpRequest struct {
	Code        string      `json:"code"`
	Message     string      `json:"message"`
	ResponseKey string      `json:"responsekey"`
	Data        interface{} `json:"data"`
}

// TcpResponse TCP响应结构体，用于发送给客户端的数据
type TcpResponse struct {
	Code        string      `json:"code"`
	Message     string      `json:"message"`
	ResponseKey string      `json:"responsekey"`
	Data        interface{} `json:"data"`
}
