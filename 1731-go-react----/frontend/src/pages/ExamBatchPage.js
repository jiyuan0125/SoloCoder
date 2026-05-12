import React, { useState, useEffect } from 'react';
import { batchApi, occupationApi } from '../api';

const LEVEL_OPTIONS = ['初级', '中级', '高级', '技师', '高级技师'];

function ExamBatchPage() {
  const [batches, setBatches] = useState([]);
  const [occupations, setOccupations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    occupationId: '',
    level: '初级',
    examDate: '',
    examRoom: '',
  });
  const [error, setError] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [batchesRes, occsRes] = await Promise.all([
        batchApi.getAll(),
        occupationApi.getAll(),
      ]);
      setBatches(batchesRes.data);
      setOccupations(occsRes.data);
    } catch (err) {
      console.error('加载失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateBatch = async (e) => {
    e.preventDefault();
    setError('');

    try {
      await batchApi.create({
        ...formData,
        examDate: new Date(formData.examDate).toISOString(),
      });
      setShowCreateModal(false);
      setFormData({ name: '', occupationId: '', level: '初级', examDate: '', examRoom: '' });
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '创建失败');
    }
  };

  const handleNextStatus = async (batch) => {
    if (!confirm(`确定要将批次状态从"${batch.status}"流转到下一步吗？`)) return;
    try {
      await batchApi.nextStatus(batch.id);
      loadData();
    } catch (err) {
      alert('状态更新失败: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleDeleteBatch = async (batch) => {
    if (!confirm('确定要删除该批次吗？已有考生报名的批次不能删除。')) return;
    try {
      await batchApi.delete(batch.id);
      loadData();
    } catch (err) {
      alert('删除失败: ' + (err.response?.data?.error || err.message));
    }
  };

  const canAdvance = (status) => status === '未开始' || status === '进行中';

  if (loading) {
    return <div className="page-container">加载中...</div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h2 className="page-title">考试批次管理</h2>
        <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
          + 新建批次
        </button>
      </div>

      {batches.length === 0 ? (
        <p style={{ color: '#666', textAlign: 'center', padding: 40 }}>
          暂无考试批次
        </p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>批次名称</th>
              <th>职业</th>
              <th>等级</th>
              <th>考试日期</th>
              <th>考场</th>
              <th>报名人数</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {batches.map((batch) => (
              <tr key={batch.id}>
                <td>{batch.name}</td>
                <td>{batch.occupationName}</td>
                <td>{batch.level}</td>
                <td>{new Date(batch.examDate).toLocaleDateString()}</td>
                <td>{batch.examRoom}</td>
                <td>{batch.candidates.length}</td>
                <td>
                  <span className={`status-badge ${
                    batch.status === '已结束' ? 'status-valid' : 
                    batch.status === '进行中' ? 'status-expired' : ''
                  }`} style={{
                    backgroundColor: batch.status === '未开始' ? '#e3f2fd' : undefined,
                    color: batch.status === '未开始' ? '#1565c0' : undefined,
                  }}>
                    {batch.status}
                  </span>
                </td>
                <td>
                  <div className="action-buttons">
                    {canAdvance(batch.status) && (
                      <button
                        className="btn btn-success btn-sm"
                        onClick={() => handleNextStatus(batch)}
                      >
                        下一步
                      </button>
                    )}
                    {batch.status === '未开始' && (
                      <button
                        className="btn btn-danger btn-sm"
                        onClick={() => handleDeleteBatch(batch)}
                      >
                        删除
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">新建考试批次</h3>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>&times;</button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <form onSubmit={handleCreateBatch}>
              <div className="form-group">
                <label className="form-label">批次名称 *</label>
                <input
                  className="form-input"
                  placeholder="例如：2026年第一批计算机程序员考试"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label">选择职业 *</label>
                <select
                  className="form-select"
                  value={formData.occupationId}
                  onChange={(e) => setFormData({ ...formData, occupationId: e.target.value })}
                >
                  <option value="">请选择职业</option>
                  {occupations.map((occ) => (
                    <option key={occ.id} value={occ.id}>{occ.name} ({occ.code})</option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">技能等级 *</label>
                <select
                  className="form-select"
                  value={formData.level}
                  onChange={(e) => setFormData({ ...formData, level: e.target.value })}
                >
                  {LEVEL_OPTIONS.map(level => (
                    <option key={level} value={level}>{level}</option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">考试日期 *</label>
                <input
                  type="datetime-local"
                  className="form-input"
                  value={formData.examDate}
                  onChange={(e) => setFormData({ ...formData, examDate: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label">考场 *</label>
                <input
                  className="form-input"
                  placeholder="例如：第一考场 A301"
                  value={formData.examRoom}
                  onChange={(e) => setFormData({ ...formData, examRoom: e.target.value })}
                />
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowCreateModal(false)}
                >
                  取消
                </button>
                <button type="submit" className="btn btn-primary">
                  创建批次
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default ExamBatchPage;
