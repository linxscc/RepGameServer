import { voyaraApi } from './client';

export interface CreatePaymentResult {
  paymentId: number;
  clientSecret?: string;
  paypalApprovalUrl?: string;
  gatewayOrderId: string;
  status: string;
}

export const paymentApi = {
  createPayment: (orderId: number, method: 'stripe' | 'paypal', returnUrl?: string, cancelUrl?: string) =>
    voyaraApi.post<CreatePaymentResult>('/payments', { orderId, method, returnUrl, cancelUrl }),
  capturePayPal: (paypalOrderId: string, orderId: number) =>
    voyaraApi.post<{ message: string }>('/payments/paypal/capture', { paypalOrderId, orderId }),
};
