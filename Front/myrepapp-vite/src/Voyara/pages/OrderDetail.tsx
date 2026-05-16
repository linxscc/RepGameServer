import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { LoadingState, ErrorState } from '../components/LoadingState';
import { orderApi, type Order } from '../api/order';

const statusLabel: Record<string, string> = {
  pending: 'order.pending',
  paid: 'order.paid',
  shipped: 'order.shipped',
  delivered: 'order.delivered',
  cancelled: 'order.cancelled',
};

export default function OrderDetail() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) { setError('Order not found'); setLoading(false); return; }
    orderApi.getOrder(Number(id))
      .then(setOrder)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load order'))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <LoadingState />;
  if (error) return <ErrorState message={error} onBack={() => navigate('/voyara/orders')} />;
  if (!order) return null;

  const address = order.shippingAddress;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '800px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h1 className="vy-heading h2">{t('order.detailTitle') || `Order #${order.id}`}</h1>
          <Link to="/voyara/orders" className="vy-btn vy-btn-secondary vy-btn-sm">{t('order.backToList') || 'Back'}</Link>
        </div>

        {/* Status */}
        <div className="vy-card" style={{ padding: '1.25rem', marginBottom: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div style={{ color: 'var(--vy-text-dim)', fontSize: '0.85rem' }}>{order.orderNo || `#${order.id}`}</div>
              <div style={{ marginTop: '0.25rem' }}>
                <span style={{
                  padding: '0.25rem 0.75rem',
                  borderRadius: '12px',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                  background: order.paymentStatus === 'paid' ? '#d4edda' : order.paymentStatus === 'cancelled' ? '#f8d7da' : '#fff3cd',
                  color: order.paymentStatus === 'paid' ? '#155724' : order.paymentStatus === 'cancelled' ? '#721c24' : '#856404',
                }}>
                  {t(statusLabel[order.paymentStatus] || order.paymentStatus)}
                </span>
              </div>
            </div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--vy-amber)' }}>${order.amount.toFixed(2)}</div>
          </div>
          {order.paymentStatus === 'pending' && (
            <div style={{ marginTop: '1rem' }}>
              <a href={`/voyara/payment?orderId=${order.id}`} className="vy-btn vy-btn-primary">
                {t('payment.payNow')}
              </a>
            </div>
          )}
        </div>

        {/* Items */}
        <div className="vy-card" style={{ padding: '1.25rem', marginBottom: '1.5rem' }}>
          <h3 style={{ marginBottom: '1rem' }}>{t('order.items') || 'Items'}</h3>
          {order.items?.map((item) => (
            <div key={item.id} style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', padding: '0.5rem 0', borderBottom: '1px solid var(--vy-border)' }}>
              <div style={{ width: '60px', height: '60px', background: 'var(--vy-surface)', borderRadius: '4px', overflow: 'hidden', flexShrink: 0 }}>
                {item.imageUrl && <img src={item.imageUrl} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />}
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 500 }}>{item.title}</div>
                <div style={{ color: 'var(--vy-text-dim)', fontSize: '0.9rem' }}>{t('order.qty') || 'Qty'}: {item.quantity} x ${item.price.toFixed(2)}</div>
              </div>
              <div style={{ fontWeight: 600 }}>${item.total.toFixed(2)}</div>
            </div>
          ))}
          <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', borderTop: '2px solid var(--vy-border)', marginTop: '0.5rem', fontSize: '1.1rem', fontWeight: 700 }}>
            <span>{t('checkout.total')}</span>
            <span style={{ color: 'var(--vy-amber)' }}>${order.amount.toFixed(2)}</span>
          </div>
        </div>

        {/* Shipping Address */}
        <div className="vy-card" style={{ padding: '1.25rem', marginBottom: '1.5rem' }}>
          <h3 style={{ marginBottom: '1rem' }}>{t('checkout.shipping')}</h3>
          <table style={{ width: '100%', fontSize: '0.9rem' }}>
            <tbody>
              <tr><td style={{ color: 'var(--vy-text-dim)', width: '100px', padding: '0.25rem 0' }}>{t('checkout.name') || 'Name'}</td><td>{address.name}</td></tr>
              <tr><td style={{ color: 'var(--vy-text-dim)', padding: '0.25rem 0' }}>{t('checkout.phone') || 'Phone'}</td><td>{address.phone}</td></tr>
              <tr><td style={{ color: 'var(--vy-text-dim)', padding: '0.25rem 0' }}>{t('checkout.address') || 'Address'}</td><td>{address.street}, {address.city}, {address.country} {address.zipCode}</td></tr>
            </tbody>
          </table>
        </div>

        {/* Tracking Info */}
        {order.shippingStatus === 'shipped' && order.trackingNumber && (
          <div className="vy-card" style={{ padding: '1.25rem' }}>
            <h3 style={{ marginBottom: '0.5rem' }}>{t('order.tracking') || 'Tracking'}</h3>
            <p style={{ fontSize: '0.9rem', color: 'var(--vy-text-dim)' }}>{order.trackingNumber}</p>
          </div>
        )}
      </div>
    </div>
  );
}
