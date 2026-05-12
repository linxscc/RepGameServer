package v1

import "github.com/gogf/gf/v2/frame/g"

type GetProfileInfoReq struct {
	g.Meta `path:"/profile-info" method:"get" summary:"获取个人信息"`
}

type GetProfileInfoRes = ProfileInfoResponse

type ProfileInfoResponse struct {
	FullName  string `json:"fullName"`
	Title     string `json:"title"`
	Tagline   string `json:"tagline"`
	AboutText string `json:"aboutText"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Languages string `json:"languages"`
}
