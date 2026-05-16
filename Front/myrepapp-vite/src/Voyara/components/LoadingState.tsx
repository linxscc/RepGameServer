export function LoadingState({ message = 'Loading...' }: { message?: string }) {
  return (
    <div className="vy-state-page">
      <span className="vy-spinner" style={{ width: '2rem', height: '2rem' }} />
      <p className="vy-state-message">{message}</p>
    </div>
  );
}

export function EmptyState({ icon = '📦', message, action }: { icon?: string; message: string; action?: { label: string; href?: string; onClick?: () => void } }) {
  return (
    <div className="vy-state-page">
      <div className="vy-state-icon">{icon}</div>
      <p className="vy-state-message">{message}</p>
      {action && (
        action.href
          ? <a href={action.href} className="vy-btn vy-btn-primary">{action.label}</a>
          : <button className="vy-btn vy-btn-primary" onClick={action.onClick}>{action.label}</button>
      )}
    </div>
  );
}

export function ErrorState({ message, onRetry, onBack }: { message: string; onRetry?: () => void; onBack?: () => void }) {
  return (
    <div className="vy-state-page">
      <div className="vy-state-icon">⚠️</div>
      <p className="vy-state-message">{message}</p>
      <div style={{ display: 'flex', gap: '0.75rem', marginTop: '0.5rem' }}>
        {onRetry && <button className="vy-btn vy-btn-primary" onClick={onRetry}>Retry</button>}
        {onBack && <button className="vy-btn vy-btn-secondary" onClick={onBack}>Go Back</button>}
      </div>
    </div>
  );
}
