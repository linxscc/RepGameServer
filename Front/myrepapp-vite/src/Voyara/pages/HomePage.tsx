import { useState, useEffect, useCallback } from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { voyaraApi } from '../api/client';
import ProductCard from '../components/ProductCard';
import { LoadingState, EmptyState, ErrorState } from '../components/LoadingState';
import type { Product, ProductCategory, ProductCondition } from '../api/types';

const CATEGORIES: { key: ProductCategory; icon: string }[] = [
  { key: 'appliance', icon: '⚡' },
  { key: 'vehicle', icon: '🚗' },
  { key: 'electronics', icon: '📱' },
  { key: 'other', icon: '📦' },
];

const CONDITIONS: ProductCondition[] = ['new', 'like_new', 'used', 'refurbished'];

const PAGE_SIZE = 20;

export default function HomePage() {
  const { t } = useLanguage();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [selectedCat, setSelectedCat] = useState<ProductCategory | ''>('');
  const [selectedCond, setSelectedCond] = useState<ProductCondition | ''>('');
  const [search, setSearch] = useState('');
  const [minPrice, setMinPrice] = useState('');
  const [maxPrice, setMaxPrice] = useState('');
  const [loadError, setLoadError] = useState('');

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const load = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const filter: Record<string, string> = {};
      if (selectedCat) filter.category = selectedCat;
      if (selectedCond) filter.condition = selectedCond;
      if (search) filter.search = search;
      if (minPrice) filter.minPrice = minPrice;
      if (maxPrice) filter.maxPrice = maxPrice;
      filter.page = String(p);
      filter.pageSize = String(PAGE_SIZE);
      const res = await voyaraApi.getProducts(filter);
      setProducts(res.items);
      setTotal(res.total);
      setPage(res.page);
      setLoadError('');
    } catch (err: unknown) { setProducts([]); setTotal(0); setLoadError(err instanceof Error ? err.message : 'Failed to load products'); }
      finally { setLoading(false); }
  }, [selectedCat, selectedCond, search, minPrice, maxPrice]);

  useEffect(() => { load(1); }, [load]);

  return (
    <div>
      {/* Hero */}
      <section className="vy-hero">
        <div className="vy-hero-bg" />
        <div className="vy-hero-content">
          <div className="vy-hero-rule" />
          <h1 className="vy-heading h1 vy-hero-title">{t('hero.title')}</h1>
          <p className="vy-hero-sub">{t('hero.subtitle')}</p>
          <div className="vy-hero-search">
            <input className="vy-input" placeholder={t('search.placeholder')} value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>
        </div>
        <div className="vy-hero-scroll" />
      </section>

      {/* Filters */}
      <section className="vy-section">
        <div className="vy-container">
          <div className="vy-filters">
            <div className="vy-filter-cats">
              <button className={`vy-filter-chip${selectedCat === '' ? ' active' : ''}`} onClick={() => setSelectedCat('')}>{t('filter.all')}</button>
              {CATEGORIES.map((c) => (
                <button key={c.key} className={`vy-filter-chip${selectedCat === c.key ? ' active' : ''}`} onClick={() => setSelectedCat(c.key)}>
                  {c.icon} {t(`category.${c.key}`)}
                </button>
              ))}
            </div>
            <div className="vy-filter-row">
              <div className="vy-filter-conds">
                <button className={`vy-filter-chip sm${selectedCond === '' ? ' active' : ''}`} onClick={() => setSelectedCond('')}>{t('condition.new')}/{t('filter.all')}</button>
                {CONDITIONS.map((c) => (
                  <button key={c} className={`vy-filter-chip sm${selectedCond === c ? ' active' : ''}`} onClick={() => setSelectedCond(c)}>
                    {t(`condition.${c}`)}
                  </button>
                ))}
              </div>
              <div className="vy-filter-prices">
                <input className="vy-input vy-price-input" type="number" placeholder="Min $" value={minPrice} onChange={(e) => setMinPrice(e.target.value)} />
                <span className="vy-price-sep">&ndash;</span>
                <input className="vy-input vy-price-input" type="number" placeholder="Max $" value={maxPrice} onChange={(e) => setMaxPrice(e.target.value)} />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Product Grid */}
      <section className="vy-section" style={{ paddingTop: 0 }}>
        <div className="vy-container">
          {loadError ? (
            <ErrorState message={loadError} onRetry={() => load(page)} />
          ) : loading ? (
            <LoadingState />
          ) : products.length === 0 ? (
            <EmptyState message="No products found" action={{ label: t('hero.cta'), onClick: () => { setSelectedCat(''); setSearch(''); setMinPrice(''); setMaxPrice(''); } }} />
          ) : (
            <div className="vy-grid-4">
              {products.map((p) => <ProductCard key={p.id} product={p} />)}
            </div>
          )}

          {/* Pagination */}
          {total > PAGE_SIZE && (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', marginTop: '2rem' }}>
              <button
                className="vy-btn vy-btn-secondary"
                disabled={page <= 1}
                onClick={() => load(page - 1)}
              >
                {t('pagination.prev', 'Prev')}
              </button>
              <span style={{ color: 'var(--vy-text-dim)', fontSize: '0.9rem' }}>
                {page} / {totalPages}
              </span>
              <button
                className="vy-btn vy-btn-secondary"
                disabled={page >= totalPages}
                onClick={() => load(page + 1)}
              >
                {t('pagination.next', 'Next')}
              </button>
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
