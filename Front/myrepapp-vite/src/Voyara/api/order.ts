import { voyaraApi } from './client';
import type { ShippingAddress } from './types';

export interface OrderItem {
  id: number;
  productId: number;
  title: string;
  price: number;
  quantity: number;
  total: number;
  imageUrl: string;
}

export interface Order {
  id: number;
  orderNo: string;
  buyerId: number;
  sellerId: number;
  itemCount: number;
  amount: number;
  subtotal: number;
  currency: string;
  paymentStatus: string;
  shippingStatus: string;
  trackingNumber: string;
  shippingAddress: ShippingAddress;
  createdAt: string;
  items?: OrderItem[];
}

export const orderApi = {
  checkout: (productIds: number[], shippingAddress: ShippingAddress, idempotencyKey?: string) =>
    voyaraApi.post<Order>('/orders', { productIds, shippingAddress, idempotencyKey }),
  getOrders: () => voyaraApi.get<Order[]>('/orders'),
  shipOrder: (id: number, trackingNumber: string) =>
    voyaraApi.put<Order>(`/orders/${id}/ship`, { trackingNumber }),
};
