import React, { useState, useRef, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import './Navigation.css';

interface DropdownItem {
  label: string;
  to?: string;
  disabled?: boolean;
}

const projectItems: DropdownItem[] = [
  { label: 'RepGameServer', to: '/download' },
  { label: 'Career Journey', to: '/zsworkexperience' },
  { label: 'Portfolio', disabled: true },
  { label: 'Personal Blog', disabled: true },
];

const Navigation: React.FC = () => {
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  useEffect(() => { setOpen(false); }, [location]);

  return (
    <nav className="nav">
      <div className="nav-inner">
        <Link to="/" className="nav-logo">KZ</Link>
        <div className="nav-links">
          <div className="nav-dropdown" ref={ref}>
            <button
              className={`nav-dropdown-trigger ${open ? 'open' : ''}`}
              onClick={() => setOpen(!open)}
              aria-expanded={open}
            >
              Projects
              <span className="nav-dropdown-chevron" aria-hidden="true" />
            </button>
            <div className={`nav-dropdown-panel ${open ? 'open' : ''}`}>
              <div className="nav-dropdown-inner">
                {projectItems.map((item) =>
                  item.disabled ? (
                    <span key={item.label} className="nav-dropdown-item disabled">
                      {item.label}
                      <span className="nav-dropdown-badge">soon</span>
                    </span>
                  ) : (
                    <Link key={item.label} to={item.to!} className="nav-dropdown-item">
                      {item.label}
                    </Link>
                  )
                )}
              </div>
            </div>
          </div>
          <Link
            to="/download"
            className={`nav-link ${location.pathname === '/download' ? 'active' : ''}`}
          >
            Download
          </Link>
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
