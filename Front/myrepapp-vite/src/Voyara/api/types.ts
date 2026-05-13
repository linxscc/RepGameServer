// ── Voyara API 类型定义 ──

export interface UserInfo {
  id: number;
  email: string;
  name: string;
  role: string;
  emailVerified: boolean;
  preferredLang: string;
}

export interface VerificationResponse {
  message: string;
}

export interface User {
  id: number;
  email: string;
  name: string;
  phone: string;
  country: string;
  preferredLang: string;
}

export interface Seller {
  id: number;
  userId: number;
  shopName: string;
  description: string;
  verified: boolean;
  rating: number;
}

export type ProductCategory = 'appliance' | 'vehicle' | 'electronics' | 'other';
export type ProductCondition = 'new' | 'like_new' | 'used' | 'refurbished';
export type ProductStatus = 'active' | 'sold' | 'inactive';

export interface Product {
  id: number;
  sellerId: number;
  sellerName?: string;
  shopName?: string;
  title: string;
  description: string;
  price: number;
  currency: string;
  category: ProductCategory;
  condition: ProductCondition;
  images: string[];
  status: ProductStatus;
  createdAt: string;
  updatedAt: string;
}

export type PaymentStatus = 'pending' | 'paid' | 'refunded';
export type ShippingStatus = 'pending' | 'shipped' | 'delivered';

export interface ShippingAddress {
  name: string;
  phone: string;
  country: string;
  city: string;
  street: string;
  zipCode: string;
}

export interface Order {
  id: number;
  buyerId: number;
  productId: number;
  productTitle?: string;
  productImage?: string;
  amount: number;
  currency: string;
  paymentStatus: PaymentStatus;
  shippingStatus: ShippingStatus;
  trackingNumber: string;
  shippingAddress: ShippingAddress;
  createdAt: string;
}

export interface Category {
  id: number;
  name: string;
  parentId: number | null;
  icon: string;
}

export interface ProductFilter {
  category?: ProductCategory;
  minPrice?: number;
  maxPrice?: number;
  condition?: ProductCondition;
  search?: string;
  page?: number;
  pageSize?: number;
}

export interface RegisterPayload {
  email: string;
  password: string;
  name: string;
  code: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  refreshToken?: string;
  user: UserInfo;
}

export interface ProductListResponse {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
}
