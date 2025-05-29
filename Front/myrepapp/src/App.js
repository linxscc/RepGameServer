import React, { useState } from 'react';
import axios from 'axios';
import './App.css';

function App() {
  const [name, setName] = useState('');
  const [msg, setMsg] = useState('');

  const handleClick = async () => {
    try {
      const res = await axios.post('http://localhost:8000/normal', { name });
      setMsg(res.data.message || JSON.stringify(res.data));
    } catch (err) {
      setMsg('请求失败: ' + err.message);
    }
  };

  return (
    <div className="app-container">
      <header className="hero">
        <h1 className="hero-title">Hi, I'm ZS</h1>
        <h2 className="hero-subtitle">A Passionate Full Stack Developer</h2>
        <p className="hero-desc">I love building web applications, exploring new technologies, and sharing knowledge with the community.</p>
        <a className="hero-btn" href="#contact">Contact Me</a>
      </header>
      <main>
        <section className="about" id="about">
          <h2>About Me</h2>
          <p>
            I'm ZS, a developer with experience in React, Go, and cloud-native technologies. I enjoy solving real-world problems and creating delightful user experiences.
          </p>
        </section>
        <section className="projects" id="projects">
          <h2>Projects</h2>
          <ul>
            <li>
              <strong>RepGameServer</strong> - A scalable backend server built with GoFrame and Nginx.
            </li>
            <li>
              <strong>My Personal Blog</strong> - A blog platform built with React and Markdown.
            </li>
            <li>
              <strong>Portfolio Website</strong> - This site! Built with React and inspired by Brittany Chiang.
            </li>
          </ul>
        </section>
        <section className="contact" id="contact">
          <h2>Contact</h2>
          <p>Email: <a href="mailto:zspersonaldomain@gmail.com">zspersonaldomain@gmail.com</a></p>
          <p>GitHub: <a href="https://github.com/zspersonaldomain" target="_blank" rel="noopener noreferrer">zspersonaldomain</a></p>
        </section>
        <section className="backend-demo" id="backend-demo" style={{marginTop:40}}>
          <h2>后端交互演示</h2>
          <input
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="请输入名字"
          />
          <button onClick={handleClick} style={{ marginLeft: 10 }}>发送到后端</button>
          <div style={{ marginTop: 20, color: 'green' }}>{msg}</div>
        </section>
      </main>
      <footer className="footer">
        <p>&copy; {new Date().getFullYear()} ZS. All rights reserved.</p>
      </footer>
    </div>
  );
}

export default App;