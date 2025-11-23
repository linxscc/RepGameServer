// 中英文翻译配置
export const translations = {
  zh: {
    // 头部
    logo: '出口文件生成器',
    login: '登录',
    signup: '注册',
    
    // Hero 区域
    badge: 'AI 智能文档生成平台',
    heroTitle: '一键生成符合出口国标准的产品文件',
    heroSubtitle: '帮助企业快速生成符合目标国家法规要求的产品审核文件，简化出口流程，节省时间成本',
    
    // 表单步骤
    step: '步骤',
    step1Title: '1. 您的公司名称是什么？',
    step1Placeholder: '请输入公司名称...',
    step2Title: '2. 您要出口的产品名称？',
    step2Placeholder: '请输入产品名称...',
    step3Title: '3. 产品类别是什么？',
    step4Title: '4. 目标出口国家/地区？',
    step5Title: '5. 请简要描述您的产品',
    step5Placeholder: '请输入产品详细描述、特性、用途等...',
    
    // 产品类别
    electronics: '电子产品',
    foodBeverage: '食品饮料',
    cosmetics: '化妆品',
    textiles: '纺织品',
    machinery: '机械设备',
    other: '其他',
    
    // 国家/地区
    usa: '美国',
    eu: '欧盟',
    japan: '日本',
    korea: '韩国',
    australia: '澳大利亚',
    southeastAsia: '东南亚',
    middleEast: '中东',
    
    // 按钮
    prevStep: '← 上一步',
    nextStep: '下一步 →',
    generating: '生成中...',
    generateDoc: '生成文件 ✓',
    
    // 错误和消息
    networkError: '网络错误，请检查后端服务是否正常运行',
    
    // 结果页面
    resultTitle: '生成的出口文件',
    requiredCerts: '📋 所需认证文件',
    complianceChecklist: '✅ 合规项目清单',
    required: '必需',
    estimateInfo: '📅 预计时间与费用',
    estimatedTime: '预计时间：',
    estimatedCost: '预计费用：',
    regulationDetails: '🔍 法规详情',
    country: '国家/地区：',
    mainRegulations: '主要法规：',
    customsRequirements: '海关要求：',
    importRestrictions: '进口限制：',
    recommendedSteps: '📝 推荐步骤',
    
    // 操作按钮
    copyDoc: '复制文件',
    downloadDoc: '下载文件',
    regenerate: '重新生成',
    
    // 页脚
    footer: '© 2025 出口文件生成器. 帮助企业简化出口流程',
  },
  
  en: {
    // Header
    logo: 'Export Doc Generator',
    login: 'Login',
    signup: 'Sign Up',
    
    // Hero section
    badge: 'AI-Powered Document Generation Platform',
    heroTitle: 'Generate Export-Compliant Product Documents Instantly',
    heroSubtitle: 'Help businesses quickly generate product certification documents that meet target country regulatory requirements, simplifying export processes and saving time',
    
    // Form steps
    step: 'Step',
    step1Title: '1. What is your company name?',
    step1Placeholder: 'Enter company name...',
    step2Title: '2. What is your product name?',
    step2Placeholder: 'Enter product name...',
    step3Title: '3. What is your product category?',
    step4Title: '4. Target export country/region?',
    step5Title: '5. Please describe your product briefly',
    step5Placeholder: 'Enter detailed product description, features, uses, etc...',
    
    // Product categories
    electronics: 'Electronics',
    foodBeverage: 'Food & Beverage',
    cosmetics: 'Cosmetics',
    textiles: 'Textiles',
    machinery: 'Machinery',
    other: 'Other',
    
    // Countries/Regions
    usa: 'United States',
    eu: 'European Union',
    japan: 'Japan',
    korea: 'South Korea',
    australia: 'Australia',
    southeastAsia: 'Southeast Asia',
    middleEast: 'Middle East',
    
    // Buttons
    prevStep: '← Previous',
    nextStep: 'Next →',
    generating: 'Generating...',
    generateDoc: 'Generate Document ✓',
    
    // Errors and messages
    networkError: 'Network error, please check if backend service is running',
    
    // Results page
    resultTitle: 'Generated Export Document',
    requiredCerts: '📋 Required Certifications',
    complianceChecklist: '✅ Compliance Checklist',
    required: 'Required',
    estimateInfo: '📅 Estimated Time & Cost',
    estimatedTime: 'Estimated Time: ',
    estimatedCost: 'Estimated Cost: ',
    regulationDetails: '🔍 Regulation Details',
    country: 'Country/Region: ',
    mainRegulations: 'Main Regulations: ',
    customsRequirements: 'Customs Requirements: ',
    importRestrictions: 'Import Restrictions: ',
    recommendedSteps: '📝 Recommended Steps',
    
    // Action buttons
    copyDoc: 'Copy Document',
    downloadDoc: 'Download Document',
    regenerate: 'Regenerate',
    
    // Footer
    footer: '© 2025 Export Doc Generator. Simplifying export processes for businesses',
  },
};

export type Language = 'zh' | 'en';
export type TranslationKey = keyof typeof translations.zh;
