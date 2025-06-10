import React from 'react';
import { DownloadItem } from '@/types';
import './DownloadPage.css';

const DownloadPage: React.FC = () => {
  const downloadItems: DownloadItem[] = [
    {
      name: 'RepGame 客户端 (Windows)',
      version: 'v1.0.0',
      size: '15 MB',
      description: 'Windows 桌面客户端，支持 Windows 10/11',
      downloadUrl: 'https://myrepgamebucket.s3.ap-southeast-2.amazonaws.com/repgame-downloads/SmallCardGame.7z',
      icon: '🖥️'
    },
    {
      name: 'RepGame 客户端 (Mac)',
      version: 'v1.0.0',
      size: '15 MB',
      description: 'macOS 桌面客户端，支持 macOS 10.15+（即将推出）',
      downloadUrl: '/download/SmallCardGame-Mac.7z',
      icon: '🍎'
    },
    {
      name: 'RepGame 客户端 (Linux)',
      version: 'v1.0.0',
      size: '15 MB',
      description: 'Linux 桌面客户端，支持主流发行版（即将推出）',
      downloadUrl: '/download/SmallCardGame-Linux.7z',
      icon: '🐧'
    }
  ];

  const handleDownload = (url: string, filename: string): void => {
    // 创建隐藏的下载链接
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="download-page">
      <div className="download-container">
        <header className="download-header">
          <h1 className="download-title">下载 RepGame 客户端</h1>
          <p className="download-subtitle">
            选择适合你操作系统的客户端版本，开始游戏之旅！
          </p>
        </header>

        <div className="download-grid">
          {downloadItems.map((item, index) => (
            <div key={index} className="download-card">
              <div className="card-icon">{item.icon}</div>
              <h3 className="card-title">{item.name}</h3>
              <div className="card-meta">
                <span className="version">版本: {item.version}</span>
                <span className="size">大小: {item.size}</span>
              </div>
              <p className="card-description">{item.description}</p>
              <button 
                className="download-btn"
                onClick={() => handleDownload(item.downloadUrl, item.name)}
              >
                <span className="btn-icon">⬇️</span>
                立即下载
              </button>
            </div>
          ))}
        </div>

        <div className="system-requirements">
          <h2>系统要求</h2>
          <div className="requirements-grid">
            <div className="req-section">
              <h3>Windows</h3>
              <ul>
                <li>Windows 10 或更高版本</li>
                <li>2GB RAM</li>
                <li>500MB 可用磁盘空间</li>
                <li>网络连接</li>
              </ul>
            </div>
            <div className="req-section">
              <h3>macOS</h3>
              <ul>
                <li>macOS 10.15 或更高版本</li>
                <li>2GB RAM</li>
                <li>500MB 可用磁盘空间</li>
                <li>网络连接</li>
              </ul>
            </div>
            <div className="req-section">
              <h3>Linux</h3>
              <ul>
                <li>Ubuntu 18.04+ / CentOS 7+</li>
                <li>2GB RAM</li>
                <li>500MB 可用磁盘空间</li>
                <li>网络连接</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="help-section">
          <h2>需要帮助？</h2>
          <p>
            如果您在下载或安装过程中遇到问题，请查看我们的 
            <a href="/help" className="help-link">帮助文档</a> 
            或联系客服支持。
          </p>
        </div>
      </div>
    </div>
  );
};

export default DownloadPage;
