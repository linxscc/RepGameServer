// API 基础配置
// 在开发环境使用 localhost，在生产环境使用空字符串（相对路径）
// 使用 ?? 运算符来正确处理空字符串
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8000';

// 产品文档生成请求接口
export interface GenerateProductDocsRequest {
  companyName: string;
  productName: string;
  productCategory: string;
  targetCountry: string;
  productDescription: string;
  language?: string; // 添加语言参数，可选，默认为 'zh'
}

// 合规项目接口
export interface ComplianceItem {
  name: string;
  description: string;
  required: boolean;
  status: string;
}

// 法规详情接口
export interface RegulationDetails {
  countryName: string;
  mainRegulations: string[];
  customsRequirements: string;
  importRestrictions: string;
}

// 产品文档数据接口
export interface ProductDocsData {
  documentContent: string;
  requiredCerts: string[];
  complianceItems: ComplianceItem[];
  estimatedTime: string;
  estimatedCost: string;
  recommendedSteps: string[];
  regulationDetails: RegulationDetails;
}

// 产品文档生成响应接口
export interface GenerateProductDocsResponse {
  code: number;
  message: string;
  data?: ProductDocsData;
}

// 生成产品出口文档
export const generateProductDocs = async (
  request: GenerateProductDocsRequest
): Promise<GenerateProductDocsResponse> => {
  try {
    console.log('发送请求:', request);
    console.log('请求 URL:', `${API_BASE_URL}/product/generate-docs`);
    
    const response = await fetch(`${API_BASE_URL}/product/generate-docs`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    console.log('响应状态:', response.status);
    console.log('响应头:', response.headers);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const text = await response.text();
    console.log('响应文本:', text);
    
    // 如果响应为空，返回错误
    if (!text || text.trim() === '' || text.trim() === '{}') {
      throw new Error('服务器返回空响应，请检查后端日志');
    }

    const data = JSON.parse(text);
    console.log('解析后的数据:', data);
    return data;
  } catch (error) {
    console.error('生成产品文档失败:', error);
    throw error;
  }
};

// 导出 API 基础 URL（用于其他可能的 API 调用）
export { API_BASE_URL };
