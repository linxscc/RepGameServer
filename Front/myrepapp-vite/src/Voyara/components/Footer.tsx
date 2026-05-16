import { useLanguage } from '../contexts/LanguageContext';

export default function Footer() {
  const { t } = useLanguage();
  return (
    <footer className="vy-footer">
      <div className="vy-footer-rule" />
      <div className="vy-container vy-footer-inner">
        <div className="vy-footer-brand">
          <span className="vy-footer-logo">{t('site.name')}</span>
          <p>{t('footer.tagline')}</p>
        </div>
        <div className="vy-footer-bottom">
          <span>&copy; {new Date().getFullYear()} Voyara. {t('footer.rights')}</span>
        </div>
      </div>
    </footer>
  );
}
