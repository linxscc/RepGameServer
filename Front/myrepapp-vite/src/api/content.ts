// ── 从后端获取页面内容的 API 封装 ──
// 复用 productDocs.ts 中的 API_BASE_URL

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

// ═══ 类型定义 ═══

export interface DownloadItemResponse {
  id: number;
  name: string;
  version: string;
  size: string;
  description: string;
  downloadUrl: string;
  icon: string;
  osType: string;
}

export interface SystemRequirementResponse {
  id: number;
  osType: string;
  osLabel: string;
  requirements: string[];
}

export interface ProfileInfoResponse {
  fullName: string;
  title: string;
  tagline: string;
  aboutText: string;
  email: string;
  phone: string;
  languages: string;
}

export interface WorkExperienceDataResponse {
  hero: Record<string, string>;
  features24: Record<string, string>;
  features25: Record<string, string>;
  steps2: Record<string, string>;
  contact10: Record<string, string>;
}

// ═══ 通用响应 wrapper ═══

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const json: ApiResponse<T> = await res.json();
  if (json.code !== 0) throw new Error(json.message || 'Unknown error');
  return json.data;
}

// ═══ API 方法 ═══

export function getDownloadItems(): Promise<DownloadItemResponse[]> {
  return fetchJSON<DownloadItemResponse[]>('/download-items');
}

export function getSystemRequirements(): Promise<SystemRequirementResponse[]> {
  return fetchJSON<SystemRequirementResponse[]>('/system-requirements');
}

export function getProfileInfo(): Promise<ProfileInfoResponse> {
  return fetchJSON<ProfileInfoResponse>('/profile-info');
}

export function getWorkExperience(): Promise<WorkExperienceDataResponse> {
  return fetchJSON<WorkExperienceDataResponse>('/work-experience');
}
