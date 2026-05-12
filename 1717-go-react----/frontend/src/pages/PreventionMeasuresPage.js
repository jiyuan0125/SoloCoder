import React, { useState, useEffect } from 'react';
import { measureAPI, infectionAPI, departmentAPI } from '../api';

const MEASURE_TYPES = {
  isolation: { label: '隔离', class: 'badge-warning' },
  hand_hygiene: { label: '手卫生强化', class: 'badge-info' },
  environment_clean: { label: '环境消毒', class: 'badge-success' },
  antibiotic_adjust: { label: '抗生素调整', class: 'badge-danger' },
  equipment_sterile: { label: '器械消毒流程整改', class: 'badge-info' },
};

function PreventionMeasuresPage() {
  const [measures, setMeasures] = useState([]);
  const [infections, setInfections] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [error, setError] = useState('');
  const [recommended, setRecommended] = useState([]);
  const [formData, setFormData] = useState({
    InfectionCaseID: '',
    MeasureType: '',
    DepartmentID: '',
    Executor: '',
    ExecuteDate: '',
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [measuresRes, infectionsRes, deptsRes] = await Promise.all([
        measureAPI.getAll(),
        infectionAPI.getAll(),
        departmentAPI.getAll(),
      ]);
      setMeasures(measuresRes.data);
      setInfections(infectionsRes.data);
      setDepartments(deptsRes.data);
    } catch (err) {
      setError(err.response?.data?.error || '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCaseChange = async (caseId) => {
    setFormData({ ...formData, InfectionCaseID: caseId });
    if (caseId) {
      const infCase = infections.find(c => c.ID === parseInt(caseId));
      if (infCase) {
        try {
          const res = await measureAPI.getRecommended(infCase.InfectionSite);
          setRecommended(res.data);
        } catch (err) {
          console.error('Failed to get recommendations:', err);
        }
      }
    } else {
      setRecommended([]);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const data = {
        ...formData,
        InfectionCaseID: parseInt(formData.InfectionCaseID),
        DepartmentID: parseInt(formData.DepartmentID),
        ExecuteDate: formData.ExecuteDate ? new Date(formData.ExecuteDate) : undefined,
      };
      await measureAPI.create(data);
      setShowModal(false);
      resetForm();
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '保存失败');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确定删除此防控措施？')) return;
    try {
      await measureAPI.delete(id);
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '删除失败');
    }
  };

  const resetForm = () => {
    setFormData({
      InfectionCaseID: '',
      MeasureType: '',
      DepartmentID: '',
      Executor: '',
      ExecuteDate: '',
    });
    setRecommended([]);
  };

  const getCaseInfo = (caseId) => {
    const c = infections.find(c => c.ID === caseId);
    return c ? `${c.PatientID} - ${c.PatientName}` : '未知';
  };

  const getDepartmentName = (id) => {
    const dept = departments.find(d => d.ID === id);
    return dept?.Name || '未知';
  };

  return (
    <div>
      <div className="page-header">
        <h2>防控措施管理</h2>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + 添加措施
        </button>
      </div>

      {error && <div className="alert-box alert-error">{error}</div>}

      <div className="card">
        <div className="table-container">
          {loading ? (
            <div className="empty-state">加载中...</div>
          ) : measures.length === 0 ? (
            <div className="empty-state">暂无防控措施记录</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>感染病例</th>
                  <th>措施类型</th>
                  <th>执行科室</th>
                  <th>执行人</th>
                  <th>执行日期</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {measures.map((m) => (
                  <tr key={m.ID}>
                    <td>{getCaseInfo(m.InfectionCaseID)}</td>
                    <td>
                      <span className={`badge ${MEASURE_TYPES[m.MeasureType]?.class}`}>
                        {MEASURE_TYPES[m.MeasureType]?.label || m.MeasureType}
                      </span>
                    </td>
                    <td>{getDepartmentName(m.DepartmentID)}</td>
                    <td>{m.Executor}</td>
                    <td>{new Date(m.ExecuteDate).toLocaleDateString()}</td>
                    <td>
                      <button className="btn btn-danger" onClick={() => handleDelete(m.ID)}>
                        删除
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {showModal && (
        <div className="modal-overlay">
          <div className="modal">
            <div className="modal-header">
              <h3>添加防控措施</h3>
              <button className="modal-close" onClick={() => { setShowModal(false); resetForm(); }}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>感染病例 *</label>
                <select
                  value={formData.InfectionCaseID}
                  onChange={(e) => handleCaseChange(e.target.value)}
                  required
                >
                  <option value="">请选择病例</option>
                  {infections.map(c => (
                    <option key={c.ID} value={c.ID}>
                      {c.PatientID} - {c.PatientName} ({new Date(c.InfectionDate).toLocaleDateString()})
                    </option>
                  ))}
                </select>
              </div>

              {recommended.length > 0 && (
                <div className="card" style={{ background: '#f7fafc', marginBottom: '16px' }}>
                  <h4 style={{ marginBottom: '12px', color: '#2d3748' }}>推荐防控措施</h4>
                  <ul style={{ listStyle: 'none', padding: 0 }}>
                    {recommended.map((r, i) => (
                      <li key={i} style={{ padding: '6px 0', color: '#4a5568' }}>
                        • {r.Description}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <div className="form-group">
                <label>措施类型 *</label>
                <select
                  value={formData.MeasureType}
                  onChange={(e) => setFormData({ ...formData, MeasureType: e.target.value })}
                  required
                >
                  <option value="">请选择</option>
                  {Object.entries(MEASURE_TYPES).map(([key, val]) => (
                    <option key={key} value={key}>{val.label}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>执行科室 *</label>
                <select
                  value={formData.DepartmentID}
                  onChange={(e) => setFormData({ ...formData, DepartmentID: e.target.value })}
                  required
                >
                  <option value="">请选择科室</option>
                  {departments.map(d => (
                    <option key={d.ID} value={d.ID}>{d.Name}</option>
                  ))}
                </select>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>执行人 *</label>
                  <input
                    type="text"
                    value={formData.Executor}
                    onChange={(e) => setFormData({ ...formData, Executor: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>执行日期</label>
                  <input
                    type="date"
                    value={formData.ExecuteDate}
                    onChange={(e) => setFormData({ ...formData, ExecuteDate: e.target.value })}
                  />
                </div>
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => { setShowModal(false); resetForm(); }}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">保存</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default PreventionMeasuresPage;
