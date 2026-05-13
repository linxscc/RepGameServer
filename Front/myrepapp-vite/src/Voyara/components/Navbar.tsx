import { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { supportedLangs, type Lang } from '../i18n';
import { useLanguage } from '../contexts/LanguageContext';
import { useCart } from '../contexts/CartContext';

export default function Navbar() {
  const location = useLocation();
  const { lang, setLang, t } = useLanguage();
  const [showLang, setShowLang] = useState(false);
  const { count } = useCart();
  const token = localStorage.getItem('voyara_token');
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const handleLang = (code: Lang) => {
    setLang(code);
    setShowLang(false);
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <nav className={`vy-navbar${scrolled ? ' scrolled' : ''}`}>
      <div className="vy-navbar-inner">
        <Link to="/voyara" className="vy-logo">
          <span className="vy-logo-mark" aria-hidden="true" />
          {t('site.name')}
        </Link>

        <div className="vy-nav-links">
          <Link to="/voyara" className={`vy-nav-link${isActive('/voyara') ? ' active' : ''}`}>
            {t('nav.browse')}
          </Link>
          <Link to="/voyara/cart" className={`vy-nav-link${isActive('/voyara/cart') ? ' active' : ''}`}>
            Cart {count > 0 && <span className="vy-badge">{count > 99 ? '99+' : count}</span>}
          </Link>
          {token && (
            <Link to="/voyara/orders" className={`vy-nav-link${isActive('/voyara/orders') ? ' active' : ''}`}>
              {t('nav.orders')}
            </Link>
          )}
          <Link to="/voyara/seller" className={`vy-nav-link${isActive('/voyara/seller') ? ' active' : ''}`}>
            {t('nav.sell')}
          </Link>

          <div className="vy-lang-switcher">
            <button className="vy-nav-link vy-lang-btn" onClick={() => setShowLang(!showLang)}>
              {lang.toUpperCase()}
            </button>
            {showLang && (
              <div className="vy-lang-dropdown">
                {supportedLangs.map((l) => (
                  <button key={l.code} className={`vy-lang-item${l.code === lang ? ' active' : ''}`} onClick={() => handleLang(l.code)}>
                    {l.label}
                  </button>
                ))}
              </div>
            )}
          </div>

          {token ? (
            <button className="vy-btn vy-btn-outline vy-btn-sm" onClick={() => { localStorage.removeItem('voyara_token'); window.location.reload(); }}>
              {t('nav.logout')}
            </button>
          ) : (
            <Link to="/voyara/login" className="vy-btn vy-btn-primary vy-btn-sm">
              {t('nav.login')}
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
