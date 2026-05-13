// ── Voyara API client ──
import type {
  AuthResponse, LoginPayload, RegisterPayload,
  Product, Category, Order, ShippingAddress,
} from './types';

const API_BASE = import.meta.env.VITE_VOYARA_API ?? '/voyara';

function getCSRFToken(): string {
  let token = sessionStorage.getItem('voyara_csrf');
  if (!token) {
    token = Array.from({ length: 64 }, () =>
      Math.random().toString(36)[2]).join('');
    sessionStorage.setItem('voyara_csrf', token);
  }
  return token;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('voyara_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((token && { Authorization: `Bearer ${token}` }) as Record<string, string>),
    ...(options?.headers as Record<string, string>),
  };

  if (options?.method && options.method !== 'GET' && options.method !== 'HEAD') {
    headers['X-CSRF-Token'] = getCSRFToken();
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }));
    throw new Error(err.message ?? 'Request failed');
  }
  const json = await res.json();
  return json.data ?? json;
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
    return get<Product[]>(`/products${params}`);
  },
  getProduct: (id: number) => get<Product>(`/products/${id}`),
  createProduct: (data: FormData | Record<string, unknown>) =>
    post<Product>('/products', data),
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
};
