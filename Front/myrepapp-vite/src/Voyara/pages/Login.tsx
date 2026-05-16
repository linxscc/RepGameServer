import { useState, useEffect, useRef, type FormEvent } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { useAuth } from '../contexts/AuthContext';
import { voyaraApi } from '../api/client';

export default function Login() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const emailRef = useRef<HTMLInputElement>(null);

  // Auto-focus email on mount
  useEffect(() => { emailRef.current?.focus(); }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await voyaraApi.login({ email, password });
      login(res.token, res.refreshToken ?? '', res.user);
      // Preserve redirect: check state first, then search param, then default
      const stateRedirect = (window.history.state as Record<string, string>)?.from;
      const redirect = searchParams.get('redirect') || stateRedirect || '/voyara';
      navigate(redirect, { replace: true });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Login failed';
      if (msg.toLowerCase().includes('locked')) {
        setError(t('auth.accountLocked') || 'Account temporarily locked. Try again later.');
      } else {
        setError(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="vy-auth-page">
      <div className="vy-auth-card vy-card">
        <h1 className="vy-heading h3">{t('auth.login')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="vy-form-group">
            <label className="vy-label">{t('auth.email')}</label>
            <input ref={emailRef} className="vy-input" type="email" value={email}
              onChange={(e) => setEmail(e.target.value)} required autoComplete="email" />
          </div>
          <div className="vy-form-group">
            <label className="vy-label">{t('auth.password')}</label>
            <div className="vy-password-wrap">
              <input className="vy-input" type={showPwd ? 'text' : 'password'} value={password}
                onChange={(e) => setPassword(e.target.value)} required autoComplete="current-password" />
              <button type="button" className="vy-password-toggle" onClick={() => setShowPwd(!showPwd)}
                aria-label={showPwd ? 'Hide password' : 'Show password'}>
                {showPwd ? '🙈' : '👁'}
              </button>
            </div>
          </div>
          <div className="vy-field-row">
            <Link to="/voyara/forgot-password" className="vy-link-forgot">
              {t('auth.forgotPassword')}
            </Link>
          </div>
          <button className="vy-btn vy-btn-primary" style={{ width: '100%' }} disabled={loading}>
            {loading ? <span className="vy-spinner" /> : t('auth.login')}
          </button>
        </form>
        <p className="vy-auth-switch">
          {t('auth.noAccount')} <Link to="/voyara/register">{t('auth.register')}</Link>
        </p>
      </div>
    </div>
  );
}
