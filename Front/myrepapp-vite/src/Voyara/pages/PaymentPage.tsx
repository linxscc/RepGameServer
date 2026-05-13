import { useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { paymentApi } from '../api/payment';

export default function PaymentPage() {
  const { t } = useLanguage();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const orderId = Number(params.get('orderId'));
  const [method, setMethod] = useState<'stripe' | 'paypal'>('stripe');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!orderId) {
    return (
      <div className="vy-section">
        <div className="vy-container" style={{ textAlign: 'center', padding: '4rem 0' }}>
          <p>{t('payment.failed')}</p>
        </div>
      </div>
    );
  }

  const handlePayPal = async () => {
    setLoading(true);
    setError('');
    try {
      const returnUrl = `${window.location.origin}/voyara/orders`;
      const cancelUrl = window.location.href;
      const result = await paymentApi.createPayment(orderId, 'paypal', returnUrl, cancelUrl);
      if (result.paypalApprovalUrl) {
        window.location.href = result.paypalApprovalUrl;
      } else {
        setError('Failed to create PayPal payment');
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Payment failed');
    } finally {
      setLoading(false);
    }
  };

  const handleStripe = async () => {
    setLoading(true);
    setError('');
    try {
      const result = await paymentApi.createPayment(orderId, 'stripe');
      alert(`Stripe PaymentIntent created: ${result.gatewayOrderId}\nStatus: ${result.status}\n\nIn production, Stripe Elements would render here.`);
      navigate('/voyara/orders');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Payment failed');
    } finally {
      setLoading(false);
    }
  };

  const handlePay = async () => {
    if (method === 'paypal') {
      await handlePayPal();
    } else {
      await handleStripe();
    }
  };

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '500px' }}>
        <h1 className="vy-heading h2">{t('payment.title')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}

        <div style={{ marginTop: '2rem' }}>
          <h3 style={{ marginBottom: '1rem' }}>{t('checkout.payment')}</h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '1rem', border: `2px solid ${method === 'stripe' ? 'var(--vy-amber)' : 'var(--vy-border)'}`, borderRadius: '8px', cursor: 'pointer' }}>
              <input type="radio" name="method" checked={method === 'stripe'} onChange={() => setMethod('stripe')} />
              <span>{t('payment.stripe')}</span>
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', padding: '1rem', border: `2px solid ${method === 'paypal' ? 'var(--vy-amber)' : 'var(--vy-border)'}`, borderRadius: '8px', cursor: 'pointer' }}>
              <input type="radio" name="method" checked={method === 'paypal'} onChange={() => setMethod('paypal')} />
              <span>{t('payment.paypal')}</span>
            </label>
          </div>

          <button className="vy-btn vy-btn-primary vy-btn-lg" style={{ width: '100%', marginTop: '2rem' }} onClick={handlePay} disabled={loading}>
            {loading ? t('payment.paying') : t('payment.payNow')}
          </button>
        </div>
      </div>
    </div>
  );
}
