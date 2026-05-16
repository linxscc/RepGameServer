import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { EmptyState, ErrorState } from '../components/LoadingState';
import { cartApi } from '../api/cart';
import { orderApi } from '../api/order';
import type { CartItem, ShippingAddress } from '../api/types';

export default function Checkout() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const location = useLocation();
  const stateProducts = (location.state as { products?: number[] } | null)?.products;
  const params = new URLSearchParams(window.location.search);
  const productIds = stateProducts || params.get('products')?.split(',').map(Number) || [];
  const [items, setItems] = useState<CartItem[]>([]);
  const [loadError, setLoadError] = useState('');
  const [address, setAddress] = useState<ShippingAddress>({
    name: '', phone: '', country: '', city: '', street: '', zipCode: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const loadCart = useCallback(() => {
    if (productIds.length === 0) return;
    setLoading(true);
    setLoadError('');
    cartApi.getCart().then((data) => {
      setItems(data.items.filter((i) => productIds.includes(i.productId) && i.selected));
    }).catch((err) => setLoadError(err instanceof Error ? err.message : 'Failed to load cart'))
    .finally(() => setLoading(false));
  }, [productIds.join(',')]);

  useEffect(() => { loadCart(); }, [loadCart]);

  const update = (key: keyof ShippingAddress) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setAddress((a) => ({ ...a, [key]: e.target.value }));

  const subtotal = items.reduce((s, i) => s + i.productPrice * i.quantity, 0);
  const total = subtotal;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (productIds.length === 0) { setError('No items selected'); return; }
    setError('');
    setLoading(true);
    try {
      const key = `checkout_${crypto.randomUUID()}`;
      const order = await orderApi.checkout(productIds, address, key);
      setSuccess(true);
      setTimeout(() => navigate(`/voyara/payment?orderId=${order.id}`, { state: { orderId: order.id } }), 1500);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Checkout failed');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="vy-section">
        <div className="vy-container" style={{ textAlign: 'center', padding: '4rem 0' }}>
          <h2 style={{ color: 'var(--vy-amber)' }}>{t('checkout.success')}</h2>
          <p style={{ color: 'var(--vy-text-dim)', marginTop: '1rem' }}>{t('payment.redirecting')}</p>
        </div>
      </div>
    );
  }

  if (productIds.length === 0) {
    return <EmptyState message={t('cart.empty')} action={{ label: t('cart.continueShopping'), href: '/voyara' }} />;
  }

  if (loadError) return <ErrorState message={loadError} onRetry={loadCart} />;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '800px' }}>
        <h1 className="vy-heading h2">{t('checkout.title')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: '2rem', marginTop: '2rem' }}>
          <div>
            <h3 style={{ marginBottom: '1rem' }}>{t('checkout.shipping')}</h3>
            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="vy-form-group">
                  <label className="vy-label">Name *</label>
                  <input className="vy-input" value={address.name} onChange={update('name')} required />
                </div>
                <div className="vy-form-group">
                  <label className="vy-label">Phone *</label>
                  <input className="vy-input" value={address.phone} onChange={update('phone')} required />
                </div>
              </div>
              <div className="vy-form-group">
                <label className="vy-label">Country *</label>
                <input className="vy-input" value={address.country} onChange={update('country')} required />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="vy-form-group">
                  <label className="vy-label">City *</label>
                  <input className="vy-input" value={address.city} onChange={update('city')} required />
                </div>
                <div className="vy-form-group">
                  <label className="vy-label">Zip Code</label>
                  <input className="vy-input" value={address.zipCode} onChange={update('zipCode')} />
                </div>
              </div>
              <div className="vy-form-group">
                <label className="vy-label">Street Address *</label>
                <input className="vy-input" value={address.street} onChange={update('street')} required />
              </div>
              <button className="vy-btn vy-btn-primary vy-btn-lg" style={{ width: '100%', marginTop: '1.5rem' }} disabled={loading}>
                {loading ? '...' : t('checkout.placeOrder')}
              </button>
            </form>
          </div>

          <div>
            <h3 style={{ marginBottom: '1rem' }}>{t('checkout.orderSummary')}</h3>
            <div style={{ border: '1px solid var(--vy-border)', borderRadius: '8px', padding: '1rem' }}>
              {items.map((item) => (
                <div key={item.id} style={{ display: 'flex', gap: '0.75rem', padding: '0.5rem 0', borderBottom: '1px solid var(--vy-border)' }}>
                  <div style={{ width: '60px', height: '60px', background: 'var(--vy-surface)', borderRadius: '4px', overflow: 'hidden', flexShrink: 0 }}>
                    {item.productImage && <img src={item.productImage} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />}
                  </div>
                  <div style={{ flex: 1, fontSize: '0.9rem' }}>
                    <div style={{ fontWeight: 500 }}>{item.productTitle}</div>
                    <div style={{ color: 'var(--vy-text-dim)' }}>x{item.quantity}</div>
                  </div>
                  <div style={{ fontWeight: 600 }}>${(item.productPrice * item.quantity).toFixed(2)}</div>
                </div>
              ))}
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', fontWeight: 600 }}>
                <span>{t('checkout.subtotal')}</span>
                <span>${subtotal.toFixed(2)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 0', borderTop: '2px solid var(--vy-border)', fontSize: '1.1rem', fontWeight: 700 }}>
                <span>{t('checkout.total')}</span>
                <span style={{ color: 'var(--vy-amber)' }}>${total.toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
