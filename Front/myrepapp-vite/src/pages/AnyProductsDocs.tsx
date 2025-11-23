import React, { useState } from 'react';
import { generateProductDocs, ProductDocsData } from '@/api/productDocs';
import { translations, Language } from '@/i18n/translations';
import './AnyProductsDocs.css';

interface FormData {
  companyName: string;
  productName: string;
  productCategory: string;
  targetCountry: string;
  productDescription: string;
  currentStep: number;
}

const AnyProductsDocs: React.FC = () => {
  const [language, setLanguage] = useState<Language>('zh');
  const t = translations[language];
  
  const [formData, setFormData] = useState<FormData>({
    companyName: '',
    productName: '',
    productCategory: '',
    targetCountry: '',
    productDescription: '',
    currentStep: 1,
  });

  const [generatedDocs, setGeneratedDocs] = useState<ProductDocsData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const totalSteps = 5;

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData({ ...formData, [field]: value });
  };

  const nextStep = () => {
    if (formData.currentStep < totalSteps) {
      setFormData({ ...formData, currentStep: formData.currentStep + 1 });
    }
  };

  const prevStep = () => {
    if (formData.currentStep > 1) {
      setFormData({ ...formData, currentStep: formData.currentStep - 1 });
    }
  };

  const generateDocuments = async () => {
    setIsLoading(true);
    setError('');
    
    console.log('准备发送的数据:', {
      companyName: formData.companyName,
      productName: formData.productName,
      productCategory: formData.productCategory,
      targetCountry: formData.targetCountry,
      productDescription: formData.productDescription,
      language: language, // 添加语言参数
    });
    
    try {
      const response = await generateProductDocs({
        companyName: formData.companyName,
        productName: formData.productName,
        productCategory: formData.productCategory,
        targetCountry: formData.targetCountry,
        productDescription: formData.productDescription,
        language: language, // 传递当前语言
      });

      console.log('收到响应:', response);

      if (response.code === 200 && response.data) {
        setGeneratedDocs(response.data);
      } else {
        setError(response.message || '生成文档失败');
      }
    } catch (err) {
      setError(t.networkError);
      console.error('生成文档错误:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const renderStep = () => {
    const productCategories = [
      { key: 'electronics', label: t.electronics },
      { key: 'foodBeverage', label: t.foodBeverage },
      { key: 'cosmetics', label: t.cosmetics },
      { key: 'textiles', label: t.textiles },
      { key: 'machinery', label: t.machinery },
      { key: 'other', label: t.other },
    ];

    const countries = [
      { key: 'usa', label: t.usa },
      { key: 'eu', label: t.eu },
      { key: 'japan', label: t.japan },
      { key: 'korea', label: t.korea },
      { key: 'australia', label: t.australia },
      { key: 'southeastAsia', label: t.southeastAsia },
      { key: 'middleEast', label: t.middleEast },
      { key: 'other', label: t.other },
    ];

    switch (formData.currentStep) {
      case 1:
        return (
          <div className="step-content">
            <h2 className="step-title">{t.step1Title}</h2>
            <input
              type="text"
              className="form-input"
              placeholder={t.step1Placeholder}
              value={formData.companyName}
              onChange={(e) => handleInputChange('companyName', e.target.value)}
              autoFocus
            />
          </div>
        );
      case 2:
        return (
          <div className="step-content">
            <h2 className="step-title">{t.step2Title}</h2>
            <input
              type="text"
              className="form-input"
              placeholder={t.step2Placeholder}
              value={formData.productName}
              onChange={(e) => handleInputChange('productName', e.target.value)}
              autoFocus
            />
          </div>
        );
      case 3:
        return (
          <div className="step-content">
            <h2 className="step-title">{t.step3Title}</h2>
            <div className="options-grid">
              {productCategories.map((category) => (
                <button
                  key={category.key}
                  className={`option-button ${formData.productCategory === category.label ? 'selected' : ''}`}
                  onClick={() => {
                    setFormData({ 
                      ...formData, 
                      productCategory: category.label,
                      currentStep: formData.currentStep + 1 
                    });
                  }}
                >
                  {category.label}
                </button>
              ))}
            </div>
          </div>
        );
      case 4:
        return (
          <div className="step-content">
            <h2 className="step-title">{t.step4Title}</h2>
            <div className="options-grid">
              {countries.map((country) => (
                <button
                  key={country.key}
                  className={`option-button ${formData.targetCountry === country.label ? 'selected' : ''}`}
                  onClick={() => {
                    setFormData({ 
                      ...formData, 
                      targetCountry: country.label,
                      currentStep: formData.currentStep + 1 
                    });
                  }}
                >
                  {country.label}
                </button>
              ))}
            </div>
          </div>
        );
      case 5:
        return (
          <div className="step-content">
            <h2 className="step-title">{t.step5Title}</h2>
            <textarea
              className="form-textarea"
              placeholder={t.step5Placeholder}
              value={formData.productDescription}
              onChange={(e) => handleInputChange('productDescription', e.target.value)}
              rows={6}
              autoFocus
            />
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="anyproductsdocs-container">
      <header className="header">
        <div className="logo">{t.logo}</div>
        <div className="header-actions">
          <button 
            className="btn-language" 
            onClick={() => setLanguage(language === 'zh' ? 'en' : 'zh')}
          >
            {language === 'zh' ? 'EN' : '中文'}
          </button>
          <button className="btn-secondary">{t.login}</button>
          <button className="btn-primary">{t.signup}</button>
        </div>
      </header>

      <div className="main-content">
        <div className="hero-section">
          <div className="badge">{t.badge}</div>
          <h1 className="hero-title">{t.heroTitle}</h1>
          <p className="hero-subtitle">
            {t.heroSubtitle}
          </p>
        </div>

        {!generatedDocs ? (
          <div className="form-container">
            {error && (
              <div className="error-message">
                <p>⚠️ {error}</p>
              </div>
            )}
            
            <div className="progress-bar">
              <div 
                className="progress-fill" 
                style={{ width: `${(formData.currentStep / totalSteps) * 100}%` }}
              />
            </div>
            
            <div className="step-indicator">
              {t.step} {formData.currentStep} / {totalSteps}
            </div>

            {renderStep()}

            <div className="navigation-buttons">
              {formData.currentStep > 1 && (
                <button className="btn-nav btn-prev" onClick={prevStep}>
                  {t.prevStep}
                </button>
              )}
              
              {formData.currentStep < totalSteps ? (
                <button 
                  className="btn-nav btn-next" 
                  onClick={nextStep}
                  disabled={
                    (formData.currentStep === 1 && !formData.companyName) ||
                    (formData.currentStep === 2 && !formData.productName) ||
                    (formData.currentStep === 3 && !formData.productCategory) ||
                    (formData.currentStep === 4 && !formData.targetCountry)
                  }
                >
                  {t.nextStep}
                </button>
              ) : (
                <button 
                  className="btn-nav btn-generate" 
                  onClick={generateDocuments}
                  disabled={!formData.productDescription || isLoading}
                >
                  {isLoading ? t.generating : t.generateDoc}
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className="results-container">
            <h2 className="results-title">{t.resultTitle}</h2>
            
            {/* 文档内容 */}
            <div className="document-preview">
              <pre>{generatedDocs.documentContent}</pre>
            </div>

            {/* 详细信息标签页 */}
            <div className="details-section">
              <h3>{t.requiredCerts}</h3>
              <ul className="cert-list">
                {generatedDocs.requiredCerts.map((cert, index) => (
                  <li key={index}>{cert}</li>
                ))}
              </ul>

              <h3>{t.complianceChecklist}</h3>
              <div className="compliance-grid">
                {generatedDocs.complianceItems.map((item, index) => (
                  <div key={index} className="compliance-card">
                    <div className="compliance-header">
                      <span className="compliance-name">{item.name}</span>
                      {item.required && <span className="required-badge">{t.required}</span>}
                    </div>
                    <p className="compliance-desc">{item.description}</p>
                    <span className={`status-badge ${item.status}`}>{item.status}</span>
                  </div>
                ))}
              </div>

              <h3>{t.estimateInfo}</h3>
              <div className="estimate-info">
                <div className="estimate-item">
                  <span className="label">{t.estimatedTime}</span>
                  <span className="value">{generatedDocs.estimatedTime}</span>
                </div>
                <div className="estimate-item">
                  <span className="label">{t.estimatedCost}</span>
                  <span className="value">{generatedDocs.estimatedCost}</span>
                </div>
              </div>

              <h3>{t.regulationDetails}</h3>
              <div className="regulation-details">
                <p><strong>{t.country}</strong>{generatedDocs.regulationDetails.countryName}</p>
                <p><strong>{t.mainRegulations}</strong></p>
                <ul>
                  {generatedDocs.regulationDetails.mainRegulations.map((reg, index) => (
                    <li key={index}>{reg}</li>
                  ))}
                </ul>
                <p><strong>{t.customsRequirements}</strong>{generatedDocs.regulationDetails.customsRequirements}</p>
                <p><strong>{t.importRestrictions}</strong>{generatedDocs.regulationDetails.importRestrictions}</p>
              </div>

              <h3>{t.recommendedSteps}</h3>
              <ol className="steps-list">
                {generatedDocs.recommendedSteps.map((step, index) => (
                  <li key={index}>{step}</li>
                ))}
              </ol>
            </div>

            <div className="results-actions">
              <button 
                className="btn-primary" 
                onClick={() => navigator.clipboard.writeText(generatedDocs.documentContent)}
              >
                {t.copyDoc}
              </button>
              <button 
                className="btn-primary" 
                onClick={() => {
                  const blob = new Blob([generatedDocs.documentContent], { type: 'text/plain' });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = `${formData.productName}-出口文件.txt`;
                  a.click();
                }}
              >
                {t.downloadDoc}
              </button>
              <button 
                className="btn-secondary" 
                onClick={() => {
                  setGeneratedDocs(null);
                  setFormData({ ...formData, currentStep: 1 });
                }}
              >
                {t.regenerate}
              </button>
            </div>
          </div>
        )}
      </div>

      <footer className="footer">
        <p>{t.footer}</p>
      </footer>
    </div>
  );
};

export default AnyProductsDocs;
