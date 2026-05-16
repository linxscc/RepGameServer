import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { useCart } from '../contexts/CartContext';
import { voyaraApi } from '../api/client';
import { cartApi } from '../api/cart';
import { LoadingState, ErrorState } from '../components/LoadingState';
import type { Product } from '../api/types';

export default function ProductDetail() {
  const { t } = useLanguage();
  const { refreshCount } = useCart();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeImg, setActiveImg] = useState(0);
  const [loadError, setLoadError] = useState('');
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [addingToCart, setAddingToCart] = useState(false);
  const [buyingNow, setBuyingNow] = useState(false);

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message });
    setTimeout(() => setToast(null), 2500);
  };

  const requireAuth = () => {
    const token = localStorage.getItem('voyara_token');
    if (!token) {
      navigate('/voyara/login?redirect=' + encodeURIComponent(window.location.pathname));
      return false;
    }
    return true;
  };

  const handleAddToCart = async () => {
    if (!requireAuth()) return;
    if (!product) return;
    setAddingToCart(true);
    try {
      await cartApi.addToCart(product.id);
      refreshCount();
      showToast('success', t('product.addedToCart'));
    } catch {
      showToast('error', t('product.addToCartFailed'));
    } finally {
      setAddingToCart(false);
    }
  };

  const handleBuyNow = async () => {
    if (!requireAuth()) return;
    if (!product) return;
    setBuyingNow(true);
    try {
      await cartApi.addToCart(product.id);
      navigate('/voyara/checkout?products=' + product.id);
    } catch {
      setBuyingNow(false);
      showToast('error', t('product.addToCartFailed'));
    }
  };

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    voyaraApi.getProduct(Number(id))
      .then(setProduct)
      .catch((err) => { setLoadError(err instanceof Error ? err.message : 'Product not found'); setProduct(null); })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <LoadingState />;
  if (loadError) return <ErrorState message={loadError} onBack={() => navigate(-1)} />;
  if (!product) return null;

  return (
    <div className="vy-product-detail">
      <div className="vy-container">
        <button className="vy-btn vy-btn-ghost vy-btn-sm" onClick={() => navigate(-1)} style={{ marginBottom: '2rem' }}>
          &larr; Back
        </button>

        <div className="vy-product-layout">
          {/* Gallery */}
          <div className="vy-product-gallery">
            {product.images?.[activeImg] ? (
              <img src={product.images[activeImg]} alt={product.title} className="vy-product-main-img" />
            ) : (
              <div className="vy-product-img-placeholder">📦</div>
            )}
            {product.images && product.images.length > 1 && (
              <div className="vy-product-thumbs">
                {product.images.map((img, i) => (
                  <button key={i} className={`vy-product-thumb${i === activeImg ? ' active' : ''}`} onClick={() => setActiveImg(i)}>
                    <img src={img} alt="" />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Info */}
          <div className="vy-product-info">
            <span className="vy-badge active">{t(`category.${product.category}`)}</span>
            <span className="vy-badge" style={{ marginLeft: '0.5rem' }}>{t(`condition.${product.condition}`)}</span>
            <h1 className="vy-heading h2" style={{ marginTop: '1.5rem' }}>{product.title}</h1>
            <p className="vy-product-desc">{product.description}</p>

            <div className="vy-product-price-block">
              <span className="vy-product-price-lg">${product.price.toLocaleString()}</span>
              <span className="vy-product-currency">USD</span>
            </div>

            {product.shopName && (
              <div className="vy-product-seller">
                <span className="vy-label">{t('product.seller')}</span>
                <span>{product.shopName}</span>
              </div>
            )}

            <div className="vy-product-actions">
              <button className="vy-btn vy-btn-outline vy-btn-lg" onClick={handleAddToCart} disabled={addingToCart || buyingNow}>
                {addingToCart ? <span className="vy-spinner" /> : t('product.addToCart')}
              </button>
              <button className="vy-btn vy-btn-primary vy-btn-lg" onClick={handleBuyNow} disabled={buyingNow || addingToCart}>
                {buyingNow ? <span className="vy-spinner" /> : t('product.buyNow')}
              </button>
            </div>
          </div>
        </div>
      </div>

      {toast && (
        <div className={`vy-toast vy-toast-${toast.type}`}>
          {toast.message}
        </div>
      )}
    </div>
  );
}
