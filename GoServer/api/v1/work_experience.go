package v1

import "github.com/gogf/gf/v2/frame/g"

type GetWorkExperienceReq struct {
	g.Meta `path:"/work-experience" method:"get" summary:"获取工作经验数据"`
}

type GetWorkExperienceRes = WorkExperienceDataResponse

type WorkExperienceDataResponse struct {
	Hero       map[string]string `json:"hero"`
	Features24 map[string]string `json:"features24"`
	Features25 map[string]string `json:"features25"`
	Steps2     map[string]string `json:"steps2"`
	Contact10  map[string]string `json:"contact10"`
}
