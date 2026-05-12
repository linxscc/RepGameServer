import React, { useEffect, useState } from 'react';
import { getDownloadItems, getSystemRequirements, DownloadItemResponse, SystemRequirementResponse } from '@/api/content';
import './DownloadPage.css';

const DownloadPage: React.FC = () => {
  const [downloadItems, setDownloadItems] = useState<DownloadItemResponse[]>([]);
  const [requirements, setRequirements] = useState<SystemRequirementResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const [items, reqs] = await Promise.all([
          getDownloadItems(),
          getSystemRequirements(),
        ]);
        setDownloadItems(items);
        setRequirements(reqs);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const handleDownload = (url: string, filename: string): void => {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loading) {
    return (
      <div className="download-page">
        <div className="download-container">
          <div className="download-header">
            <h1 className="download-title">Loading...</h1>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="download-page">
        <div className="download-container">
          <div className="download-header">
            <h1 className="download-title">Error</h1>
            <p className="download-subtitle">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="download-page">
      <div className="download-container">
        <header className="download-header">
          <h1 className="download-title">Download RepGame Client</h1>
          <p className="download-subtitle">
            Choose your platform and start the journey.
          </p>
        </header>

        <div className="download-grid">
          {downloadItems.map((item) => (
            <div key={item.id} className="download-card">
              <div className="card-icon">{item.icon}</div>
              <h3 className="card-title">{item.name}</h3>
              <div className="card-meta">
                <span className="version">Version: {item.version}</span>
                <span className="size">Size: {item.size}</span>
              </div>
              <p className="card-description">{item.description}</p>
              <button
                className="download-btn"
                onClick={() => handleDownload(item.downloadUrl, item.name)}
              >
                <span className="btn-icon">&darr;</span>
                Download
              </button>
            </div>
          ))}
        </div>

        {requirements.length > 0 && (
          <div className="system-requirements">
            <h2>System Requirements</h2>
            <div className="requirements-grid">
              {requirements.map((req) => (
                <div key={req.id} className="req-section">
                  <h3>{req.osLabel}</h3>
                  <ul>
                    {req.requirements.map((r, i) => (
                      <li key={i}>{r}</li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="help-section">
          <h2>Need Help?</h2>
          <p>
            If you encounter any issues during download or installation, please
            check our help documentation or contact support.
          </p>
        </div>
      </div>
    </div>
  );
};

export default DownloadPage;
