package controller

import (
	v1 "GoServer/api/v1"
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ProductDocs struct{}

// GenerateDocs 生成产品出口文档
func (p *ProductDocs) GenerateDocs(ctx context.Context, req *v1.GenerateProductDocsReq) (res *v1.GenerateProductDocsRes, err error) {
	// 默认语言为中文
	if req.Language == "" {
		req.Language = "zh"
	}

	g.Log().Infof(ctx, "收到产品文档生成请求 - 公司: %s, 产品: %s, 目标国家: %s, 语言: %s",
		req.CompanyName, req.ProductName, req.TargetCountry, req.Language)

	// 根据不同国家和语言生成相应的文档内容
	docsData := generateDocumentsByCountry(req)

	message := "文档生成成功"
	if req.Language == "en" {
		message = "Document generated successfully"
	}

	result := &v1.GenerateProductDocsRes{
		Code:    200,
		Message: message,
		Data:    docsData,
	}

	g.Log().Infof(ctx, "生成的文档数据: %+v", result)

	// 获取 HTTP 请求对象并直接写入响应
	r := g.RequestFromCtx(ctx)
	if r != nil {
		r.Response.WriteJson(result)
		return nil, nil
	}

	return result, nil
}

// generateDocumentsByCountry 根据国家生成文档
func generateDocumentsByCountry(req *v1.GenerateProductDocsReq) *v1.ProductDocsData {
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	var regulationDetails *v1.RegulationDetails
	var requiredCerts []string
	var complianceItems []v1.ComplianceItem
	var recommendedSteps []string
	var estimatedTime, estimatedCost string

	// 根据目标国家设置不同的要求
	switch req.TargetCountry {
	case "美国", "United States":
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "United States",
				MainRegulations:     []string{"FDA Certification", "FCC Certification", "UL Certification", "CPSC Compliance"},
				CustomsRequirements: "Commercial invoice, packing list, bill of lading, certificate of origin required",
				ImportRestrictions:  "Comply with Federal Trade Commission (FTC) and Customs and Border Protection (CBP) regulations",
			}
			requiredCerts = []string{"FDA Certification", "FCC ID", "UL Certification", "Certificate of Origin", "Commercial Invoice"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "美国",
				MainRegulations:     []string{"FDA认证", "FCC认证", "UL认证", "CPSC合规"},
				CustomsRequirements: "需提供商业发票、装箱单、提单、原产地证明",
				ImportRestrictions:  "符合美国联邦贸易委员会(FTC)和海关边境保护局(CBP)的规定",
			}
			requiredCerts = []string{"FDA认证", "FCC ID", "UL认证", "原产地证明", "商业发票"}
		}
		estimatedTime = "4-8周"
		estimatedCost = "$3,000 - $8,000"

	case "欧盟", "European Union":
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "European Union",
				MainRegulations:     []string{"CE Certification", "REACH Regulation", "RoHS Directive", "WEEE Directive"},
				CustomsRequirements: "EUR.1 certificate, commercial invoice, packing list required",
				ImportRestrictions:  "Comply with EU product safety directives and environmental requirements",
			}
			requiredCerts = []string{"CE Certification", "REACH Compliance Certificate", "RoHS Certificate", "EUR.1 Certificate of Origin", "Declaration of Conformity"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "欧盟",
				MainRegulations:     []string{"CE认证", "REACH法规", "RoHS指令", "WEEE指令"},
				CustomsRequirements: "需提供EUR.1证书、商业发票、装箱单",
				ImportRestrictions:  "符合欧盟产品安全指令和环保要求",
			}
			requiredCerts = []string{"CE认证", "REACH合规证明", "RoHS证书", "EUR.1原产地证书", "符合性声明"}
		}
		estimatedTime = "6-10周"
		estimatedCost = "€4,000 - €10,000"

	case "日本", "Japan":
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "Japan",
				MainRegulations:     []string{"PSE Certification", "JIS Standards", "Ministry of Health Approval"},
				CustomsRequirements: "Import license, commercial invoice, certificate of origin required",
				ImportRestrictions:  "Comply with Japanese Industrial Standards (JIS) and Food Sanitation Law",
			}
			requiredCerts = []string{"PSE Certification", "JIS Certification", "Certificate of Origin", "Import License", "Inspection Certificate"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "日本",
				MainRegulations:     []string{"PSE认证", "JIS标准", "厚生劳动省批准"},
				CustomsRequirements: "需提供进口许可证、商业发票、原产地证明",
				ImportRestrictions:  "符合日本工业标准(JIS)和食品卫生法",
			}
			requiredCerts = []string{"PSE认证", "JIS认证", "原产地证明", "进口许可证", "检验检疫证明"}
		}
		estimatedTime = "5-9周"
		estimatedCost = "¥500,000 - ¥1,200,000"

	case "韩国", "South Korea":
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "South Korea",
				MainRegulations:     []string{"KC Certification", "KS Standards"},
				CustomsRequirements: "Commercial invoice, packing list, certificate of origin required",
				ImportRestrictions:  "Comply with Korean Standards (KS) and Electrical Appliances Safety Control Act",
			}
			requiredCerts = []string{"KC Certification", "KS Certification", "Certificate of Origin", "Commercial Invoice"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "韩国",
				MainRegulations:     []string{"KC认证", "KS标准"},
				CustomsRequirements: "需提供商业发票、装箱单、原产地证明",
				ImportRestrictions:  "符合韩国国家标准(KS)和电气用品安全管理法",
			}
			requiredCerts = []string{"KC认证", "KS认证", "原产地证明", "商业发票"}
		}
		estimatedTime = "4-7周"
		estimatedCost = "₩4,000,000 - ₩9,000,000"

	case "澳大利亚", "Australia":
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "Australia",
				MainRegulations:     []string{"RCM Certification", "Australian Standards (AS/NZS)", "TGA Certification"},
				CustomsRequirements: "Commercial invoice, packing list, certificate of origin, import permit required",
				ImportRestrictions:  "Comply with Australian Competition and Consumer Commission (ACCC) requirements",
			}
			requiredCerts = []string{"RCM Certification", "AS/NZS Compliance Certificate", "Certificate of Origin", "Commercial Invoice"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         "澳大利亚",
				MainRegulations:     []string{"RCM认证", "澳洲标准(AS/NZS)", "TGA认证"},
				CustomsRequirements: "需提供商业发票、装箱单、原产地证明、进口许可证",
				ImportRestrictions:  "符合澳大利亚竞争和消费者法案(ACCC)要求",
			}
			requiredCerts = []string{"RCM认证", "AS/NZS合规证明", "原产地证明", "商业发票"}
		}
		estimatedTime = "5-8周"
		estimatedCost = "A$3,500 - A$9,000"

	default:
		if req.Language == "en" {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         req.TargetCountry,
				MainRegulations:     []string{"ISO 9001 Quality Management", "Product Safety Certification", "Customs Compliance"},
				CustomsRequirements: "Standard trade documents required",
				ImportRestrictions:  "Comply with target country import regulations",
			}
			requiredCerts = []string{"Product Quality Certificate", "Certificate of Origin", "Commercial Invoice", "Packing List"}
		} else {
			regulationDetails = &v1.RegulationDetails{
				CountryName:         req.TargetCountry,
				MainRegulations:     []string{"ISO 9001质量管理", "产品安全认证", "海关合规"},
				CustomsRequirements: "需提供标准贸易文件",
				ImportRestrictions:  "遵守目标国家的进口法规",
			}
			requiredCerts = []string{"产品质量证书", "原产地证明", "商业发票", "装箱单"}
		}
		estimatedTime = "3-6周"
		estimatedCost = "$2,000 - $6,000"
	}

	// 生成合规项目清单
	if req.Language == "en" {
		complianceItems = []v1.ComplianceItem{
			{
				Name:        "Product Safety Certification",
				Description: fmt.Sprintf("Obtain required product safety certification for %s", req.TargetCountry),
				Required:    true,
				Status:      "Pending",
			},
			{
				Name:        "Quality Test Report",
				Description: "Product quality test report issued by third-party laboratory",
				Required:    true,
				Status:      "Pending",
			},
			{
				Name:        "Certificate of Origin",
				Description: "Official document certifying product origin",
				Required:    true,
				Status:      "Pending",
			},
			{
				Name:        "Customs Declaration Documents",
				Description: "Complete customs import and export declaration materials",
				Required:    true,
				Status:      "Pending",
			},
			{
				Name:        "Product Labeling Compliance",
				Description: fmt.Sprintf("Ensure product labels comply with %s requirements", req.TargetCountry),
				Required:    true,
				Status:      "Pending",
			},
		}
	} else {
		complianceItems = []v1.ComplianceItem{
			{
				Name:        "产品安全认证",
				Description: fmt.Sprintf("获得%s所需的产品安全认证", req.TargetCountry),
				Required:    true,
				Status:      "待办",
			},
			{
				Name:        "质量检测报告",
				Description: "第三方实验室出具的产品质量检测报告",
				Required:    true,
				Status:      "待办",
			},
			{
				Name:        "原产地证明",
				Description: "证明产品原产地的官方文件",
				Required:    true,
				Status:      "待办",
			},
			{
				Name:        "海关申报文件",
				Description: "完整的海关进出口申报材料",
				Required:    true,
				Status:      "待办",
			},
			{
				Name:        "产品标签合规",
				Description: fmt.Sprintf("确保产品标签符合%s的要求", req.TargetCountry),
				Required:    true,
				Status:      "待办",
			},
		}
	}

	// 推荐步骤
	if req.Language == "en" {
		recommendedSteps = []string{
			"1. Conduct product classification and HS code confirmation",
			"2. Determine required certification types and standards",
			"3. Select certification body and submit application",
			"4. Prepare product samples for testing",
			"5. Obtain test reports and certification certificates",
			"6. Prepare customs declaration documents",
			"7. Apply for certificate of origin",
			"8. Ensure product labeling and packaging compliance",
			"9. Arrange shipping and customs clearance",
			"10. Submit all documents to complete import",
		}
	} else {
		recommendedSteps = []string{
			"1. 进行产品分类和HS编码确认",
			"2. 确定所需的认证类型和标准",
			"3. 选择认证机构并提交申请",
			"4. 准备产品样品进行测试",
			"5. 获取测试报告和认证证书",
			"6. 准备海关申报文件",
			"7. 办理原产地证明",
			"8. 确认产品标签和包装合规",
			"9. 安排货运和清关",
			"10. 提交所有文件完成进口",
		}
	}

	// 生成文档内容
	documentContent := generateDocumentContent(req, currentTime, regulationDetails, requiredCerts, estimatedTime, estimatedCost)

	return &v1.ProductDocsData{
		DocumentContent:   documentContent,
		RequiredCerts:     requiredCerts,
		ComplianceItems:   complianceItems,
		EstimatedTime:     estimatedTime,
		EstimatedCost:     estimatedCost,
		RecommendedSteps:  recommendedSteps,
		RegulationDetails: regulationDetails,
	}
}

// formatStringSlice 格式化字符串切片
func formatStringSlice(items []string) string {
	result := ""
	for i, item := range items {
		result += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	return result
}

// generateDocumentContent 根据语言生成文档内容
func generateDocumentContent(req *v1.GenerateProductDocsReq, currentTime string, regulationDetails *v1.RegulationDetails, requiredCerts []string, estimatedTime, estimatedCost string) string {
	if req.Language == "en" {
		// 英文版本
		documentContent := fmt.Sprintf(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
        Product Export Compliance Document - %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【Basic Information】
Company Name: %s
Product Name: %s
Product Category: %s
Target Market: %s
Generated Time: %s

【Product Description】
%s

【Regulatory Requirements】
Main Applicable Regulations:
%s

【Required Certifications】
`, req.TargetCountry, req.CompanyName, req.ProductName,
			req.ProductCategory, req.TargetCountry, currentTime,
			req.ProductDescription,
			formatStringSlice(regulationDetails.MainRegulations))

		for i, cert := range requiredCerts {
			documentContent += fmt.Sprintf("%d. %s\n", i+1, cert)
		}

		documentContent += fmt.Sprintf(`
【Customs Requirements】
%s

【Import Restrictions】
%s

【Estimated Time & Cost】
Estimated Certification Time: %s
Estimated Total Cost: %s

【Compliance Statement】
This product complies with the relevant regulations and standards of %s.
Exporters should ensure that the product meets all applicable safety, quality, and environmental standards.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Important Notes:
1. This document is a reference template; please refer to the latest regulations of the target country for specific requirements
2. It is recommended to consult with professional trade compliance consultants
3. Certification processes and costs may vary by product type
4. Please ensure the authenticity and accuracy of all documents
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Generated By: Export Doc Generator
Technical Support: AI-Powered Document Generation Platform
`, regulationDetails.CustomsRequirements,
			regulationDetails.ImportRestrictions,
			estimatedTime, estimatedCost, req.TargetCountry)

		return documentContent
	}

	// 中文版本（默认）
	documentContent := fmt.Sprintf(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
        产品出口合规文件 - %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【基本信息】
公司名称：%s
产品名称：%s
产品类别：%s
目标市场：%s
生成时间：%s

【产品描述】
%s

【法规要求】
主要适用法规：
%s

【所需认证文件】
`, req.TargetCountry, req.CompanyName, req.ProductName,
		req.ProductCategory, req.TargetCountry, currentTime,
		req.ProductDescription,
		formatStringSlice(regulationDetails.MainRegulations))

	for i, cert := range requiredCerts {
		documentContent += fmt.Sprintf("%d. %s\n", i+1, cert)
	}

	documentContent += fmt.Sprintf(`
【海关要求】
%s

【进口限制】
%s

【预计时间与费用】
预计认证时间：%s
预计总费用：%s

【合规性声明】
本产品符合%s的相关法规和标准要求。
出口企业应确保产品符合所有适用的安全、质量和环保标准。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
注意事项：
1. 本文件为参考模板，具体要求请以目标国家最新法规为准
2. 建议咨询专业的贸易合规顾问
3. 认证流程和费用可能因产品类型而异
4. 请确保所有文件的真实性和准确性
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

生成机构：出口文件生成器
技术支持：AI智能文档生成平台
`, regulationDetails.CustomsRequirements,
		regulationDetails.ImportRestrictions,
		estimatedTime, estimatedCost, req.TargetCountry)

	return documentContent
}
