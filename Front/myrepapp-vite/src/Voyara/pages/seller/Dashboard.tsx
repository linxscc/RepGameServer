import { Link } from 'react-router-dom';
import { useLanguage } from '../../contexts/LanguageContext';

export default function SellerDashboard() {
  const { t } = useLanguage();
  const token = localStorage.getItem('voyara_token');
  if (!token) {
    return (
      <div className="vy-section">
        <div className="vy-container" style={{ textAlign: 'center' }}>
          <h2 className="vy-heading h2">Seller Dashboard</h2>
          <p style={{ color: 'var(--vy-text-dim)', marginTop: '1rem' }}>
            Please <Link to="/voyara/login" style={{ color: 'var(--vy-amber)' }}>sign in</Link> to access seller features.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '800px' }}>
        <h1 className="vy-heading h2">{t('seller.dashboard')}</h1>
        <div className="vy-grid-2" style={{ marginTop: '2.5rem' }}>
          <Link to="/voyara/seller/products/new" className="vy-seller-card vy-card">
            <span className="vy-seller-icon">+</span>
            <h3 className="vy-heading h3">{t('seller.newProduct')}</h3>
            <p style={{ color: 'var(--vy-text-dim)', marginTop: '0.5rem' }}>List a new item for sale</p>
          </Link>
          <Link to="/voyara/seller/products" className="vy-seller-card vy-card">
            <span className="vy-seller-icon">📋</span>
            <h3 className="vy-heading h3">{t('seller.myProducts')}</h3>
            <p style={{ color: 'var(--vy-text-dim)', marginTop: '0.5rem' }}>Manage your listings</p>
          </Link>
        </div>
      </div>
    </div>
  );
}
