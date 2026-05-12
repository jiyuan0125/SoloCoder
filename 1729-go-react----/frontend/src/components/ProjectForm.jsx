import React, { useState } from 'react';
import { format } from 'date-fns';

export default function ProjectForm({ departments, users, onSubmit, onClose, initial }) {
  const [form, setForm] = useState({
    name: initial?.name || '',
    lead_department: initial?.lead_department || '',
    participating_depts: initial?.participating_depts || [],
    leader_id: initial?.leader_id || '',
    start_date: initial?.start_date ? format(new Date(initial.start_date), 'yyyy-MM-dd') : '',
    end_date: initial?.end_date ? format(new Date(initial.end_date), 'yyyy-MM-dd') : '',
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    const data = {
      ...form,
      start_date: new Date(form.start_date).toISOString(),
      end_date: new Date(form.end_date).toISOString(),
    };
    onSubmit(data);
  };

  const handleDeptToggle = (deptId) => {
    const current = form.participating_depts || [];
    setForm({
      ...form,
      participating_depts: current.includes(deptId)
        ? current.filter((d) => d !== deptId)
        : [...current, deptId],
    });
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>{initial ? '编辑项目' : '新建项目'}</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>项目名称 *</label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>牵头部门 *</label>
            <select
              required
              value={form.lead_department}
              onChange={(e) => setForm({ ...form, lead_department: e.target.value })}
            >
              <option value="">请选择</option>
              {Object.values(departments).map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>参与部门</label>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
              {Object.values(departments).map((d) => (
                <label
                  key={d.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.25rem',
                    padding: '0.5rem',
                    background: '#f7fafc',
                    borderRadius: '4px',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={form.participating_depts.includes(d.id)}
                    onChange={() => handleDeptToggle(d.id)}
                  />
                  {d.name}
                </label>
              ))}
            </div>
          </div>

          <div className="form-group">
            <label>负责人 *</label>
            <select
              required
              value={form.leader_id}
              onChange={(e) => setForm({ ...form, leader_id: e.target.value })}
            >
              <option value="">请选择</option>
              {Object.values(users).map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>开始日期 *</label>
            <input
              type="date"
              required
              value={form.start_date}
              onChange={(e) => setForm({ ...form, start_date: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>截止日期 *</label>
            <input
              type="date"
              required
              value={form.end_date}
              onChange={(e) => setForm({ ...form, end_date: e.target.value })}
            />
          </div>

          <div className="form-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={onClose}
            >
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
