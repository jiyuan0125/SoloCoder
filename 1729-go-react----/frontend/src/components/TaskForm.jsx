import React, { useState } from 'react';

export default function TaskForm({ projectId, project, users, departments, onSubmit, onClose }) {
  const [form, setForm] = useState({
    project_id: projectId,
    title: '',
    assignee_id: '',
    due_date: '',
    priority: 'medium',
  });

  const eligibleUsers = Object.values(users).filter(
    (u) =>
      u.department_id === project.lead_department ||
      (project.participating_depts || []).includes(u.department_id)
  );

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit({
      ...form,
      due_date: new Date(form.due_date).toISOString(),
    });
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>新建子任务</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>任务名称 *</label>
            <input
              type="text"
              required
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>负责人 *</label>
            <select
              required
              value={form.assignee_id}
              onChange={(e) => setForm({ ...form, assignee_id: e.target.value })}
            >
              <option value="">请选择</option>
              {eligibleUsers.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name} ({departments[u.department_id]?.name})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>截止日期 *</label>
            <input
              type="date"
              required
              value={form.due_date}
              onChange={(e) => setForm({ ...form, due_date: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>优先级</label>
            <select
              value={form.priority}
              onChange={(e) => setForm({ ...form, priority: e.target.value })}
            >
              <option value="high">高</option>
              <option value="medium">中</option>
              <option value="low">低</option>
            </select>
          </div>

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              取消
            </button>
            <button type="submit" className="btn btn-primary">
              创建
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
