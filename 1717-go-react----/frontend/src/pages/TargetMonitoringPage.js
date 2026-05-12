import React, { useState, useEffect } from 'react';
import { monitoringAPI } from '../api';

function TargetMonitoringPage() {
  const [monitorings, setMonitorings] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    department_id: '',
    month: '',
    hospitalization_days: '',
    ventilator_days: '',
    vap_cases: '',
    central_line_days: '',
    clabsi_cases: '',
    catheter_days: '',
    cauti_cases: '',
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [monRes, deptsRes] = await Promise.all([
        monitoringAPI.getAll(),
        monitoringAPI.getTargetDepartments(),
      ]);
      setMonitorings(monRes.data);
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
        department_id: parseInt(formData.department_id),
        hospitalization_days: parseInt(formData.hospitalization_days) || 0,
        ventilator_days: parseInt(formData.ventilator_days) || 0,
        vap_cases: parseInt(formData.vap_cases) || 0,
        central_line_days: parseInt(formData.central_line_days) || 0,
        clabsi_cases: parseInt(formData.clabsi_cases) || 0,
        catheter_days: parseInt(formData.catheter_days) || 0,
        cauti_cases: parseInt(formData.cauti_cases) || 0,
      };
      await monitoringAPI.create(data);
      setShowModal(false);
      resetForm();
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '保存失败');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确定删除此监测记录？')) return;
    try {
      await monitoringAPI.delete(id);
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '删除失败');
    }
  };

  const resetForm = () => {
    setFormData({
      department_id: '',
      month: '',
      hospitalization_days: '',
      ventilator_days: '',
      vap_cases: '',
      central_line_days: '',
      clabsi_cases: '',
      catheter_days: '',
      cauti_cases: '',
    });
  };

  const getDepartmentName = (id) => {
    const dept = departments.find(d => d.id === id);
    return dept?.name || '未知';
  };

  return (
    <div>
      <div className="page-header">
        <h2>目标性监测</h2>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + 录入数据
        </button>
      </div>

      {error && <div className="alert-box alert-error">{error}</div>}

      <div className="card">
        <h3>重点科室：ICU、新生儿科、烧伤科、血液科</h3>
        <p style={{ color: '#718096', marginBottom: '16px', fontSize: '14px' }}>
          监测指标：器械使用率（使用天数/住院天数×100%）、相关感染发病率（感染例数/器械使用天数×1000）
        </p>
      </div>

      <div className="card">
        <div className="table-container">
          {loading ? (
            <div className="empty-state">加载中...</div>
          ) : monitorings.length === 0 ? (
            <div className="empty-state">暂无目标性监测数据</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>科室</th>
                  <th>月份</th>
                  <th>住院天数</th>
                  <th>呼吸机使用率</th>
                  <th>VAP发病率</th>
                  <th>中心导管使用率</th>
                  <th>CLABSI发病率</th>
                  <th>导尿管使用率</th>
                  <th>CAUTI发病率</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {monitorings.map((m) => (
                  <tr key={m.id}>
                    <td>{m.department_name || getDepartmentName(m.department_id)}</td>
                    <td>{m.month}</td>
                    <td>{m.hospitalization_days}</td>
                    <td>{m.ventilator_usage_rate}%</td>
                    <td>{m.vap_rate}‰</td>
                    <td>{m.central_line_usage_rate}%</td>
                    <td>{m.clabsi_rate}‰</td>
                    <td>{m.catheter_usage_rate}%</td>
                    <td>{m.cauti_rate}‰</td>
                    <td>
                      <button className="btn btn-danger" onClick={() => handleDelete(m.id)}>
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
              <h3>录入目标性监测数据</h3>
              <button className="modal-close" onClick={() => { setShowModal(false); resetForm(); }}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label>科室 *</label>
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
                <div className="form-group">
                  <label>月份 *</label>
                  <input
                    type="month"
                    value={formData.month}
                    onChange={(e) => setFormData({ ...formData, month: e.target.value })}
                    required
                  />
                </div>
              </div>

              <h4 style={{ margin: '20px 0 12px', color: '#2d3748', fontSize: '14px' }}>呼吸机相关肺炎 (VAP)</h4>
              <div className="form-row-3">
                <div className="form-group">
                  <label>住院天数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.hospitalization_days}
                    onChange={(e) => setFormData({ ...formData, hospitalization_days: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>呼吸机使用天数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.ventilator_days}
                    onChange={(e) => setFormData({ ...formData, ventilator_days: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>VAP例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.vap_cases}
                    onChange={(e) => setFormData({ ...formData, vap_cases: e.target.value })}
                  />
                </div>
              </div>

              <h4 style={{ margin: '20px 0 12px', color: '#2d3748', fontSize: '14px' }}>导管相关血流感染 (CLABSI)</h4>
              <div className="form-row">
                <div className="form-group">
                  <label>中心静脉导管使用天数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.central_line_days}
                    onChange={(e) => setFormData({ ...formData, central_line_days: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>CLABSI例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.clabsi_cases}
                    onChange={(e) => setFormData({ ...formData, clabsi_cases: e.target.value })}
                  />
                </div>
              </div>

              <h4 style={{ margin: '20px 0 12px', color: '#2d3748', fontSize: '14px' }}>导尿管相关尿路感染 (CAUTI)</h4>
              <div className="form-row">
                <div className="form-group">
                  <label>导尿管使用天数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.catheter_days}
                    onChange={(e) => setFormData({ ...formData, catheter_days: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>CAUTI例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.cauti_cases}
                    onChange={(e) => setFormData({ ...formData, cauti_cases: e.target.value })}
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

export default TargetMonitoringPage;
