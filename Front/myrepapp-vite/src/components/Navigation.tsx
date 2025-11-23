import React from 'react';
import { Link } from 'react-router-dom';
import './Navigation.css';

const Navigation: React.FC = () => {
  return (
    <nav className="navigation">
      <div className="nav-container">
        <Link to="/" className="nav-logo">
          Kern
        </Link>        
        <div className="nav-links">
          {/* 导航链接 */}
        </div>
      </div>
    </nav>
  );
};

export default Navigation;

