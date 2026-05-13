import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { cartApi, type CartResponse } from '../api/cart';
import type { CartItem } from '../api/types';

export default function CartPage() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [cart, setCart] = useState<CartResponse>({ items: [], count: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadCart = async () => {
    try {
      setLoading(true);
      const data = await cartApi.getCart();
      setCart(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load cart');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadCart(); }, []);

  const groupedItems = (): { sellerId: number; shopName: string; items: CartItem[] }[] => {
    const groups: Record<number, { sellerId: number; shopName: string; items: CartItem[] }> = {};
    for (const item of cart.items) {
      if (!groups[item.sellerId]) {
        groups[item.sellerId] = { sellerId: item.sellerId, shopName: item.sellerShopName, items: [] };
      }
      groups[item.sellerId].items.push(item);
    }
    return Object.values(groups);
  };

  const selectedItems = cart.items.filter((i) => i.selected && i.available);
  const totalAmount = selectedItems.reduce((sum, i) => sum + i.productPrice * i.quantity, 0);
  const allSelected = cart.items.length > 0 && cart.items.every((i) => i.selected || !i.available);

  const handleSelectAll = async () => {
    await cartApi.selectAll(!allSelected);
    await loadCart();
  };

  const handleSelect = async (item: CartItem) => {
    await cartApi.toggleSelect(item.id, !item.selected);
    await loadCart();
  };

  const handleQuantity = async (item: CartItem, delta: number) => {
    const q = Math.max(1, Math.min(50, item.quantity + delta));
    await cartApi.updateQuantity(item.id, q);
    await loadCart();
  };

  const handleRemove = async (id: number) => {
    await cartApi.removeItem(id);
    await loadCart();
  };

  const handleCheckout = () => {
    const ids = selectedItems.map((i) => i.productId).join(',');
    navigate(`/voyara/checkout?products=${ids}`);
  };

  if (loading) return <div className="vy-section"><div className="vy-container"><p>Loading...</p></div></div>;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '960px' }}>
        <h1 className="vy-heading h2">{t('cart.title')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}
        {cart.items.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '4rem 0', color: 'var(--vy-text-dim)' }}>
            <p style={{ fontSize: '1.2rem', marginBottom: '1rem' }}>{t('cart.empty')}</p>
            <Link to="/voyara" className="vy-btn vy-btn-primary">{t('cart.continueShopping')}</Link>
          </div>
        ) : (
          <>
            <div className="vy-cart-toolbar" style={{ display: 'flex', alignItems: 'center', gap: '1rem', padding: '0.75rem 0', borderBottom: '1px solid var(--vy-border)' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                <input type="checkbox" checked={allSelected} onChange={handleSelectAll} />
                {t('cart.selectAll')}
              </label>
              <span style={{ color: 'var(--vy-text-dim)', fontSize: '0.9rem' }}>({cart.count} {t('cart.items')})</span>
            </div>
            {groupedItems().map((group) => (
              <div key={group.sellerId} className="vy-cart-group" style={{ marginTop: '1rem', border: '1px solid var(--vy-border)', borderRadius: '8px', overflow: 'hidden' }}>
                <div style={{ padding: '0.75rem 1rem', fontWeight: 600, background: 'var(--vy-surface)' }}>
                  {group.shopName || `Seller #${group.sellerId}`}
                </div>
                {group.items.map((item) => (
                  <div key={item.id} className="vy-cart-item" style={{ display: 'flex', alignItems: 'center', gap: '1rem', padding: '1rem', borderTop: '1px solid var(--vy-border)' }}>
                    <input type="checkbox" checked={item.selected} onChange={() => handleSelect(item)} disabled={!item.available} />
                    <div style={{ width: '80px', height: '80px', background: 'var(--vy-surface)', borderRadius: '8px', overflow: 'hidden', flexShrink: 0 }}>
                      {item.productImage ? <img src={item.productImage} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : null}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <Link to={`/voyara/product/${item.productId}`} style={{ fontWeight: 500, color: 'inherit', textDecoration: 'none' }}>
                        {item.productTitle}
                      </Link>
                      {!item.available && <span style={{ color: 'var(--vy-error)', fontSize: '0.85rem', marginLeft: '0.5rem' }}>(Unavailable)</span>}
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <button className="vy-btn vy-btn-sm" onClick={() => handleQuantity(item, -1)} disabled={item.quantity <= 1}>-</button>
                      <span style={{ minWidth: '2rem', textAlign: 'center' }}>{item.quantity}</span>
                      <button className="vy-btn vy-btn-sm" onClick={() => handleQuantity(item, 1)} disabled={item.quantity >= 50}>+</button>
                    </div>
                    <div style={{ fontWeight: 600, minWidth: '80px', textAlign: 'right' }}>
                      ${(item.productPrice * item.quantity).toFixed(2)}
                    </div>
                    <button className="vy-btn vy-btn-sm" style={{ color: 'var(--vy-error)' }} onClick={() => handleRemove(item.id)}>
                      {t('cart.remove')}
                    </button>
                  </div>
                ))}
              </div>
            ))}
            <div className="vy-cart-footer" style={{ position: 'sticky', bottom: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem', background: 'var(--vy-surface)', borderTop: '1px solid var(--vy-border)', borderRadius: '0 0 8px 8px', marginTop: '1rem' }}>
              <div>
                <span style={{ color: 'var(--vy-text-dim)' }}>{t('cart.total')}: </span>
                <span style={{ fontSize: '1.3rem', fontWeight: 700, color: 'var(--vy-amber)' }}>${totalAmount.toFixed(2)}</span>
              </div>
              <button className="vy-btn vy-btn-primary vy-btn-lg" onClick={handleCheckout} disabled={selectedItems.length === 0}>
                {t('cart.checkout')} ({selectedItems.length})
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
