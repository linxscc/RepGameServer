import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import './Navigation.css';

const Navigation: React.FC = () => {
  const location = useLocation();

  return (
    <nav className="navigation">
      <div className="nav-container">
        <Link to="/" className="nav-logo">
          Kern Zhou Portfolio
        </Link>
        <div className="nav-links">
          <Link 
            to="/" 
            className={`nav-link ${location.pathname === '/' ? 'active' : ''}`}
          >
            首页
          </Link>
          <Link 
            to="/download" 
            className={`nav-link ${location.pathname === '/download' ? 'active' : ''}`}
          >
            下载客户端
          </Link>
          <Link 
            to="/zsworkexperience" 
            className={`nav-link ${location.pathname === '/zsworkexperience' ? 'active' : ''}`}
          >
            工作经验
          </Link>
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
