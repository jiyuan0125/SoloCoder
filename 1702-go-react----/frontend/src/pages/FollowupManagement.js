import { useState, useEffect } from 'react';
import { api } from '../api';

function formatDate(dateStr) {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${month}月${day}日`;
}

const FOLLOW_UP_METHODS = [
  { value: '电话', label: '电话' },
  { value: '微信', label: '微信' },
  { value: '短信', label: '短信' },
];

const FEEDBACK_TYPES = [
  { value: '满意', label: '满意' },
  { value: '一般', label: '一般' },
  { value: '不满意', label: '不满意' },
];

function getStatusTag(status) {
  switch (status) {
    case 'pending':
      return { text: '待回访', className: 'tag-yellow' };
    case 'completed':
      return { text: '已完成', className: 'tag-green' };
    default:
      return { text: status, className: 'tag-gray' };
  }
}

function getFeedbackTag(feedback) {
  switch (feedback) {
    case '满意':
      return { text: '满意', className: 'tag-green' };
    case '一般':
      return { text: '一般', className: 'tag-yellow' };
    case '不满意':
      return { text: '不满意', className: 'tag-red' };
    default:
      return null;
  }
}

function getPriorityTag(priority) {
  switch (priority) {
    case 'high':
      return { text: '高', className: 'tag-red' };
    case 'medium':
      return { text: '中', className: 'tag-yellow' };
    case 'low':
      return { text: '低', className: 'tag-green' };
    default:
      return { text: priority, className: 'tag-gray' };
  }
}

export default function FollowUpManagement() {
  const [followUps, setFollowUps] = useState([]);
  const [todos, setTodos] = useState([]);
  const [patients, setPatients] = useState([]);
  const [doctors, setDoctors] = useState([]);
  const [activeTab, setActiveTab] = useState('pending');
  const [selectedIds, setSelectedIds] = useState([]);
  const [showModal, setShowModal] = useState(false);
  const [editingFollowUp, setEditingFollowUp] = useState(null);
  const [formData, setFormData] = useState({
    method: '电话',
    feedback: '满意',
    notes: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const [followUpsData, todosData, patientsData, doctorsData] = await Promise.all([
        api.getFollowUps(),
        api.getTodos(),
        api.getPatients(),
        api.getDoctors(),
      ]);
      setFollowUps(followUpsData || []);
      setTodos(todosData || []);
      setPatients(patientsData);
      setDoctors(doctorsData);
    } catch (err) {
      console.error(err);
    }
  }

  const pendingFollowUps = followUps.filter(f => f.status === 'pending');
  const completedFollowUps = followUps.filter(f => f.status === 'completed');
  const displayedFollowUps = activeTab === 'pending' ? pendingFollowUps : completedFollowUps;

  function handleSelectAll() {
    if (selectedIds.length === displayedFollowUps.length) {
      setSelectedIds([]);
    } else {
      setSelectedIds(displayedFollowUps.map(f => f.id));
    }
  }

  function handleSelectOne(id) {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(i => i !== id));
    } else {
      setSelectedIds([...selectedIds, id]);
    }
  }

  function openModal(followUp) {
    setEditingFollowUp(followUp);
    setFormData({
      method: '电话',
      feedback: '满意',
      notes: '',
    });
    setShowModal(true);
    setError('');
    setSuccess('');
  }

  function closeModal() {
    setShowModal(false);
    setEditingFollowUp(null);
    setSelectedIds([]);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (!editingFollowUp) return;

    setLoading(true);
    setError('');

    try {
      await api.completeFollowUp(
        editingFollowUp.id,
        formData.method,
        formData.feedback,
        formData.notes,
      );
      setSuccess('回访已完成');
      await loadData();
      setTimeout(closeModal, 1500);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleBatchSubmit() {
    if (selectedIds.length === 0) return;

    setLoading(true);
    setError('');

    try {
      const result = await api.batchCompleteFollowUps(
        selectedIds,
        formData.method,
        formData.feedback,
        formData.notes,
      );
      setSuccess(`已批量完成 ${result.completed} 条回访`);
      setSelectedIds([]);
      await loadData();
      setTimeout(() => setSuccess(''), 2000);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function getPatientName(patientId) {
    return patients.find(p => p.id === patientId)?.name || '未知';
  }

  function getDoctorName(doctorId) {
    return doctors.find(d => d.id === doctorId)?.name || '未知';
  }

  return (
    <div className="container">
      <div className="page-header">
        <h1>回访管理</h1>
        <p>管理患者回访任务和待办事项</p>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <div className="card-header">
          <div className="follow-up-tabs">
            <div
              className={`follow-up-tab ${activeTab === 'pending' ? 'active' : ''}`}
              onClick={() => { setActiveTab('pending'); setSelectedIds([]); }}
            >
              待回访 ({pendingFollowUps.length})
            </div>
            <div
              className={`follow-up-tab ${activeTab === 'completed' ? 'active' : ''}`}
              onClick={() => { setActiveTab('completed'); setSelectedIds([]); }}
            >
              已完成 ({completedFollowUps.length})
            </div>
          </div>
        </div>

        {activeTab === 'pending' && selectedIds.length > 0 && (
          <div className="batch-actions">
            <span>已选择 {selectedIds.length} 条</span>
            <select
              className="form-select"
              style={{ width: 'auto', minWidth: '100px' }}
              value={formData.method}
              onChange={e => setFormData({...formData, method: e.target.value})}
            >
              {FOLLOW_UP_METHODS.map(m => (
                <option key={m.value} value={m.value}>
                  {m.label}
                </option>
              ))}
            </select>
            <select
              className="form-select"
              style={{ width: 'auto', minWidth: '100px' }}
              value={formData.feedback}
              onChange={e => setFormData({...formData, feedback: e.target.value})}
            >
              {FEEDBACK_TYPES.map(f => (
                <option key={f.value} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
            <input
              type="text"
              className="form-input"
              style={{ width: '200px' }}
              placeholder="备注（可选）"
              value={formData.notes}
              onChange={e => setFormData({...formData, notes: e.target.value})}
            />
            <button
              className="btn btn-primary"
              onClick={handleBatchSubmit}
              disabled={loading}
            >
              批量完成
            </button>
          </div>
        )}

        {displayedFollowUps.length === 0 ? (
          <div className="empty-state">
            {activeTab === 'pending' ? '暂无待回访任务' : '暂无已完成回访'}
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
              {activeTab === 'pending' && (
                <th style={{ width: '40px' }}>
                  <input
                    type="checkbox"
                    className="checkbox"
                    checked={selectedIds.length === displayedFollowUps.length && displayedFollowUps.length > 0}
                    onChange={handleSelectAll}
                  />
                </th>
              )}
              <th>患者</th>
              <th>医生</th>
              <th>回访日期</th>
              <th>状态</th>
              <th>回访方式</th>
              <th>反馈</th>
              <th>操作</th>
            </tr>
            </thead>
            <tbody>
              {displayedFollowUps.map(f => (
              <tr key={f.id}>
                {activeTab === 'pending' && (
                  <td>
                    <input
                      type="checkbox"
                      className="checkbox"
                      checked={selectedIds.includes(f.id)}
                      onChange={() => handleSelectOne(f.id)}
                    />
                  </td>
                )}
                <td>{getPatientName(f.patient_id)}</td>
                <td>{getDoctorName(f.doctor_id)}</td>
                <td>{formatDate(f.due_date)}</td>
                <td>
                  {(() => {
                    const tag = getStatusTag(f.status);
                    return <span className={`tag ${tag.className}`}>{tag.text}</span>;
                  })()}
                </td>
                <td>{f.method || '-'}</td>
                <td>
                  {(() => {
                    const tag = getFeedbackTag(f.feedback);
                    return tag ? <span className={`tag ${tag.className}`}>{tag.text}</span> : '-';
                  })()}
                </td>
                <td>
                  {f.status === 'pending' && (
                    <button
                      className="btn btn-primary btn-sm"
                      onClick={() => openModal(f)}
                    >
                        完成回访
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {todos.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">待办事项</h2>
            <span className="tag tag-red">{todos.filter(t => !t.done).length} 条未处理</span>
          </div>
          <table className="table">
            <thead>
              <tr>
                <th>优先级</th>
                <th>标题</th>
                <th>患者</th>
                <th>医生</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              {todos.map(t => (
                <tr key={t.id}>
                  <td>
                    {(() => {
                      const tag = getPriorityTag(t.priority);
                      return <span className={`tag ${tag.className}`}>{tag.text}</span>;
                    })()}
                  </td>
                  <td>{t.title}</td>
                  <td>{getPatientName(t.patient_id)}</td>
                  <td>{getDoctorName(t.doctor_id)}</td>
                  <td>{formatDate(t.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && editingFollowUp && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">完成回访</h3>
              <button className="modal-close" onClick={closeModal}>×</button>
            </div>
            <div className="modal-body">
              <div className="mb-2">
                <strong>患者：</strong>{getPatientName(editingFollowUp.patient_id)}
              </div>
              <div className="mb-2">
                <strong>医生：</strong>{getDoctorName(editingFollowUp.doctor_id)}
              </div>
              <div className="mb-3">
                <strong>回访日期：</strong>{formatDate(editingFollowUp.due_date)}
              </div>

              {error && <div className="alert alert-danger">{error}</div>}
              {success && <div className="alert alert-success">{success}</div>}

              <form onSubmit={handleSubmit}>
                <div className="form-group">
                  <label className="form-label">回访方式</label>
                  <select
                    className="form-select"
                    value={formData.method}
                    onChange={e => setFormData({...formData, method: e.target.value})}
                  >
                    {FOLLOW_UP_METHODS.map(m => (
                      <option key={m.value} value={m.value}>
                        {m.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">患者反馈</label>
                  <select
                    className="form-select"
                    value={formData.feedback}
                    onChange={e => setFormData({...formData, feedback: e.target.value})}
                  >
                    {FEEDBACK_TYPES.map(f => (
                      <option key={f.value} value={f.value}>
                        {f.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">备注</label>
                  <textarea
                    className="form-input"
                    rows="3"
                    value={formData.notes}
                    onChange={e => setFormData({...formData, notes: e.target.value})}
                    placeholder="请输入备注信息（可选）"
                  />
                </div>
              </form>
            </div>
            <div className="modal-footer">
              <button className="btn btn-outline" onClick={closeModal}>取消</button>
              <button className="btn btn-primary" onClick={handleSubmit} disabled={loading}>
                {loading ? '提交中...' : '确认完成'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
