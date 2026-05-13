import { useState, useEffect } from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { orderApi, type Order } from '../api/order';

const statusLabel: Record<string, string> = {
  pending: 'order.pending',
  paid: 'order.paid',
  shipped: 'order.shipped',
  delivered: 'order.delivered',
  cancelled: 'order.cancelled',
};

export default function Orders() {
  const { t } = useLanguage();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    orderApi.getOrders()
      .then(setOrders)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load orders'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="vy-section"><div className="vy-container"><p>Loading...</p></div></div>;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '800px' }}>
        <h1 className="vy-heading h2">{t('order.title')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}
        {orders.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--vy-text-dim)' }}>
            {t('order.noOrders')}
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginTop: '2rem' }}>
            {orders.map((order) => (
              <div key={order.id} className="vy-card" style={{ padding: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                  <span style={{ color: 'var(--vy-text-dim)', fontSize: '0.85rem' }}>
                    {order.orderNo || `#${order.id}`}
                  </span>
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
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {order.items?.map((item) => (
                    <div key={item.id} style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                      <div style={{ width: '48px', height: '48px', background: 'var(--vy-surface)', borderRadius: '4px', overflow: 'hidden', flexShrink: 0 }}>
                        {item.imageUrl && <img src={item.imageUrl} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />}
                      </div>
                      <div style={{ flex: 1, fontSize: '0.9rem' }}>
                        <div style={{ fontWeight: 500 }}>{item.title}</div>
                        <div style={{ color: 'var(--vy-text-dim)' }}>x{item.quantity}</div>
                      </div>
                      <div style={{ fontWeight: 600 }}>${item.total.toFixed(2)}</div>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--vy-border)' }}>
                  <div>
                    {order.shippingStatus === 'shipped' && order.trackingNumber && (
                      <span style={{ fontSize: '0.85rem', color: 'var(--vy-text-dim)' }}>
                        {t('order.tracking')}: {order.trackingNumber}
                      </span>
                    )}
                  </div>
                  <div style={{ fontWeight: 700, color: 'var(--vy-amber)' }}>
                    ${order.amount.toFixed(2)}
                  </div>
                </div>
                {order.paymentStatus === 'pending' && (
                  <div style={{ marginTop: '0.75rem' }}>
                    <a href={`/voyara/payment?orderId=${order.id}`} className="vy-btn vy-btn-primary vy-btn-sm">
                      {t('payment.payNow')}
                    </a>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
