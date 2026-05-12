import React, { useEffect, useRef, useState } from 'react';
import { getProfileInfo, ProfileInfoResponse } from '@/api/content';
import './HomePage.css';

const HomePage: React.FC = () => {
  const heroRef = useRef<HTMLElement>(null);
  const [profile, setProfile] = useState<ProfileInfoResponse | null>(null);

  useEffect(() => {
    getProfileInfo().then(setProfile).catch(() => {});
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) entry.target.classList.add('revealed');
        });
      },
      { threshold: 0.15, rootMargin: '0px 0px -40px 0px' }
    );
    document.querySelectorAll('.reveal').forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [profile]);

  const nameParts = (profile?.fullName ?? 'Kern Zhou').split(' ');

  return (
    <div className="home">
      {/* ── Hero ── */}
      <section className="hero" ref={heroRef}>
        <div className="hero-bg-glow" />
        <div className="hero-content">
          <div className="hero-rule" />
          <h1 className="hero-name">
            {nameParts.map((part, i) => (
              <span key={i} className="hero-name-line">{part}</span>
            ))}
          </h1>
          <p className="hero-title">{profile?.title ?? 'Software Engineer'}</p>
          <p className="hero-tagline">{profile?.tagline ?? ''}</p>
          <div className="hero-actions">
            <a href="#contact" className="btn btn-gold">Get in touch</a>
          </div>
        </div>
        <div className="hero-scroll-hint">
          <span className="scroll-line" />
        </div>
      </section>

      {/* ── About ── */}
      {profile?.aboutText && (
        <section className="about section reveal" id="about">
          <div className="section-inner">
            <span className="section-num">01</span>
            <div className="section-content">
              <h2 className="section-heading">About</h2>
              <p className="about-text">{profile.aboutText}</p>
              <div className="about-stats">
                <div className="stat">
                  <span className="stat-value">8+</span>
                  <span className="stat-label">Years experience</span>
                </div>
                <div className="stat">
                  <span className="stat-value">3</span>
                  <span className="stat-label">Countries worked</span>
                </div>
                <div className="stat">
                  <span className="stat-value">3</span>
                  <span className="stat-label">Languages</span>
                </div>
              </div>
            </div>
          </div>
        </section>
      )}

      {/* ── Contact ── */}
      <section className="contact section reveal" id="contact">
        <div className="section-inner">
          <span className="section-num">02</span>
          <div className="section-content">
            <h2 className="section-heading">Contact</h2>
            <div className="contact-grid">
              {profile?.email && (
                <a href={`mailto:${profile.email}`} className="contact-item">
                  <span className="contact-label">Email</span>
                  <span className="contact-value">{profile.email}</span>
                </a>
              )}
              {profile?.phone && (
                <div className="contact-item">
                  <span className="contact-label">Phone</span>
                  <span className="contact-value">{profile.phone}</span>
                </div>
              )}
              {profile?.languages && (
                <div className="contact-item">
                  <span className="contact-label">Languages</span>
                  <span className="contact-value">{profile.languages}</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="footer">
        <div className="footer-rule" />
        <p>&copy; {new Date().getFullYear()} {profile?.fullName ?? 'Kern Zhou'}</p>
        <span className="footer-credit">Designed with precision</span>
      </footer>
    </div>
  );
};

export default HomePage;
