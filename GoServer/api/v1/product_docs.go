package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GenerateProductDocsReq 生成产品出口文档请求
type GenerateProductDocsReq struct {
	g.Meta             `path:"/product/generate-docs" method:"post" summary:"生成产品出口文档"`
	CompanyName        string `json:"companyName" v:"required" dc:"公司名称"`
	ProductName        string `json:"productName" v:"required" dc:"产品名称"`
	ProductCategory    string `json:"productCategory" v:"required" dc:"产品类别"`
	TargetCountry      string `json:"targetCountry" v:"required" dc:"目标国家"`
	ProductDescription string `json:"productDescription" v:"required" dc:"产品描述"`
	Language           string `json:"language" dc:"语言(zh/en)，默认zh"`
}

// GenerateProductDocsRes 生成产品出口文档响应
type GenerateProductDocsRes struct {
	Code    int              `json:"code" dc:"状态码"`
	Message string           `json:"message" dc:"消息"`
	Data    *ProductDocsData `json:"data,omitempty" dc:"文档数据"`
}

// ProductDocsData 产品文档数据
type ProductDocsData struct {
	DocumentContent   string             `json:"documentContent" dc:"文档内容"`
	RequiredCerts     []string           `json:"requiredCerts" dc:"所需认证"`
	ComplianceItems   []ComplianceItem   `json:"complianceItems" dc:"合规项目"`
	EstimatedTime     string             `json:"estimatedTime" dc:"预计时间"`
	EstimatedCost     string             `json:"estimatedCost" dc:"预计费用"`
	RecommendedSteps  []string           `json:"recommendedSteps" dc:"推荐步骤"`
	RegulationDetails *RegulationDetails `json:"regulationDetails" dc:"法规详情"`
}

// ComplianceItem 合规项目
type ComplianceItem struct {
	Name        string `json:"name" dc:"项目名称"`
	Description string `json:"description" dc:"项目描述"`
	Required    bool   `json:"required" dc:"是否必需"`
	Status      string `json:"status" dc:"状态"`
}

// RegulationDetails 法规详情
type RegulationDetails struct {
	CountryName         string   `json:"countryName" dc:"国家名称"`
	MainRegulations     []string `json:"mainRegulations" dc:"主要法规"`
	CustomsRequirements string   `json:"customsRequirements" dc:"海关要求"`
	ImportRestrictions  string   `json:"importRestrictions" dc:"进口限制"`
}
