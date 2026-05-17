import { useState } from 'react';
import { useLanguage } from '../contexts/LanguageContext';
import { useAuth } from '../contexts/AuthContext';
import { voyaraApi } from '../api/client';

export default function AccountPage() {
  const { t } = useLanguage();
  const { user } = useAuth();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [changing, setChanging] = useState(false);

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setMessage('');
    if (!oldPassword || !newPassword) return;
    setChanging(true);
    try {
      await voyaraApi.changePassword(oldPassword, newPassword);
      setMessage(t('account.passwordChanged'));
      setOldPassword('');
      setNewPassword('');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to change password');
    } finally {
      setChanging(false);
    }
  };

  return (
    <div className="vy-section">
      <div className="vy-container vy-account-container" style={{ maxWidth: '600px' }}>
        <h1 className="vy-heading h2">{t('account.title')}</h1>

        {/* User Info */}
        <div className="vy-card" style={{ padding: '1.5rem', marginTop: '2rem' }}>
          <h3 style={{ marginBottom: '1rem', fontWeight: 600 }}>{t('account.info')}</h3>
          <div style={{ display: 'grid', gap: '0.75rem' }}>
            <div>
              <span className="vy-label">{t('auth.name')}: </span>
              <span>{user?.name}</span>
            </div>
            <div>
              <span className="vy-label">{t('auth.email')}: </span>
              <span>{user?.email}</span>
            </div>
            <div>
              <span className="vy-label">{t('account.role')}: </span>
              <span style={{ textTransform: 'capitalize' }}>{user?.role}</span>
            </div>
          </div>
        </div>

        {/* Change Password */}
        <div className="vy-card" style={{ padding: '1.5rem', marginTop: '1.5rem' }}>
          <h3 style={{ marginBottom: '1rem', fontWeight: 600 }}>{t('auth.changePassword')}</h3>
          {message && <div className="vy-auth-success">{message}</div>}
          {error && <div className="vy-auth-error">{error}</div>}
          <form onSubmit={handleChangePassword}>
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.oldPassword')}</label>
              <input className="vy-input" type="password" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} required />
            </div>
            <div className="vy-form-group">
              <label className="vy-label">{t('auth.newPassword')}</label>
              <input className="vy-input" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required minLength={8} />
            </div>
            <button className="vy-btn vy-btn-primary" disabled={changing}>
              {changing ? '...' : t('auth.changePassword')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
