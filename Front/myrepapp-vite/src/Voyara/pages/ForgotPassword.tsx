import { useState, useEffect, useRef, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { voyaraApi } from '../api/client';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;

type Step = 'email' | 'reset';

export default function ForgotPassword() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const emailRef = useRef<HTMLInputElement>(null);

  useEffect(() => { emailRef.current?.focus(); }, []);

  const handleSendCode = async (e: FormEvent) => {
    e.preventDefault();
    if (!EMAIL_REGEX.test(email)) {
      setError(t('validation.invalidEmail'));
      return;
    }
    setError('');
    setLoading(true);
    try {
      await voyaraApi.forgotPassword(email);
      setMessage(t('auth.codeSent'));
      setStep('reset');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to send code');
    } finally {
      setLoading(false);
    }
  };

  const handleReset = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (code.length !== 6) {
      setError(t('validation.invalidCode'));
      return;
    }
    if (!PASSWORD_REGEX.test(newPassword)) {
      setError(t('auth.passwordRequirements'));
      return;
    }

    setLoading(true);
    try {
      await voyaraApi.resetPassword(email, code, newPassword);
      navigate('/voyara/login', { replace: true });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Reset failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="vy-auth-page">
      <div className="vy-auth-card vy-card">
        <h1 className="vy-heading h3">{t('auth.resetPassword')}</h1>

        {error && <div className="vy-auth-error">{error}</div>}
        {message && <div className="vy-auth-success">{message}</div>}

        {step === 'email' ? (
          <form onSubmit={handleSendCode}>
            <p style={{ fontSize: '0.9rem', marginBottom: '1.5rem', color: 'var(--vy-text-dim)' }}>
              Enter your email address and we'll send you a verification code.
            </p>
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.email')}</label>
              <input ref={emailRef} className="vy-input" type="email" value={email}
                onChange={(e) => setEmail(e.target.value)} required autoComplete="email" />
            </div>
            <button className="vy-btn vy-btn-primary" style={{ width: '100%' }} disabled={loading}>
              {loading ? <span className="vy-spinner" /> : t('auth.sendCode')}
            </button>
            <p className="vy-auth-switch" style={{ marginTop: '1rem' }}>
              <Link to="/voyara/login">{t('auth.login')}</Link>
            </p>
          </form>
        ) : (
          <form onSubmit={handleReset}>
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.verificationCode')}</label>
              <input className="vy-input" type="text" value={code}
                onChange={(e) => setCode(e.target.value)} required maxLength={6}
                placeholder="6-digit code" />
            </div>
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.newPassword')}</label>
              <div className="vy-password-wrap">
                <input className="vy-input" type={showPwd ? 'text' : 'password'} value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)} required minLength={8}
                  autoComplete="new-password" />
                <button type="button" className="vy-password-toggle" onClick={() => setShowPwd(!showPwd)}
                  aria-label={showPwd ? 'Hide password' : 'Show password'}>
                  {showPwd ? '🙈' : '👁'}
                </button>
              </div>
              <span className="vy-field-hint" style={{ color: 'var(--vy-text-dim)' }}>
                {t('auth.passwordRequirements')}
              </span>
            </div>
            <button className="vy-btn vy-btn-primary" style={{ width: '100%' }} disabled={loading}>
              {loading ? <span className="vy-spinner" /> : t('auth.resetPassword')}
            </button>
            <p className="vy-auth-switch" style={{ marginTop: '1rem' }}>
              <button type="button" className="vy-link-btn" onClick={() => setStep('email')}>
                &larr; Change email
              </button>
            </p>
          </form>
        )}
      </div>
    </div>
  );
}
