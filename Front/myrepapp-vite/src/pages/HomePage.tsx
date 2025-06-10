import React from 'react';
import { Link } from 'react-router-dom';
import './HomePage.css';

const HomePage: React.FC = () => {
  return (
    <div className="app-container">
      <header className="hero">
        <h1 className="hero-title">Hi, I'm Kern Zhou</h1>
        <h2 className="hero-subtitle">Experienced Software Engineer</h2>
        <p className="hero-desc">专注于企业级系统开发、游戏开发和数据库优化，具备多年跨国项目经验。</p>
        <a className="hero-btn" href="#contact">Contact Me</a>
      </header>
      <main>
        <section className="about" id="about">
          <h2>About Me</h2>
          <p>
            I'm Kern Zhou, a software engineer with extensive experience in enterprise systems, game development, and database optimization. 
            I have worked on various projects across different countries including China, Japan, and Australia, 
            with expertise in Python, C#, Unity, and large-scale data processing.
          </p>
        </section>
        <section className="projects" id="projects">
          <h2>Projects</h2>
          <ul>
            <li>
              <strong>RepGameServer</strong> - A scalable backend server built with GoFrame and Nginx.
              <br />
              <Link to="/download" className="download-link">
                📥 下载客户端
              </Link>
            </li>
            <li>
              <strong>My Personal Blog</strong> - A blog platform built with React and Markdown.
            </li>
            <li>
              <strong>Portfolio Website</strong> - This site! Built with React and inspired by Brittany Chiang.
            </li>
            <li>
              <strong>Work Experience</strong> - My professional journey and technical growth.
              <br />
              <Link to="/zsworkexperience" className="download-link">
                💼 查看工作经验
              </Link>
            </li>
          </ul>
        </section>
        <section className="contact" id="contact">
          <h2>Contact</h2>
          <p>Email: <a href="mailto:kern.zhou1995@gmail.com">kern.zhou1995@gmail.com</a></p>
          <p>Phone: +081 80 2484 1107</p>
          <p>Languages: Chinese (Native), Japanese (N1), English (B1)</p>
        </section>
      </main>
      <footer className="footer">
        <p>&copy; {new Date().getFullYear()} Kern Zhou. All rights reserved.</p>
      </footer>
    </div>
  );
}

export default HomePage;
