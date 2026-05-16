// ── Voyara API client ──
import type {
  AuthResponse, LoginPayload, RegisterPayload,
  Product, ProductListResponse, Category, Order, ShippingAddress,
} from './types';

const API_BASE = import.meta.env.VITE_VOYARA_API ?? '/voyara';

// 401 interceptor: wired up by AuthContext on mount
let refreshHandler: (() => Promise<string | null>) | null = null;
let logoutHandler: (() => void) | null = null;

export function setAuthHandlers(
  refresh: () => Promise<string | null>,
  logout: () => void,
) {
  refreshHandler = refresh;
  logoutHandler = logout;
}

let csrfPromise: Promise<string> | null = null;

async function fetchCSRFToken(): Promise<string> {
  const cached = sessionStorage.getItem('voyara_csrf');
  if (cached) return cached;

  if (!csrfPromise) {
    csrfPromise = (async () => {
      try {
        const res = await fetch(`${API_BASE}/auth/csrf-token`);
        if (!res.ok) throw new Error('CSRF fetch failed');
        const json = await res.json();
        const data = json.data ?? json;
        sessionStorage.setItem('voyara_csrf', data.token);
        return data.token;
      } catch {
        // Fallback: generate locally if server unavailable
        const fallback = Array.from({ length: 64 }, () =>
          Math.random().toString(36)[2]).join('');
        sessionStorage.setItem('voyara_csrf', fallback);
        return fallback;
      } finally {
        csrfPromise = null;
      }
    })();
  }
  return csrfPromise;
}

async function getCSRFToken(): Promise<string> {
  return fetchCSRFToken();
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('voyara_token');
  const headers: Record<string, string> = {
    ...((token && { Authorization: `Bearer ${token}` }) as Record<string, string>),
    ...(options?.headers as Record<string, string>),
  };

  // Only set Content-Type for non-FormData bodies; fetch sets multipart boundary automatically
  if (!(options?.body instanceof FormData) && headers['Content-Type'] === undefined) {
    headers['Content-Type'] = 'application/json';
  }

  if (options?.method && options.method !== 'GET' && options.method !== 'HEAD') {
    headers['X-CSRF-Token'] = await getCSRFToken();
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  // 401 interceptor: try token refresh once
  if (res.status === 401 && refreshHandler) {
    const newToken = await refreshHandler();
    if (newToken) {
      headers['Authorization'] = `Bearer ${newToken}`;
      const retryRes = await fetch(`${API_BASE}${path}`, { ...options, headers });
      if (retryRes.ok) {
        const retryJson = await retryRes.json();
        return retryJson.data ?? retryJson;
      }
    }
    logoutHandler?.();
    throw new Error('Session expired');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }));
    throw new Error(err.message ?? 'Request failed');
  }
  const json = await res.json();
  if (json.code !== undefined && json.code !== 0) {
    throw new Error(json.message || 'Request failed');
  }
  return json.data != null ? json.data : json;
}

function get<T>(path: string): Promise<T> {
  return request<T>(path);
}

function post<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) });
}

function put<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PUT', body: JSON.stringify(body) });
}

function del<T>(path: string): Promise<T> {
  return request<T>(path, { method: 'DELETE' });
}

export const voyaraApi = {
  // Generic methods
  get: <T>(path: string) => get<T>(path),
  post: <T>(path: string, body: unknown) => post<T>(path, body),
  put: <T>(path: string, body: unknown) => put<T>(path, body),
  delete: <T>(path: string) => del<T>(path),

  login: (data: LoginPayload) => post<AuthResponse>('/auth/login', data),
  register: (data: RegisterPayload) => post<AuthResponse>('/auth/register', data),

  getProducts: (filter?: Record<string, string>) => {
    const params = filter ? '?' + new URLSearchParams(filter).toString() : '';
    return get<ProductListResponse>(`/products${params}`);
  },
  getProduct: (id: number) => get<Product>(`/products/${id}`),
  createProduct: (data: FormData | Record<string, unknown>) => {
    if (data instanceof FormData) {
      return request<Product>('/products', { method: 'POST', body: data });
    }
    return post<Product>('/products', data);
  },
  updateProduct: (id: number, data: Record<string, unknown>) =>
    put<Product>(`/products/${id}`, data),

  getCategories: () => get<Category[]>('/categories'),

  createOrder: (data: { productId: number; shippingAddress: ShippingAddress }) =>
    post<Order>('/orders', data),
  getOrders: () => get<Order[]>('/orders'),
  shipOrder: (id: number, trackingNumber: string) =>
    put<Order>(`/orders/${id}/ship`, { trackingNumber }),

  sendVerificationCode: (email: string, purpose: string) =>
    post<{ message: string }>('/auth/send-verification', { email, purpose }),
  verifyEmail: (email: string, code: string) =>
    post<{ message: string }>('/auth/verify-email', { email, code }),
  refreshToken: (refreshToken: string) =>
    post<{ token: string; refreshToken: string }>('/auth/refresh', { refreshToken }),
  changePassword: (oldPassword: string, newPassword: string) =>
    post<{ message: string }>('/auth/change-password', { oldPassword, newPassword }),
  forgotPassword: (email: string) =>
    post<{ message: string }>('/auth/forgot-password', { email }),
  resetPassword: (email: string, code: string, newPassword: string) =>
    post<{ message: string }>('/auth/reset-password', { email, code, newPassword }),

  // ── Admin ──
  getAdminProducts: (status?: string) => {
    const params = status ? `?status=${status}` : '';
    return get<{ items: Product[] }>(`/admin/products${params}`);
  },
  updateProductStatus: (id: number, status: string) =>
    put<{ message: string }>(`/admin/products/${id}/status`, { status }),

  // ── Seller ──
  getSellerProducts: () => get<{ items: Product[] }>('/seller/products'),
};
