import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useLanguage } from '../contexts/LanguageContext';
import { voyaraApi } from '../api/client';

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/;

export default function Register() {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [codeSent, setCodeSent] = useState(false);
  const [codeSending, setCodeSending] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

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
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to send code');
    } finally {
      setCodeSending(false);
    }
  };

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
      localStorage.setItem('voyara_token', res.token);
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
            <input className="vy-input" type="text" value={name}
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
            <input className={`vy-input${password && !PASSWORD_REGEX.test(password) ? ' vy-input-error' : ''}`}
              type="password" value={password}
              onChange={(e) => setPassword(e.target.value)} required minLength={8} />
            {password && !PASSWORD_REGEX.test(password) && (
              <span className="vy-field-hint">{t('auth.passwordRequirements')}</span>
            )}
          </div>

          {!codeSent ? (
            <button type="button" className="vy-btn vy-btn-outline" style={{ width: '100%' }}
              onClick={handleSendCode} disabled={codeSending || !email || !EMAIL_REGEX.test(email)}>
              {codeSending ? '...' : t('auth.sendCode')}
            </button>
          ) : (
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.verificationCode')}</label>
              <input className="vy-input" type="text" value={code}
                onChange={(e) => setCode(e.target.value)} required maxLength={6}
                placeholder="6-digit code" />
            </div>
          )}

          <button className="vy-btn vy-btn-primary" style={{ width: '100%', marginTop: '1rem' }}
            disabled={loading || !codeSent}>
            {loading ? '...' : t('auth.register')}
          </button>
        </form>
        <p className="vy-auth-switch">
          {t('auth.hasAccount')} <Link to="/voyara/login">{t('auth.login')}</Link>
        </p>
      </div>
    </div>
  );
}
