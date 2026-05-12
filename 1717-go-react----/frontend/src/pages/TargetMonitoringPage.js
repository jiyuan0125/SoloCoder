import React, { useState, useEffect } from 'react';
import { monitoringAPI } from '../api';

function TargetMonitoringPage() {
  const [monitorings, setMonitorings] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    DepartmentID: '',
    Month: '',
    HospitalizationDays: '',
    VentilatorDays: '',
    VAPCases: '',
    CentralLineDays: '',
    CLABSICases: '',
    CatheterDays: '',
    CAUTICases: '',
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
        DepartmentID: parseInt(formData.DepartmentID),
        HospitalizationDays: parseInt(formData.HospitalizationDays) || 0,
        VentilatorDays: parseInt(formData.VentilatorDays) || 0,
        VAPCases: parseInt(formData.VAPCases) || 0,
        CentralLineDays: parseInt(formData.CentralLineDays) || 0,
        CLABSICases: parseInt(formData.CLABSICases) || 0,
        CatheterDays: parseInt(formData.CatheterDays) || 0,
        CAUTICases: parseInt(formData.CAUTICases) || 0,
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
      DepartmentID: '',
      Month: '',
      HospitalizationDays: '',
      VentilatorDays: '',
      VAPCases: '',
      CentralLineDays: '',
      CLABSICases: '',
      CatheterDays: '',
      CAUTICases: '',
    });
  };

  const getDepartmentName = (id) => {
    const dept = departments.find(d => d.ID === id);
    return dept?.Name || '未知';
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
                  <tr key={m.ID}>
                    <td>{m.DepartmentName || getDepartmentName(m.DepartmentID)}</td>
                    <td>{m.Month}</td>
                    <td>{m.HospitalizationDays}</td>
                    <td>{m.VentilatorUsageRate}%</td>
                    <td>{m.VAPRate}‰</td>
                    <td>{m.CentralLineUsageRate}%</td>
                    <td>{m.CLABSIRate}‰</td>
                    <td>{m.CatheterUsageRate}%</td>
                    <td>{m.CAUTIRate}‰</td>
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
              <h3>录入目标性监测数据</h3>
              <button className="modal-close" onClick={() => { setShowModal(false); resetForm(); }}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label>科室 *</label>
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
                <div className="form-group">
                  <label>月份 *</label>
                  <input
                    type="month"
                    value={formData.Month}
                    onChange={(e) => setFormData({ ...formData, Month: e.target.value })}
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
                    value={formData.HospitalizationDays}
                    onChange={(e) => setFormData({ ...formData, HospitalizationDays: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>呼吸机使用天数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.VentilatorDays}
                    onChange={(e) => setFormData({ ...formData, VentilatorDays: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>VAP例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.VAPCases}
                    onChange={(e) => setFormData({ ...formData, VAPCases: e.target.value })}
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
                    value={formData.CentralLineDays}
                    onChange={(e) => setFormData({ ...formData, CentralLineDays: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>CLABSI例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.CLABSICases}
                    onChange={(e) => setFormData({ ...formData, CLABSICases: e.target.value })}
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
                    value={formData.CatheterDays}
                    onChange={(e) => setFormData({ ...formData, CatheterDays: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label>CAUTI例数</label>
                  <input
                    type="number"
                    min="0"
                    value={formData.CAUTICases}
                    onChange={(e) => setFormData({ ...formData, CAUTICases: e.target.value })}
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
