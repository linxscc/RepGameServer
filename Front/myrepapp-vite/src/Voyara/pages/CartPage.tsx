import { useState, useEffect, useRef, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { useCart } from '../contexts/CartContext';
import { LoadingState, EmptyState } from '../components/LoadingState';
import { cartApi } from '../api/cart';
import type { CartItem } from '../api/types';

export default function CartPage() {
  const { t } = useLanguage();
  const { refreshCount } = useCart();
  const navigate = useNavigate();
  const [items, setItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [error, setError] = useState('');
  const [dirty, setDirty] = useState(false);
  const dirtyRef = useRef(false);
  const origRef = useRef<CartItem[]>([]);

  // Load cart from API on mount
  useEffect(() => {
    cartApi.getCart()
      .then((data) => {
        setItems(data.items);
        origRef.current = JSON.parse(JSON.stringify(data.items));
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load cart'))
      .finally(() => setLoading(false));
  }, []);

  // Sync to backend — called before checkout or on unmount
  const syncCart = useCallback(async () => {
    if (!dirtyRef.current) return;
    setSyncing(true);
    const orig = origRef.current;
    const cur = items; // current state at sync time
    const origMap = new Map(orig.map((i) => [i.id, i]));
    const curMap = new Map(cur.map((i) => [i.id, i]));

    const ops: Promise<unknown>[] = [];

    // Removed items
    for (const [id] of origMap) {
      if (!curMap.has(id)) {
        ops.push(cartApi.removeItem(id).catch(() => {}));
      }
    }

    // Modified items
    for (const item of cur) {
      const origItem = origMap.get(item.id);
      if (!origItem) continue;
      if (origItem.quantity !== item.quantity) {
        ops.push(cartApi.updateQuantity(item.id, item.quantity).catch(() => {}));
      }
      if (origItem.selected !== item.selected) {
        ops.push(cartApi.toggleSelect(item.id, item.selected).catch(() => {}));
      }
    }

    await Promise.all(ops);
    dirtyRef.current = false;
    setDirty(false);
    origRef.current = JSON.parse(JSON.stringify(cur));
    refreshCount();
    setSyncing(false);
  }, [items]);

  // ── Local mutations only ──

  const markDirty = (next: CartItem[]) => {
    dirtyRef.current = true;
    setDirty(true);
    setItems(next);
  };

  const handleSelectAll = () => {
    const allSelected = items.length > 0 && items.every((i) => i.selected || !i.available);
    markDirty(items.map((i) => (i.available ? { ...i, selected: !allSelected } : i)));
  };

  const handleSelect = (item: CartItem) => {
    markDirty(items.map((i) => (i.id === item.id ? { ...i, selected: !i.selected } : i)));
  };

  const handleQuantity = (item: CartItem, delta: number) => {
    const q = Math.max(1, Math.min(50, item.quantity + delta));
    markDirty(items.map((i) => (i.id === item.id ? { ...i, quantity: q } : i)));
  };

  const handleRemove = async (id: number) => {
    try {
      await cartApi.removeItem(id);
      origRef.current = origRef.current.filter((i) => i.id !== id);
      setItems((prev) => prev.filter((i) => i.id !== id));
      refreshCount();
    } catch {
      // backend sync failed — item stays in local state
    }
  };

  const handleCheckout = async () => {
    const selected = items.filter((i) => i.selected && i.available);
    if (selected.length === 0) return;
    await syncCart();
    navigate('/voyara/checkout', { state: { products: selected.map((i) => i.productId) } });
  };

  // ── Derived state ──

  const selectedItems = items.filter((i) => i.selected && i.available);
  const totalAmount = selectedItems.reduce((sum, i) => sum + i.productPrice * i.quantity, 0);
  const allSelected = items.length > 0 && items.every((i) => i.selected || !i.available);

  const groupedItems = () => {
    const groups: Record<number, { sellerId: number; shopName: string; items: CartItem[] }> = {};
    for (const item of items) {
      if (!groups[item.sellerId]) {
        groups[item.sellerId] = { sellerId: item.sellerId, shopName: item.sellerShopName, items: [] };
      }
      groups[item.sellerId].items.push(item);
    }
    return Object.values(groups);
  };

  if (loading) return <LoadingState />;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '960px' }}>
        <h1 className="vy-heading h2">{t('cart.title')}</h1>
        {error && <div className="vy-auth-error" style={{ marginTop: '1rem' }}>{error}</div>}
        {items.length === 0 ? (
          <EmptyState message={t('cart.empty')} action={{ label: t('cart.continueShopping'), href: '/voyara' }} />
        ) : (
          <>
            <div className="vy-cart-toolbar" style={{ display: 'flex', alignItems: 'center', gap: '1rem', padding: '0.75rem 0', borderBottom: '1px solid var(--vy-border)' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                <input type="checkbox" checked={allSelected} onChange={handleSelectAll} />
                {t('cart.selectAll')}
              </label>
              <span style={{ color: 'var(--vy-text-dim)', fontSize: '0.9rem' }}>({items.length} {t('cart.items')})</span>
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
                    <button className="vy-btn vy-btn-sm" style={{ color: 'var(--vy-error)', border: '1px solid var(--vy-error)' }} onClick={() => handleRemove(item.id)}>
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
                {dirty && <span style={{ fontSize: '0.75rem', color: 'var(--vy-text-dim)', marginLeft: '0.5rem' }}>(unsaved)</span>}
              </div>
              <button className="vy-btn vy-btn-primary vy-btn-lg" onClick={handleCheckout} disabled={selectedItems.length === 0 || syncing}>
                {syncing ? <span className="vy-spinner" /> : `${t('cart.checkout')} (${selectedItems.length})`}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
