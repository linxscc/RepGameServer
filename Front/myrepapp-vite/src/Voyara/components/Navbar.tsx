import { useState, useEffect, useRef } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { useAuth } from '../contexts/AuthContext';
import { useCart } from '../contexts/CartContext';
import { supportedLangs, type Lang } from '../i18n';

export default function Navbar() {
  const location = useLocation();
  const { lang, setLang, t } = useLanguage();
  const { user, isAuthenticated, logout } = useAuth();
  const { count } = useCart();
  const [showLang, setShowLang] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  // Close user menu on outside click
  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowUserMenu(false);
      }
    };
    if (showUserMenu) {
      document.addEventListener('mousedown', onClick);
      return () => document.removeEventListener('mousedown', onClick);
    }
  }, [showUserMenu]);

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
            {t('nav.cart')} {count > 0 && <span className="vy-badge">{count > 99 ? '99+' : count}</span>}
          </Link>
          {isAuthenticated && (
            <Link to="/voyara/orders" className={`vy-nav-link${isActive('/voyara/orders') ? ' active' : ''}`}>
              {t('nav.orders')}
            </Link>
          )}
          {user?.role === 'admin' && (
            <Link to="/voyara/admin/products" className={`vy-nav-link${isActive('/voyara/admin/products') ? ' active' : ''}`}
              style={{ color: 'var(--vy-amber)', fontWeight: 600 }}>
              Admin
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

          {isAuthenticated ? (
            <div className="vy-user-menu" ref={menuRef}>
              <button className="vy-nav-link vy-user-btn" onClick={() => setShowUserMenu(!showUserMenu)}>
                <span className="vy-user-greeting">
                  Hello, {user?.name?.split(' ')[0] || user?.name}
                  {user && !user.emailVerified && <span className="vy-badge-warn" title="Email not verified">!</span>}
                </span>
                <span className="vy-user-arrow">&#9662;</span>
              </button>
              {showUserMenu && (
                <div className="vy-user-dropdown">
                  <div className="vy-user-dropdown-header">
                    <div className="vy-user-dropdown-name">{user?.name}</div>
                    <div className="vy-user-dropdown-email">{user?.email}</div>
                  </div>
                  <div className="vy-user-dropdown-items">
                    <Link to="/voyara/orders" className="vy-user-dropdown-item" onClick={() => setShowUserMenu(false)}>
                      {t('nav.orders')}
                    </Link>
                    <Link to="/voyara/account" className="vy-user-dropdown-item" onClick={() => setShowUserMenu(false)}>
                      {t('nav.account')}
                    </Link>
                    <div className="vy-user-dropdown-divider" />
                    <button className="vy-user-dropdown-item" onClick={() => { logout(); window.location.href = '/voyara'; }}>
                      {t('nav.logout')}
                    </button>
                  </div>
                </div>
              )}
            </div>
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
