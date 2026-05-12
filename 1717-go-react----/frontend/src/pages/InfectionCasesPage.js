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
  const [filters, setFilters] = useState({ department_id: '', type: '' });
  const [formData, setFormData] = useState({
    patient_id: '',
    patient_name: '',
    gender: '',
    age: '',
    department_id: '',
    admission_date: '',
    infection_date: '',
    infection_site: '',
    pathogen: '',
    drug_sensitivity: false,
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
        age: parseInt(formData.age) || 0,
        department_id: parseInt(formData.department_id),
        admission_date: new Date(formData.admission_date),
        infection_date: new Date(formData.infection_date),
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
      patient_id: '',
      patient_name: '',
      gender: '',
      age: '',
      department_id: '',
      admission_date: '',
      infection_date: '',
      infection_site: '',
      pathogen: '',
      drug_sensitivity: false,
    });
  };

  const getDepartmentName = (id) => {
    const dept = departments.find(d => d.id === id);
    return dept?.name || '未知';
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
            value={filters.department_id}
            onChange={(e) => setFilters({ ...filters, department_id: e.target.value })}
          >
            <option value="">全部科室</option>
            {departments.map(d => (
              <option key={d.id} value={d.id}>{d.name}</option>
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
                  <tr key={c.id}>
                    <td>{c.patient_id}</td>
                    <td>{c.patient_name}</td>
                    <td>{c.age}</td>
                    <td>{getDepartmentName(c.department_id)}</td>
                    <td>{new Date(c.admission_date).toLocaleDateString()}</td>
                    <td>{new Date(c.infection_date).toLocaleDateString()}</td>
                    <td>{INFECTION_SITES.find(s => s.value === c.infection_site)?.label || c.infection_site}</td>
                    <td>
                      <span className={`badge ${INFECTION_TYPES[c.infection_type]?.class}`}>
                        {INFECTION_TYPES[c.infection_type]?.label}
                      </span>
                    </td>
                    <td>{c.pathogen}</td>
                    <td>
                      <button className="btn btn-danger" onClick={() => handleDelete(c.id)}>
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
                    value={formData.patient_id}
                    onChange={(e) => setFormData({ ...formData, patient_id: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>姓名 *</label>
                  <input
                    type="text"
                    value={formData.patient_name}
                    onChange={(e) => setFormData({ ...formData, patient_name: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>性别</label>
                  <select
                    value={formData.gender}
                    onChange={(e) => setFormData({ ...formData, gender: e.target.value })}
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
                    value={formData.age}
                    onChange={(e) => setFormData({ ...formData, age: e.target.value })}
                  />
                </div>
              </div>
              <div className="form-group">
                <label>入院科室 *</label>
                <select
                  value={formData.department_id}
                  onChange={(e) => setFormData({ ...formData, department_id: e.target.value })}
                  required
                >
                  <option value="">请选择科室</option>
                  {departments.map(d => (
                    <option key={d.id} value={d.id}>{d.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>入院日期 *</label>
                  <input
                    type="date"
                    value={formData.admission_date}
                    onChange={(e) => setFormData({ ...formData, admission_date: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>感染日期 *</label>
                  <input
                    type="date"
                    value={formData.infection_date}
                    onChange={(e) => setFormData({ ...formData, infection_date: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>感染部位 *</label>
                  <select
                    value={formData.infection_site}
                    onChange={(e) => setFormData({ ...formData, infection_site: e.target.value })}
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
                    value={formData.pathogen}
                    onChange={(e) => setFormData({ ...formData, pathogen: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-group">
                <label>
                  <input
                    type="checkbox"
                    checked={formData.drug_sensitivity}
                    onChange={(e) => setFormData({ ...formData, drug_sensitivity: e.target.checked })}
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
