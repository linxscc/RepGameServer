import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useLanguage } from '../../contexts/LanguageContext';
import { LoadingState, EmptyState, ErrorState } from '../../components/LoadingState';
import { voyaraApi } from '../../api/client';
import type { Product } from '../../api/types';

export default function MyProducts() {
  const { t } = useLanguage();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    voyaraApi.getSellerProducts()
      .then((res) => setProducts(res.items))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load products'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingState />;
  if (error) return <ErrorState message={error} />;

  return (
    <div className="vy-section">
      <div className="vy-container vy-seller-products-container" style={{ maxWidth: '900px' }}>
        <div className="vy-seller-page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1 className="vy-heading h2">{t('seller.myProducts')}</h1>
          <Link to="/voyara/seller/products/new" className="vy-btn vy-btn-primary vy-btn-sm">+ {t('seller.newProduct')}</Link>
        </div>
        {products.length === 0 ? (
          <EmptyState message="No listings yet" action={{ label: '+ ' + t('seller.newProduct'), href: '/voyara/seller/products/new' }} />
        ) : (
          <div className="vy-my-products" style={{ marginTop: '2rem' }}>
            {products.map((p) => (
              <div key={p.id} className="vy-my-product-item vy-card">
                <div className="vy-my-product-row" style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
                  {p.images?.[0] ? (
                    <img src={p.images[0]} alt="" style={{ width: '80px', height: '80px', objectFit: 'cover' }} />
                  ) : (
                    <div style={{ width: '80px', height: '80px', background: 'var(--vy-surface-elevated)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '1.5rem' }}>📦</div>
                  )}
                  <div style={{ flex: 1 }}>
                    <h3 style={{ fontFamily: 'var(--vy-font-display)', fontSize: '1.1rem' }}>{p.title}</h3>
                    <span style={{ color: 'var(--vy-amber)', fontWeight: 500 }}>${p.price.toLocaleString()}</span>
                    <span className="vy-badge" style={{ marginLeft: '0.75rem' }}>{t(`condition.${p.condition}`)}</span>
                  </div>
                  <span className={`vy-badge${p.status === 'active' ? ' active' : ''}`}>{p.status}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
