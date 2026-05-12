import React, { useState } from 'react';

export default function DataForm({ projectId, onSubmit, onClose }) {
  const [form, setForm] = useState({
    name: '',
    description: '',
    format: 'csv',
    file_size: 1024,
    project_id: projectId,
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit(form);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>上传数据元数据</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>数据名称 *</label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>描述</label>
            <textarea
              rows={3}
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>格式 *</label>
            <select
              required
              value={form.format}
              onChange={(e) => setForm({ ...form, format: e.target.value })}
            >
              <option value="csv">CSV</option>
              <option value="json">JSON</option>
              <option value="excel">Excel</option>
              <option value="image">图片</option>
              <option value="other">其他</option>
            </select>
          </div>

          <div className="form-group">
            <label>文件大小（字节）*</label>
            <input
              type="number"
              required
              min={1}
              value={form.file_size}
              onChange={(e) => setForm({ ...form, file_size: parseInt(e.target.value) })}
            />
          </div>

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              取消
            </button>
            <button type="submit" className="btn btn-primary">
              保存
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
