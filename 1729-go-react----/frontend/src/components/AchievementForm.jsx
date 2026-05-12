import React, { useState, useEffect } from 'react';
import { format } from 'date-fns';

export default function AchievementForm({ projects, initial, onSubmit, onClose }) {
  const [form, setForm] = useState({
    type: initial?.type || 'paper',
    title: initial?.title || '',
    output_date: initial?.output_date ? format(new Date(initial.output_date), 'yyyy-MM-dd') : '',
    participants: initial?.participants?.join('\n') || '',
    status: initial?.status || 'submitted',
    contributions: initial?.contributions || [],
  });

  const [newContribution, setNewContribution] = useState({
    project_id: '',
    ratio: 50,
  });

  const totalRatio = form.contributions.reduce((sum, c) => sum + c.ratio, 0);

  const addContribution = () => {
    if (!newContribution.project_id || newContribution.ratio <= 0) return;
    if (form.contributions.find((c) => c.project_id === newContribution.project_id)) {
      alert('该项目已添加');
      return;
    }
    setForm({
      ...form,
      contributions: [...form.contributions, { ...newContribution }],
    });
    setNewContribution({ project_id: '', ratio: 50 });
  };

  const removeContribution = (projectId) => {
    setForm({
      ...form,
      contributions: form.contributions.filter((c) => c.project_id !== projectId),
    });
  };

  const updateContributionRatio = (projectId, ratio) => {
    setForm({
      ...form,
      contributions: form.contributions.map((c) =>
        c.project_id === projectId ? { ...c, ratio: parseInt(ratio) || 0 } : c
      ),
    });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (form.contributions.length > 0 && totalRatio !== 100) {
      alert('贡献比例之和必须等于 100%');
      return;
    }
    onSubmit({
      type: form.type,
      title: form.title,
      output_date: new Date(form.output_date).toISOString(),
      participants: form.participants.split('\n').filter((s) => s.trim()),
      status: form.status,
      contributions: form.contributions,
    });
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '700px' }}>
        <h2>{initial ? '编辑成果' : '新建成果'}</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>成果类型 *</label>
            <select
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
            >
              <option value="paper">论文</option>
              <option value="patent">专利</option>
              <option value="software">软著</option>
              <option value="report">技术报告</option>
            </select>
          </div>

          <div className="form-group">
            <label>标题 *</label>
            <input
              type="text"
              required
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>产出日期 *</label>
            <input
              type="date"
              required
              value={form.output_date}
              onChange={(e) => setForm({ ...form, output_date: e.target.value })}
            />
          </div>

          <div className="form-group">
            <label>参与人员（每行一个）</label>
            <textarea
              rows={3}
              value={form.participants}
              onChange={(e) => setForm({ ...form, participants: e.target.value })}
              placeholder="张三&#10;李四"
            />
          </div>

          <div className="form-group">
            <label>状态</label>
            <select
              value={form.status}
              onChange={(e) => setForm({ ...form, status: e.target.value })}
            >
              <option value="submitted">已提交</option>
              <option value="published">已发表</option>
              <option value="authorized">已授权</option>
            </select>
          </div>

          <div className="form-group">
            <label>关联项目及贡献比例（可选）</label>
            {form.contributions.length > 0 && (
              <div style={{ marginBottom: '1rem' }}>
                <table className="data-table" style={{ fontSize: '0.875rem' }}>
                  <thead>
                    <tr>
                      <th>项目</th>
                      <th>比例</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {form.contributions.map((c) => (
                      <tr key={c.project_id}>
                        <td>{projects[c.project_id]?.name || c.project_id}</td>
                        <td>
                          <input
                            type="number"
                            min={1}
                            max={100}
                            value={c.ratio}
                            onChange={(e) => updateContributionRatio(c.project_id, e.target.value)}
                            style={{ width: '80px' }}
                          />
                          %
                        </td>
                        <td>
                          <button
                            type="button"
                            className="btn btn-danger"
                            style={{ padding: '0.25rem 0.5rem', fontSize: '0.75rem' }}
                            onClick={() => removeContribution(c.project_id)}
                          >
                            移除
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <div style={{ marginTop: '0.5rem', fontWeight: totalRatio === 100 ? 'normal' : 'bold', color: totalRatio === 100 ? '#38a169' : '#e53e3e' }}>
                  总比例：{totalRatio}% {totalRatio === 100 ? '✓' : '(必须等于 100%)'}
                </div>
              </div>
            )}

            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'flex-end' }}>
              <select
                value={newContribution.project_id}
                onChange={(e) => setNewContribution({ ...newContribution, project_id: e.target.value })}
                style={{ flex: 1 }}
              >
                <option value="">选择项目</option>
                {Object.values(projects)
                  .filter((p) => !form.contributions.find((c) => c.project_id === p.id))
                  .map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
              </select>
              <input
                type="number"
                min={1}
                max={100}
                value={newContribution.ratio}
                onChange={(e) => setNewContribution({ ...newContribution, ratio: parseInt(e.target.value) || 0 })}
                style={{ width: '100px' }}
              />
              <span style={{ paddingBottom: '0.625rem' }}>%</span>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={addContribution}
                disabled={!newContribution.project_id}
              >
                添加
              </button>
            </div>
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
