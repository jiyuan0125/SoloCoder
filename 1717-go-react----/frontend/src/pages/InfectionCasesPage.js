import React, { useState, useEffect } from 'react';
import { infectionAPI, departmentAPI } from '../api';

const INFECTION_SITES = [
  { value: 'respiratory', label: '呼吸道' },
  { value: 'surgical_incision', label: '手术切口' },
  { value: 'urinary_tract', label: '泌尿道' },
  { value: 'bloodstream', label: '血液' },
  { value: 'digestive', label: '消化系统' },
  { value: 'skin_soft_tissue', label: '皮肤软组织' },
];

const INFECTION_TYPES = {
  community: { label: '社区感染', class: 'badge-info' },
  nosocomial: { label: '院内感染', class: 'badge-danger' },
};

function InfectionCasesPage() {
  const [cases, setCases] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({ DepartmentID: '', type: '' });
  const [formData, setFormData] = useState({
    PatientID: '',
    PatientName: '',
    Gender: '',
    Age: '',
    DepartmentID: '',
    AdmissionDate: '',
    InfectionDate: '',
    InfectionSite: '',
    Pathogen: '',
    DrugSensitivity: false,
  });

  useEffect(() => {
    loadData();
  }, [filters]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [casesRes, deptsRes] = await Promise.all([
        infectionAPI.getAll(filters),
        departmentAPI.getAll(),
      ]);
      setCases(casesRes.data);
      setDepartments(deptsRes.data);
    } catch (err) {
      setError(err.response?.data?.error || '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const data = {
        ...formData,
        Age: parseInt(formData.Age) || 0,
        DepartmentID: parseInt(formData.DepartmentID),
        AdmissionDate: new Date(formData.AdmissionDate),
        InfectionDate: new Date(formData.InfectionDate),
      };
      await infectionAPI.create(data);
      setShowModal(false);
      resetForm();
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '保存失败');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确定删除此感染病例？')) return;
    try {
      await infectionAPI.delete(id);
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '删除失败');
    }
  };

  const resetForm = () => {
    setFormData({
      PatientID: '',
      PatientName: '',
      Gender: '',
      Age: '',
      DepartmentID: '',
      AdmissionDate: '',
      InfectionDate: '',
      InfectionSite: '',
      Pathogen: '',
      DrugSensitivity: false,
    });
  };

  const getDepartmentName = (id) => {
    const dept = departments.find(d => d.ID === id);
    return dept?.Name || '未知';
  };

  return (
    <div>
      <div className="page-header">
        <h2>感染病例管理</h2>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + 新增病例
        </button>
      </div>

      {error && <div className="alert-box alert-error">{error}</div>}

      <div className="card">
        <div className="filter-bar">
          <select
            value={filters.DepartmentID}
            onChange={(e) => setFilters({ ...filters, DepartmentID: e.target.value })}
          >
            <option value="">全部科室</option>
            {departments.map(d => (
              <option key={d.ID} value={d.ID}>{d.Name}</option>
            ))}
          </select>
          <select
            value={filters.type}
            onChange={(e) => setFilters({ ...filters, type: e.target.value })}
          >
            <option value="">全部类型</option>
            <option value="community">社区感染</option>
            <option value="nosocomial">院内感染</option>
          </select>
        </div>

        <div className="table-container">
          {loading ? (
            <div className="empty-state">加载中...</div>
          ) : cases.length === 0 ? (
            <div className="empty-state">暂无感染病例数据</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>住院号</th>
                  <th>姓名</th>
                  <th>年龄</th>
                  <th>科室</th>
                  <th>入院日期</th>
                  <th>感染日期</th>
                  <th>感染部位</th>
                  <th>类型</th>
                  <th>病原体</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {cases.map((c) => (
                  <tr key={c.ID}>
                    <td>{c.PatientID}</td>
                    <td>{c.PatientName}</td>
                    <td>{c.Age}</td>
                    <td>{getDepartmentName(c.DepartmentID)}</td>
                    <td>{new Date(c.AdmissionDate).toLocaleDateString()}</td>
                    <td>{new Date(c.InfectionDate).toLocaleDateString()}</td>
                    <td>{INFECTION_SITES.find(s => s.value === c.InfectionSite)?.label || c.InfectionSite}</td>
                    <td>
                      <span className={`badge ${INFECTION_TYPES[c.InfectionType]?.class}`}>
                        {INFECTION_TYPES[c.InfectionType]?.label}
                      </span>
                    </td>
                    <td>{c.Pathogen}</td>
                    <td>
                      <button className="btn btn-danger" onClick={() => handleDelete(c.ID)}>
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
              <h3>新增感染病例</h3>
              <button className="modal-close" onClick={() => { setShowModal(false); resetForm(); }}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label>住院号 *</label>
                  <input
                    type="text"
                    value={formData.PatientID}
                    onChange={(e) => setFormData({ ...formData, PatientID: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>姓名 *</label>
                  <input
                    type="text"
                    value={formData.PatientName}
                    onChange={(e) => setFormData({ ...formData, PatientName: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>性别</label>
                  <select
                    value={formData.Gender}
                    onChange={(e) => setFormData({ ...formData, Gender: e.target.value })}
                  >
                    <option value="">请选择</option>
                    <option value="男">男</option>
                    <option value="女">女</option>
                  </select>
                </div>
                <div className="form-group">
                  <label>年龄</label>
                  <input
                    type="number"
                    value={formData.Age}
                    onChange={(e) => setFormData({ ...formData, Age: e.target.value })}
                  />
                </div>
              </div>
              <div className="form-group">
                <label>入院科室 *</label>
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
                  <label>入院日期 *</label>
                  <input
                    type="date"
                    value={formData.AdmissionDate}
                    onChange={(e) => setFormData({ ...formData, AdmissionDate: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>感染日期 *</label>
                  <input
                    type="date"
                    value={formData.InfectionDate}
                    onChange={(e) => setFormData({ ...formData, InfectionDate: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>感染部位 *</label>
                  <select
                    value={formData.InfectionSite}
                    onChange={(e) => setFormData({ ...formData, InfectionSite: e.target.value })}
                    required
                  >
                    <option value="">请选择</option>
                    {INFECTION_SITES.map(s => (
                      <option key={s.value} value={s.value}>{s.label}</option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label>病原体 *</label>
                  <input
                    type="text"
                    value={formData.Pathogen}
                    onChange={(e) => setFormData({ ...formData, Pathogen: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-group">
                <label>
                  <input
                    type="checkbox"
                    checked={formData.DrugSensitivity}
                    onChange={(e) => setFormData({ ...formData, DrugSensitivity: e.target.checked })}
                  />
                  {' '}已做药敏试验
                </label>
              </div>
              <p style={{ fontSize: '12px', color: '#718096', marginBottom: '16px' }}>
                感染类型由系统自动判断：入院48小时内为社区感染，48小时后为院内感染
              </p>
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

export default InfectionCasesPage;
