// 全局类型定义
export interface DownloadItem {
  name: string;
  version: string;
  size: string;
  description: string;
  downloadUrl: string;
  icon: string;
}

export interface NavigationProps {
  // Navigation组件的props类型
}

export interface HeroProps {
  heading1?: React.ReactElement;
  content1?: React.ReactElement;
  action1?: React.ReactElement;
  action2?: React.ReactElement;
}

export interface FeatureProps {
  feature1Title?: React.ReactElement;
  feature2Title?: React.ReactElement;
  feature3Title?: React.ReactElement;
  feature1Description?: React.ReactElement;
  feature2Description?: React.ReactElement;
  feature3Description?: React.ReactElement;
  feature1ImgSrc?: string;
  feature2ImgSrc?: string;
  feature3ImgSrc?: string;
  feature1ImgAlt?: string;
  feature2ImgAlt?: string;
  feature3ImgAlt?: string;
}

export interface StepProps {
  step1Title?: React.ReactElement;
  step2Title?: React.ReactElement;
  step3Title?: React.ReactElement;
  step4Title?: React.ReactElement;
  step1Description?: React.ReactElement;
  step2Description?: React.ReactElement;
  step3Description?: React.ReactElement;
  step4Description?: React.ReactElement;
}

export interface ContactProps {
  content1?: React.ReactElement;
  heading1?: React.ReactElement;
  email1?: React.ReactElement;
}
