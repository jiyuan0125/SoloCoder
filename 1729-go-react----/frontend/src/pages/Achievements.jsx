import React, { useState, useEffect } from 'react';
import { format } from 'date-fns';
import { achievementsAPI, projectsAPI } from '../services/api';
import AchievementForm from '../components/AchievementForm';

const typeLabels = {
  paper: '论文',
  patent: '专利',
  software: '软著',
  report: '技术报告',
};

const statusLabels = {
  submitted: '已提交',
  published: '已发表',
  authorized: '已授权',
};

export default function Achievements() {
  const [achievements, setAchievements] = useState([]);
  const [projects, setProjects] = useState({});
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [achRes, projRes] = await Promise.all([
        achievementsAPI.getAll(),
        projectsAPI.getAll(),
      ]);
      setAchievements(achRes.data);
      const projMap = {};
      projRes.data.forEach((p) => (projMap[p.id] = p));
      setProjects(projMap);
    } catch (err) {
      alert('加载数据失败');
    }
  };

  const handleCreate = async (data) => {
    try {
      if (editing) {
        await achievementsAPI.update(editing.id, data);
      } else {
        await achievementsAPI.create(data);
      }
      setShowForm(false);
      setEditing(null);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '保存失败');
    }
  };

  const handleExport = async () => {
    try {
      const res = await achievementsAPI.export();
      const dataStr = JSON.stringify(res.data, null, 2);
      const blob = new Blob([dataStr], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `achievements_export_${format(new Date(), 'yyyyMMdd')}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      alert('导出失败');
    }
  };

  return (
    <div>
      <div className="page-header">
        <h1>成果管理</h1>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className="btn btn-secondary" onClick={handleExport}>
            导出
          </button>
          <button className="btn btn-primary" onClick={() => { setEditing(null); setShowForm(true); }}>
            新建成果
          </button>
        </div>
      </div>

      {achievements.length === 0 ? (
        <p style={{ textAlign: 'center', padding: '3rem', color: '#718096' }}>
          暂无成果
        </p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>类型</th>
              <th>标题</th>
              <th>产出日期</th>
              <th>参与人员</th>
              <th>关联项目</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {achievements.map((ach) => (
              <tr key={ach.id}>
                <td>{typeLabels[ach.type]}</td>
                <td>{ach.title}</td>
                <td>{format(new Date(ach.output_date), 'yyyy-MM-dd')}</td>
                <td>{ach.participants?.join('、') || '-'}</td>
                <td>
                  {ach.contributions
                    ?.map((c) => `${projects[c.project_id]?.name || c.project_id} (${c.ratio}%)`)
                    .join('、') || '-'}
                </td>
                <td>
                  <span className="badge badge-in_progress">
                    {statusLabels[ach.status]}
                  </span>
                </td>
                <td>
                  <button
                    className="btn btn-primary"
                    style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem' }}
                    onClick={() => { setEditing(ach); setShowForm(true); }}
                  >
                    编辑
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showForm && (
        <AchievementForm
          projects={projects}
          initial={editing}
          onSubmit={handleCreate}
          onClose={() => { setShowForm(false); setEditing(null); }}
        />
      )}
    </div>
  );
}
