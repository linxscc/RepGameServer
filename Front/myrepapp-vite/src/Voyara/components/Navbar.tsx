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
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  useEffect(() => {
    setMobileMenuOpen(false);
    setShowLang(false);
    setShowUserMenu(false);
  }, [location.pathname]);

  useEffect(() => {
    document.body.classList.toggle('vy-mobile-menu-lock', mobileMenuOpen);
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMobileMenuOpen(false);
    };
    if (mobileMenuOpen) {
      document.addEventListener('keydown', onKeyDown);
    }
    return () => {
      document.body.classList.remove('vy-mobile-menu-lock');
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [mobileMenuOpen]);

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
  const closeMobileMenu = () => setMobileMenuOpen(false);

  const renderNavItems = () => (
    <>
      <Link to="/voyara" className={`vy-nav-link${isActive('/voyara') ? ' active' : ''}`} onClick={closeMobileMenu}>
        {t('nav.browse')}
      </Link>
      <Link to="/voyara/cart" className={`vy-nav-link${isActive('/voyara/cart') ? ' active' : ''}`} onClick={closeMobileMenu}>
        {t('nav.cart')} {count > 0 && <span className="vy-badge">{count > 99 ? '99+' : count}</span>}
      </Link>
      {isAuthenticated && (
        <Link to="/voyara/orders" className={`vy-nav-link${isActive('/voyara/orders') ? ' active' : ''}`} onClick={closeMobileMenu}>
          {t('nav.orders')}
        </Link>
      )}
      {user?.role === 'admin' && (
        <Link
          to="/voyara/admin/products"
          className={`vy-nav-link${isActive('/voyara/admin/products') ? ' active' : ''}`}
          style={{ color: 'var(--vy-amber)', fontWeight: 600 }}
          onClick={closeMobileMenu}
        >
          Admin
        </Link>
      )}
      <Link to="/voyara/seller" className={`vy-nav-link${isActive('/voyara/seller') ? ' active' : ''}`} onClick={closeMobileMenu}>
        {t('nav.sell')}
      </Link>
    </>
  );

  return (
    <nav className={`vy-navbar${scrolled ? ' scrolled' : ''}`}>
      <div className="vy-navbar-inner">
        <Link to="/voyara" className="vy-logo">
          <span className="vy-logo-mark" aria-hidden="true" />
          {t('site.name')}
        </Link>

        <div className="vy-nav-links">
          {renderNavItems()}

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

        <div className="vy-mobile-actions">
          <Link
            to="/voyara/cart"
            className={`vy-mobile-cart${isActive('/voyara/cart') ? ' active' : ''}`}
            aria-label={t('nav.cart')}
          >
            <span>Cart</span>
            {count > 0 && <span className="vy-mobile-cart-badge">{count > 99 ? '99+' : count}</span>}
          </Link>
          <button
            className={`vy-hamburger${mobileMenuOpen ? ' open' : ''}`}
            type="button"
            aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
            aria-expanded={mobileMenuOpen}
            onClick={() => setMobileMenuOpen((open) => !open)}
          >
            <span />
            <span />
            <span />
          </button>
        </div>
      </div>

      <button
        className={`vy-mobile-overlay${mobileMenuOpen ? ' open' : ''}`}
        type="button"
        aria-label="Close menu"
        onClick={() => setMobileMenuOpen(false)}
      />

      <aside className={`vy-mobile-drawer${mobileMenuOpen ? ' open' : ''}`} aria-hidden={!mobileMenuOpen}>
        <div className="vy-mobile-drawer-header">
          <span className="vy-mobile-drawer-title">{t('site.name')}</span>
          <button className="vy-mobile-close" type="button" onClick={() => setMobileMenuOpen(false)}>
            Close
          </button>
        </div>

        <div className="vy-mobile-nav">{renderNavItems()}</div>

        <div className="vy-mobile-lang">
          <span className="vy-mobile-section-label">Language</span>
          <div className="vy-mobile-lang-grid">
            {supportedLangs.map((l) => (
              <button
                key={l.code}
                className={`vy-lang-item${l.code === lang ? ' active' : ''}`}
                onClick={() => handleLang(l.code)}
              >
                {l.label}
              </button>
            ))}
          </div>
        </div>

        <div className="vy-mobile-account">
          {isAuthenticated ? (
            <>
              <span className="vy-mobile-section-label">
                Hello, {user?.name?.split(' ')[0] || user?.name}
                {user && !user.emailVerified && <span className="vy-badge-warn" title="Email not verified">!</span>}
              </span>
              <Link to="/voyara/account" className="vy-user-dropdown-item" onClick={closeMobileMenu}>
                {t('nav.account')}
              </Link>
              <button className="vy-user-dropdown-item" onClick={() => { logout(); window.location.href = '/voyara'; }}>
                {t('nav.logout')}
              </button>
            </>
          ) : (
            <Link to="/voyara/login" className="vy-btn vy-btn-primary" onClick={closeMobileMenu}>
              {t('nav.login')}
            </Link>
          )}
        </div>
      </aside>
    </nav>
  );
}
