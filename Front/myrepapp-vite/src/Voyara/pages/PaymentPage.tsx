import { useState, useEffect, useRef } from 'react';
import { useSearchParams, useNavigate, useLocation } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { LoadingState, ErrorState } from '../components/LoadingState';
import { paymentApi } from '../api/payment';
import { orderApi } from '../api/order';

export default function PaymentPage() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const location = useLocation();
  const [params] = useSearchParams();
  const stateOrderId = (location.state as { orderId?: number } | null)?.orderId;
  const orderId = stateOrderId || Number(params.get('orderId'));
  const paypalToken = params.get('token');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [verifying, setVerifying] = useState(true);
  const captured = useRef(false);

  // Verify order belongs to current user (prevent payment for others' orders)
  useEffect(() => {
    if (!orderId || paypalToken) { setVerifying(false); return; }
    orderApi.getOrders().then((orders) => {
      const owned = orders.some((o) => o.id === orderId && o.paymentStatus === 'pending');
      if (!owned) setError('Order not found or already paid');
    }).catch(() => {}).finally(() => setVerifying(false));
  }, [orderId, paypalToken]);

  // Handle PayPal return: user approved on PayPal, now capture
  useEffect(() => {
    if (!orderId || !paypalToken || captured.current) return;
    captured.current = true;
    setLoading(true);
    paymentApi.capturePayPal(paypalToken, orderId)
      .then(() => navigate(`/voyara/order/${orderId}`))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Payment capture failed');
        setLoading(false);
      });
  }, [orderId, paypalToken, navigate]);

  if (!orderId) {
    return <ErrorState message={t('payment.failed')} onBack={() => navigate('/voyara/orders')} />;
  }

  if (verifying) return <LoadingState />;

  if (paypalToken) {
    return (
      <div className="vy-section">
        <div className="vy-container" style={{ textAlign: 'center', padding: '4rem 0' }}>
          {error && <div className="vy-auth-error">{error}</div>}
          {!error && <p style={{ color: 'var(--vy-text-dim)' }}>{t('payment.paying')}</p>}
        </div>
      </div>
    );
  }

  const handlePay = async () => {
    setLoading(true);
    setError('');
    try {
      const returnUrl = `${window.location.origin}/voyara/payment?orderId=${orderId}`;
      const cancelUrl = `${window.location.origin}/voyara/orders`;
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

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '500px' }}>
        <h1 className="vy-heading h2">{t('payment.title')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}

        <div style={{ marginTop: '2rem', textAlign: 'center' }}>
          <p style={{ marginBottom: '1.5rem', color: 'var(--vy-text-dim)' }}>
            {t('payment.paypal')}
          </p>
          <button className="vy-btn vy-btn-primary vy-btn-lg" style={{ width: '100%' }} onClick={handlePay} disabled={loading}>
            {loading ? t('payment.paying') : t('payment.payNow')}
          </button>
        </div>
      </div>
    </div>
  );
}
