import { useState, useEffect, useCallback } from 'react';
import { LoadingState, EmptyState, ErrorState } from '../../components/LoadingState';
import { voyaraApi } from '../../api/client';
import type { Product } from '../../api/types';

type FilterStatus = '' | 'in_review' | 'active' | 'inactive';

export default function PendingProducts() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<FilterStatus>('in_review');
  const [actionMsg, setActionMsg] = useState('');

  const load = useCallback(() => {
    setLoading(true);
    setError('');
    voyaraApi.getAdminProducts(filter || undefined)
      .then((res) => setProducts(res.items))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load'))
      .finally(() => setLoading(false));
  }, [filter]);

  useEffect(() => { load(); }, [load]);

  const updateStatus = async (id: number, status: 'active' | 'inactive') => {
    setActionMsg('');
    try {
      await voyaraApi.updateProductStatus(id, status);
      setProducts((prev) => prev.filter((p) => p.id !== id));
      setActionMsg(`Product #${id} ${status === 'active' ? 'approved' : 'rejected'}`);
    } catch (err) {
      setActionMsg(err instanceof Error ? err.message : 'Update failed');
    }
  };

  const statusLabel = (s: string) => {
    switch (s) {
      case 'active': return 'Active';
      case 'inactive': return 'Inactive';
      case 'in_review': return 'In Review';
      case 'sold': return 'Sold';
      default: return s;
    }
  };

  if (loading && products.length === 0) return <LoadingState />;
  if (error && products.length === 0) return <ErrorState message={error} onRetry={load} />;

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '1000px' }}>
        <h1 className="vy-heading h2" style={{ marginBottom: '0.5rem' }}>Admin — Product Review</h1>

        {/* Status filter tabs */}
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem' }}>
          {(['in_review', '', 'active', 'inactive'] as FilterStatus[]).map((f) => (
            <button
              key={f}
              className={`vy-btn vy-btn-sm${filter === f ? ' vy-btn-primary' : ''}`}
              onClick={() => setFilter(f)}
            >
              {f === '' ? 'All' : statusLabel(f)}
            </button>
          ))}
        </div>

        {actionMsg && (
          <div style={{
            padding: '0.75rem 1rem', borderRadius: '8px', marginBottom: '1rem',
            background: actionMsg.includes('approved') || actionMsg.includes('rejected')
              ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
            color: actionMsg.includes('approved') || actionMsg.includes('rejected')
              ? '#22c55e' : '#ef4444',
          }}>
            {actionMsg}
          </div>
        )}

        {products.length === 0 ? (
          <EmptyState message={`No products in ${filter || 'any'} status`} />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {products.map((p) => (
              <div key={p.id} className="vy-card" style={{ padding: '1rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
                <div style={{ width: '70px', height: '70px', borderRadius: '6px', overflow: 'hidden', flexShrink: 0, background: 'var(--vy-surface)' }}>
                  {p.images?.[0] ? (
                    <img src={p.images[0]} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  ) : (
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: 'var(--vy-text-dim)' }}>No img</div>
                  )}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 600, marginBottom: '0.25rem' }}>{p.title}</div>
                  <div style={{ fontSize: '0.85rem', color: 'var(--vy-text-dim)' }}>
                    Seller: {p.sellerName || p.shopName || `#${p.sellerId}`} &middot; ${p.price.toLocaleString()} &middot; {p.category}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--vy-text-dim)', marginTop: '0.15rem' }}>
                    Created: {p.createdAt ? new Date(p.createdAt).toLocaleDateString() : 'N/A'}
                  </div>
                </div>
                <span className={`vy-badge${p.status === 'active' ? ' active' : ''}`} style={{
                  background: p.status === 'in_review' ? 'rgba(234,179,8,0.15)' : undefined,
                  color: p.status === 'in_review' ? '#eab308' : undefined,
                }}>
                  {statusLabel(p.status)}
                </span>
                {p.status === 'in_review' && (
                  <div style={{ display: 'flex', gap: '0.5rem', flexShrink: 0 }}>
                    <button className="vy-btn vy-btn-sm" style={{ background: 'rgba(34,197,94,0.15)', color: '#22c55e', border: '1px solid rgba(34,197,94,0.3)' }}
                      onClick={() => updateStatus(p.id, 'active')}>Approve</button>
                    <button className="vy-btn vy-btn-sm" style={{ background: 'rgba(239,68,68,0.1)', color: '#ef4444', border: '1px solid rgba(239,68,68,0.25)' }}
                      onClick={() => updateStatus(p.id, 'inactive')}>Reject</button>
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
