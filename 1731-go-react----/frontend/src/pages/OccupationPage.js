import React, { useState, useEffect } from 'react';
import { occupationApi } from '../api';

const INDUSTRY_OPTIONS = ['制造', '服务', 'IT', '建筑', '医疗', '其他'];
const LEVEL_OPTIONS = ['初级', '中级', '高级', '技师', '高级技师'];
const SUBJECT_OPTIONS = ['理论知识', '实操技能', '综合评审'];

function OccupationPage() {
  const [occupations, setOccupations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    code: '',
    industry: '制造',
    levels: [],
  });
  const [error, setError] = useState('');

  useEffect(() => {
    loadOccupations();
  }, []);

  const loadOccupations = async () => {
    try {
      const res = await occupationApi.getAll();
      setOccupations(res.data);
    } catch (err) {
      console.error('加载失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const openCreateModal = () => {
    setEditingId(null);
    setFormData({ name: '', code: '', industry: '制造', levels: [] });
    setError('');
    setShowModal(true);
  };

  const openEditModal = (occ) => {
    setEditingId(occ.id);
    setFormData({
      name: occ.name,
      code: occ.code,
      industry: occ.industry,
      levels: [...occ.levels],
    });
    setError('');
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (!formData.name || !formData.code) {
      setError('请填写职业名称和代码');
      return;
    }
    if (formData.levels.length === 0) {
      setError('请至少添加一个技能等级');
      return;
    }

    try {
      if (editingId) {
        await occupationApi.update(editingId, formData);
      } else {
        await occupationApi.create(formData);
      }
      setShowModal(false);
      loadOccupations();
    } catch (err) {
      setError(err.response?.data?.error || '操作失败');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确定要删除该职业吗？')) return;
    try {
      await occupationApi.delete(id);
      loadOccupations();
    } catch (err) {
      alert('删除失败: ' + (err.response?.data?.error || err.message));
    }
  };

  const addLevel = () => {
    const newLevels = [...formData.levels, { level: '初级', subjects: ['理论知识', '实操技能'] }];
    setFormData({ ...formData, levels: newLevels });
  };

  const removeLevel = (index) => {
    const newLevels = formData.levels.filter((_, i) => i !== index);
    setFormData({ ...formData, levels: newLevels });
  };

  const updateLevel = (index, field, value) => {
    const newLevels = [...formData.levels];
    newLevels[index][field] = value;
    setFormData({ ...formData, levels: newLevels });
  };

  const toggleSubject = (levelIndex, subject) => {
    const newLevels = [...formData.levels];
    const subjects = newLevels[levelIndex].subjects;
    if (subjects.includes(subject)) {
      newLevels[levelIndex].subjects = subjects.filter(s => s !== subject);
    } else {
      newLevels[levelIndex].subjects = [...subjects, subject];
    }
    setFormData({ ...formData, levels: newLevels });
  };

  if (loading) {
    return <div className="page-container">加载中...</div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h2 className="page-title">职业技能标准管理</h2>
        <button className="btn btn-primary" onClick={openCreateModal}>
          + 新建职业
        </button>
      </div>

      {occupations.length === 0 ? (
        <p style={{ color: '#666', textAlign: 'center', padding: 40 }}>
          暂无职业数据，点击上方按钮添加
        </p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>职业代码</th>
              <th>职业名称</th>
              <th>行业分类</th>
              <th>技能等级</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {occupations.map((occ) => (
              <tr key={occ.id}>
                <td>{occ.code}</td>
                <td>{occ.name}</td>
                <td>{occ.industry}</td>
                <td>
                  {occ.levels.map(l => l.level).join('、')}
                </td>
                <td>
                  <div className="action-buttons">
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => openEditModal(occ)}
                    >
                      编辑
                    </button>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => handleDelete(occ.id)}
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">{editingId ? '编辑职业' : '新建职业'}</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>&times;</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label className="form-label">职业代码 *</label>
                <input
                  className="form-input"
                  placeholder="例如：001"
                  value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label">职业名称 *</label>
                <input
                  className="form-input"
                  placeholder="例如：计算机程序设计员"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label">行业分类</label>
                <select
                  className="form-select"
                  value={formData.industry}
                  onChange={(e) => setFormData({ ...formData, industry: e.target.value })}
                >
                  {INDUSTRY_OPTIONS.map(ind => (
                    <option key={ind} value={ind}>{ind}</option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">技能等级配置</label>
                {formData.levels.map((levelConfig, index) => (
                  <div
                    key={index}
                    style={{
                      padding: 12,
                      marginBottom: 8,
                      border: '1px solid #ddd',
                      borderRadius: 4,
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                      <select
                        className="form-select"
                        style={{ width: 'auto', minWidth: 120 }}
                        value={levelConfig.level}
                        onChange={(e) => updateLevel(index, 'level', e.target.value)}
                      >
                        {LEVEL_OPTIONS.map(level => (
                          <option key={level} value={level}>{level}</option>
                        ))}
                      </select>
                      <button
                        type="button"
                        className="btn btn-danger btn-sm"
                        onClick={() => removeLevel(index)}
                      >
                        移除
                      </button>
                    </div>
                    <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
                      {SUBJECT_OPTIONS.map(subject => (
                        <label key={subject} style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                          <input
                            type="checkbox"
                            checked={levelConfig.subjects.includes(subject)}
                            onChange={() => toggleSubject(index, subject)}
                          />
                          {subject}
                        </label>
                      ))}
                    </div>
                  </div>
                ))}
                <button
                  type="button"
                  className="btn btn-secondary btn-sm"
                  onClick={addLevel}
                >
                  + 添加等级
                </button>
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowModal(false)}
                >
                  取消
                </button>
                <button type="submit" className="btn btn-primary">
                  {editingId ? '保存' : '创建'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default OccupationPage;
