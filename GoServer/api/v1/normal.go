// 放入 api/v1/hello.go
package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type NormalReq struct {
	g.Meta `path:"/normal" method:"post" summary:"Normal Post Request"`
	Name   string `v:"required" dc:"Your name"`
}

type NormalRes struct {
	CODE    int    `json:"code"`
	Message string `json:"message"`
}
