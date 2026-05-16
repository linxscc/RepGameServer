import { useState, useRef, useEffect, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { useAuth } from '../contexts/AuthContext';
import { voyaraApi } from '../api/client';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;
const RESEND_COOLDOWN = 60;

export default function Register() {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { login } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [code, setCode] = useState('');
  const [codeSent, setCodeSent] = useState(false);
  const [codeSending, setCodeSending] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const emailRef = useRef<HTMLInputElement>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => { emailRef.current?.focus(); }, []);

  useEffect(() => {
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, []);

  const startCountdown = () => {
    setCountdown(RESEND_COOLDOWN);
    timerRef.current = setInterval(() => {
      setCountdown((c) => {
        if (c <= 1) { if (timerRef.current !== null) clearInterval(timerRef.current); return 0; }
        return c - 1;
      });
    }, 1000);
  };

  const handleSendCode = async () => {
    if (!EMAIL_REGEX.test(email)) {
      setError(t('validation.invalidEmail'));
      return;
    }
    setError('');
    setCodeSending(true);
    try {
      await voyaraApi.sendVerificationCode(email, 'register');
      setCodeSent(true);
      startCountdown();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to send code');
    } finally {
      setCodeSending(false);
    }
  };

  const calcStrength = (p: string): { label: string; pct: number; cls: string } => {
    if (!p) return { label: '', pct: 0, cls: '' };
    let score = 0;
    if (p.length >= 8) score += 25;
    if (p.length >= 12) score += 10;
    if (/[a-z]/.test(p) && /[A-Z]/.test(p)) score += 25;
    if (/\d/.test(p)) score += 20;
    if (/[^a-zA-Z0-9]/.test(p)) score += 20;
    if (score < 40) return { label: 'Weak', pct: Math.max(score, 15), cls: 'weak' };
    if (score < 70) return { label: 'Medium', pct: score, cls: 'medium' };
    return { label: 'Strong', pct: score, cls: 'strong' };
  };
  const strength = calcStrength(password);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (!EMAIL_REGEX.test(email)) {
      setError(t('validation.invalidEmail'));
      return;
    }
    if (!PASSWORD_REGEX.test(password)) {
      setError(t('auth.passwordRequirements'));
      return;
    }
    if (!codeSent) {
      setError('Please request a verification code first');
      return;
    }
    if (code.length !== 6) {
      setError(t('validation.invalidCode'));
      return;
    }

    setLoading(true);
    try {
      const res = await voyaraApi.register({ email, password, name, code });
      login(res.token, res.refreshToken ?? '', res.user);
      navigate('/voyara');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Register failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="vy-auth-page">
      <div className="vy-auth-card vy-card">
        <h1 className="vy-heading h3">{t('auth.register')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="vy-form-group">
            <label className="vy-label">{t('auth.name')}</label>
            <input ref={emailRef} className="vy-input" type="text" value={name}
              onChange={(e) => setName(e.target.value)} required minLength={1} />
          </div>
          <div className="vy-form-group">
            <label className="vy-label">{t('auth.email')}</label>
            <input className={`vy-input${email && !EMAIL_REGEX.test(email) ? ' vy-input-error' : ''}`}
              type="email" value={email}
              onChange={(e) => { setEmail(e.target.value); setCodeSent(false); }}
              required placeholder="you@example.com" />
            {email && !EMAIL_REGEX.test(email) && (
              <span className="vy-field-hint">{t('validation.invalidEmail')}</span>
            )}
          </div>
          <div className="vy-form-group">
            <label className="vy-label">{t('auth.password')}</label>
            <div className="vy-password-wrap">
              <input className={`vy-input${password && !PASSWORD_REGEX.test(password) ? ' vy-input-error' : ''}`}
                type={showPwd ? 'text' : 'password'} value={password}
                onChange={(e) => setPassword(e.target.value)} required minLength={8}
                autoComplete="new-password" />
              <button type="button" className="vy-password-toggle" onClick={() => setShowPwd(!showPwd)}
                aria-label={showPwd ? 'Hide password' : 'Show password'}>
                {showPwd ? '🙈' : '👁'}
              </button>
            </div>
            {password && !PASSWORD_REGEX.test(password) && (
              <span className="vy-field-hint">{t('auth.passwordRequirements')}</span>
            )}
            {password && (
              <div className="vy-pwd-strength">
                <div className="vy-pwd-strength-bar">
                  <div className={`vy-pwd-strength-fill ${strength.cls}`} style={{ width: `${strength.pct}%` }} />
                </div>
                <span className={`vy-pwd-strength-label ${strength.cls}`}>{strength.label}</span>
              </div>
            )}
          </div>

          {!codeSent ? (
            <button type="button" className="vy-btn vy-btn-outline" style={{ width: '100%' }}
              onClick={handleSendCode} disabled={codeSending || !email || !EMAIL_REGEX.test(email)}>
              {codeSending ? <span className="vy-spinner" /> : t('auth.sendCode')}
            </button>
          ) : (
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.verificationCode')}</label>
              <input className="vy-input" type="text" value={code}
                onChange={(e) => setCode(e.target.value)} required maxLength={6}
                placeholder="6-digit code" />
              <div style={{ marginTop: '0.5rem', fontSize: '0.85rem' }}>
                {countdown > 0 ? (
                  <span style={{ color: '#888' }}>{t('auth.resendCode')} ({countdown}s)</span>
                ) : (
                  <button type="button" className="vy-link-btn" onClick={handleSendCode} disabled={codeSending}>
                    {t('auth.resendCode')}
                  </button>
                )}
              </div>
            </div>
          )}

          <button className="vy-btn vy-btn-primary" style={{ width: '100%', marginTop: '1rem' }}
            disabled={loading || !codeSent}>
            {loading ? <span className="vy-spinner" /> : t('auth.register')}
          </button>
        </form>
        <p className="vy-auth-switch">
          {t('auth.hasAccount')} <Link to="/voyara/login">{t('auth.login')}</Link>
        </p>
      </div>
    </div>
  );
}
