import { useState, useRef, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../../contexts/LanguageContext';
import { voyaraApi } from '../../api/client';
import type { ProductCategory, ProductCondition } from '../../api/types';

const CATEGORIES: ProductCategory[] = ['appliance', 'vehicle', 'electronics', 'other'];
const CONDITIONS: ProductCondition[] = ['new', 'like_new', 'used', 'refurbished'];
const MAX_IMAGES = 5;
const MAX_FILE_SIZE = 1 * 1024 * 1024; // 1MB

export default function NewProduct() {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const fileRef = useRef<HTMLInputElement>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [category, setCategory] = useState<ProductCategory>('appliance');
  const [condition, setCondition] = useState<ProductCondition>('used');
  const [error, setError] = useState('');
  const [fileError, setFileError] = useState('');
  const [loading, setLoading] = useState(false);
  const [imageFiles, setImageFiles] = useState<File[]>([]);
  const [previews, setPreviews] = useState<string[]>([]);
  const [dragOver, setDragOver] = useState(false);

  const handleFiles = (fileList: FileList | null) => {
    setFileError('');
    if (!fileList) return;
    const selected = Array.from(fileList);
    const remaining = MAX_IMAGES - imageFiles.length;
    const batch = selected.slice(0, remaining);

    // Validate file sizes
    const oversized = batch.find((f) => f.size > MAX_FILE_SIZE);
    if (oversized) {
      const mb = (oversized.size / (1024 * 1024)).toFixed(1);
      setFileError(`${oversized.name} (${mb}MB) — ${t('form.imagesLimit')}`);
      if (fileRef.current) fileRef.current.value = '';
      return;
    }

    setImageFiles((prev) => [...prev, ...batch]);
    for (const f of batch) {
      const url = URL.createObjectURL(f);
      setPreviews((prev) => [...prev, url]);
    }

    if (fileRef.current) fileRef.current.value = '';
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleFiles(e.dataTransfer.files);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  };

  const removeImage = (idx: number) => {
    URL.revokeObjectURL(previews[idx]);
    setImageFiles((prev) => prev.filter((_, i) => i !== idx));
    setPreviews((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const fd = new FormData();
      fd.append('title', title);
      fd.append('description', description);
      fd.append('price', price);
      fd.append('category', category);
      fd.append('condition', condition);
      for (const f of imageFiles) {
        fd.append('images', f);
      }
      await voyaraApi.createProduct(fd);
      navigate('/voyara/seller/products');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create product');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="vy-section">
      <div className="vy-container" style={{ maxWidth: '640px' }}>
        <h1 className="vy-heading h2">{t('seller.newProduct')}</h1>
        {error && <div className="vy-auth-error">{error}</div>}
        <form onSubmit={handleSubmit} encType="multipart/form-data" style={{ marginTop: '2rem' }}>
          <div className="vy-form-group">
            <label className="vy-label">{t('form.title')}</label>
            <input className="vy-input" value={title} onChange={(e) => setTitle(e.target.value)} required />
          </div>
          <div className="vy-form-group">
            <label className="vy-label">{t('form.description')}</label>
            <textarea className="vy-input" rows={5} value={description} onChange={(e) => setDescription(e.target.value)} required style={{ resize: 'vertical' }} />
          </div>
          <div className="vy-grid-2">
            <div className="vy-form-group">
              <label className="vy-label">Category</label>
              <select className="vy-select" value={category} onChange={(e) => setCategory(e.target.value as ProductCategory)}>
                {CATEGORIES.map((c) => <option key={c} value={c}>{t(`category.${c}`)}</option>)}
              </select>
            </div>
            <div className="vy-form-group">
              <label className="vy-label">{t('filter.condition')}</label>
              <select className="vy-select" value={condition} onChange={(e) => setCondition(e.target.value as ProductCondition)}>
                {CONDITIONS.map((c) => <option key={c} value={c}>{t(`condition.${c}`)}</option>)}
              </select>
            </div>
          </div>
          <div className="vy-form-group">
            <label className="vy-label">{t('form.price')}</label>
            <input className="vy-input" type="number" step="0.01" min="0" value={price} onChange={(e) => setPrice(e.target.value)} required />
          </div>

          {/* Image upload */}
          <div className="vy-form-group">
            <label className="vy-label">{t('form.images')}</label>

            {/* Hidden file input */}
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              multiple
              onChange={(e) => handleFiles(e.target.files)}
              style={{ display: 'none' }}
            />

            {/* Styled upload zone */}
            <div
              onClick={() => !fileError && imageFiles.length < MAX_IMAGES && fileRef.current?.click()}
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              style={{
                border: `2px dashed ${dragOver ? 'var(--vy-amber)' : fileError ? '#e74c3c' : 'var(--vy-border)'}`,
                borderRadius: '10px',
                padding: '2rem 1rem',
                textAlign: 'center',
                cursor: imageFiles.length < MAX_IMAGES ? 'pointer' : 'not-allowed',
                background: dragOver ? 'rgba(255,193,7,0.08)' : 'var(--vy-surface)',
                transition: 'all 0.2s',
              }}
            >
              {imageFiles.length >= MAX_IMAGES ? (
                <span style={{ color: 'var(--vy-text-dim)', fontSize: '0.9rem' }}>
                  Max {MAX_IMAGES} images
                </span>
              ) : (
                <>
                  <div style={{ fontSize: '2rem', marginBottom: '0.5rem', opacity: 0.5 }}>📷</div>
                  <div style={{ fontWeight: 500, color: 'var(--vy-text)' }}>
                    Click or drag images here
                  </div>
                  <div style={{ fontSize: '0.8rem', color: 'var(--vy-text-dim)', marginTop: '0.3rem' }}>
                    JPG, PNG, WEBP &middot; {t('form.imagesLimit')}
                  </div>
                </>
              )}
            </div>

            {fileError && (
              <div style={{ color: '#e74c3c', fontSize: '0.85rem', marginTop: '0.4rem' }}>
                {fileError}
              </div>
            )}

            {previews.length > 0 && (
              <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', marginTop: '12px' }}>
                {previews.map((url, idx) => (
                  <div key={idx} style={{ position: 'relative', width: '80px', height: '80px' }}>
                    <img src={url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', borderRadius: '8px', border: '1px solid var(--vy-border)' }} />
                    <button
                      type="button"
                      onClick={() => removeImage(idx)}
                      style={{
                        position: 'absolute', top: '-6px', right: '-6px',
                        width: '20px', height: '20px', borderRadius: '50%',
                        border: 'none', background: '#e74c3c', color: '#fff',
                        fontSize: '12px', cursor: 'pointer', lineHeight: '20px', textAlign: 'center',
                      }}
                    >&times;</button>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="vy-product-actions" style={{ marginTop: '2rem' }}>
            <button className="vy-btn vy-btn-primary" disabled={loading}>
              {loading ? '...' : t('form.submit')}
            </button>
            <button type="button" className="vy-btn vy-btn-ghost" onClick={() => navigate(-1)}>
              {t('form.cancel')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
