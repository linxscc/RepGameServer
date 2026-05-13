import { voyaraApi } from './client';
import type { CartItem } from './types';

export interface CartResponse {
  items: CartItem[];
  count: number;
}

export const cartApi = {
  getCart: () => voyaraApi.get<CartResponse>('/cart'),
  addToCart: (productId: number, quantity: number = 1) =>
    voyaraApi.post<{ message: string }>('/cart', { productId, quantity }),
  updateQuantity: (id: number, quantity: number) =>
    voyaraApi.put<CartItem>(`/cart/${id}`, { quantity }),
  toggleSelect: (id: number, selected: boolean) =>
    voyaraApi.put<{ message: string }>(`/cart/${id}/select`, { selected }),
  selectAll: (selected: boolean) =>
    voyaraApi.put<{ message: string }>('/cart/select-all', { selected }),
  removeItem: (id: number) =>
    voyaraApi.delete<{ message: string }>(`/cart/${id}`),
  clearCart: () =>
    voyaraApi.delete<{ message: string }>('/cart'),
};
