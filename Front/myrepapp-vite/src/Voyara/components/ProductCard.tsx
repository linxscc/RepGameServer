import { Link } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import type { Product } from '../api/types';

interface Props {
  product: Product;
}

const categoryEmoji: Record<string, string> = {
  appliance: '⚡',
  vehicle: '🚗',
  electronics: '📱',
  other: '📦',
};

export default function ProductCard({ product }: Props) {
  const { t } = useLanguage();
  return (
    <Link to={`/voyara/product/${product.id}`} className="vy-product-card vy-card">
      <div className="vy-product-card-img">
        {product.images?.[0] ? (
          <img src={product.images[0]} alt={product.title} loading="lazy" />
        ) : (
          <div className="vy-product-card-placeholder">
            <span>{categoryEmoji[product.category] ?? '📦'}</span>
          </div>
        )}
        <span className="vy-product-card-badge">{t(`condition.${product.condition}`)}</span>
      </div>
      <div className="vy-product-card-body">
        <span className="vy-product-card-category">{t(`category.${product.category}`)}</span>
        <h3 className="vy-product-card-title">{product.title}</h3>
        <div className="vy-product-card-footer">
          <span className="vy-product-card-price">
            <span className="vy-price-symbol">$</span>
            {product.price.toLocaleString()}
          </span>
          {product.shopName && (
            <span className="vy-product-card-shop">{product.shopName}</span>
          )}
        </div>
      </div>
    </Link>
  );
}
