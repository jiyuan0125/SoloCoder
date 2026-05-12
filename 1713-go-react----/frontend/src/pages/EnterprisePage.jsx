import { useState, useEffect } from 'react';
import { api } from '../api';

const CATEGORIES = ['粉尘类', '化学毒物类', '物理因素类', '生物因素类'];
const FACTORS = {
  '粉尘类': ['矽尘', '煤尘', '石棉尘'],
  '化学毒物类': ['苯', '铅', '汞', '甲醛'],
  '物理因素类': ['噪声', '高温', '辐射'],
  '生物因素类': ['布鲁氏菌'],
};
const HAZARD_LEVELS = ['一般', '较重', '严重'];

export default function EnterprisePage() {
  const [enterprises, setEnterprises] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editingEnt, setEditingEnt] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    unified_social_code: '',
    industry: '',
    region: '',
    contact_person: '',
    contact_phone: '',
    address: '',
  });
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const [selectedEnt, setSelectedEnt] = useState(null);
  const [factors, setFactors] = useState([]);
  const [workers, setWorkers] = useState([]);
  const [showFactorModal, setShowFactorModal] = useState(false);
  const [showWorkerModal, setShowWorkerModal] = useState(false);
  const [factorForm, setFactorForm] = useState({
    workshop: '',
    post_name: '',
    category: '粉尘类',
    factor_name: '矽尘',
    hazard_level: '一般',
    protective_measures: '',
    last_monitor_value: 0,
  });
  const [workerForm, setWorkerForm] = useState({
    name: '',
    id_card: '',
    gender: '男',
    post_name: '',
    workshop: '',
  });

  useEffect(() => {
    loadEnterprises();
  }, []);

  async function loadEnterprises() {
    setLoading(true);
    try {
      const res = await api.enterprises.list();
      setEnterprises(res.data || []);
    } catch (e) {
      setError(e.message);
    }
    setLoading(false);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError(null);
    try {
      if (editingEnt) {
        await api.enterprises.update(editingEnt.id, formData);
        setSuccess('企业信息更新成功');
      } else {
        await api.enterprises.create(formData);
        setSuccess('企业登记成功');
      }
      setShowModal(false);
      setEditingEnt(null);
      resetForm();
      loadEnterprises();
      setTimeout(() => setSuccess(null), 3000);
    } catch (e) {
      setError(e.message);
    }
  }

  function resetForm() {
    setFormData({
      name: '',
      unified_social_code: '',
      industry: '',
      region: '',
      contact_person: '',
      contact_phone: '',
      address: '',
    });
  }

  function openEdit(ent) {
    setEditingEnt(ent);
    setFormData({
      name: ent.name,
      unified_social_code: ent.unified_social_code,
      industry: ent.industry || '',
      region: ent.region || '',
      contact_person: ent.contact_person || '',
      contact_phone: ent.contact_phone || '',
      address: ent.address || '',
    });
    setShowModal(true);
  }

  async function handleDelete(id) {
    if (!confirm('确定要删除该企业吗？')) return;
    try {
      await api.enterprises.remove(id);
      loadEnterprises();
      if (selectedEnt?.id === id) {
        setSelectedEnt(null);
      }
    } catch (e) {
      setError(e.message);
    }
  }

  async function selectEnterprise(ent) {
    setSelectedEnt(ent);
    try {
      const [factorRes, workerRes] = await Promise.all([
        api.hazardFactors.list({ enterprise_id: ent.id, page: 1, size: 100 }),
        api.workers.list({ enterprise_id: ent.id, page: 1, size: 100 }),
      ]);
      setFactors(factorRes.data || []);
      setWorkers(workerRes.data || []);
    } catch (e) {
      setError(e.message);
    }
  }

  async function handleFactorSubmit(e) {
    e.preventDefault();
    setError(null);
    try {
      await api.hazardFactors.create({
        enterprise_id: selectedEnt.id,
        ...factorForm,
      });
      setShowFactorModal(false);
      setFactorForm({
        workshop: '',
        post_name: '',
        category: '粉尘类',
        factor_name: '矽尘',
        hazard_level: '一般',
        protective_measures: '',
        last_monitor_value: 0,
      });
      if (selectedEnt) selectEnterprise(selectedEnt);
    } catch (e) {
      setError(e.message);
    }
  }

  async function handleWorkerSubmit(e) {
    e.preventDefault();
    setError(null);
    try {
      await api.workers.create({
        enterprise_id: selectedEnt.id,
        ...workerForm,
      });
      setShowWorkerModal(false);
      setWorkerForm({
        name: '',
        id_card: '',
        gender: '男',
        post_name: '',
        workshop: '',
      });
      if (selectedEnt) selectEnterprise(selectedEnt);
    } catch (e) {
      setError(e.message);
    }
  }

  function formatDate(d) {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('zh-CN');
  }

  return (
    <div>
      <div className="page-header">
        <h1>企业管理</h1>
        <button className="btn btn-primary" onClick={() => { setEditingEnt(null); resetForm(); setShowModal(true); }}>
          + 登记企业
        </button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <h3>企业列表</h3>
        {loading ? (
          <div className="loading">加载中...</div>
        ) : enterprises.length === 0 ? (
          <div className="empty">暂无企业数据，请点击上方按钮登记企业</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>企业名称</th>
                <th>统一社会信用代码</th>
                <th>行业</th>
                <th>地区</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {enterprises.map((ent) => (
                <tr key={ent.id} onClick={() => selectEnterprise(ent)} style={{ cursor: 'pointer' }}>
                  <td>{ent.name}</td>
                  <td>{ent.unified_social_code}</td>
                  <td>{ent.industry || '-'}</td>
                  <td>{ent.region || '-'}</td>
                  <td><span className="badge badge-status">{ent.status}</span></td>
                  <td>
                    <button className="btn btn-sm btn-secondary" onClick={(e) => { e.stopPropagation(); openEdit(ent); }}>编辑</button>
                    <button className="btn btn-sm btn-danger" onClick={(e) => { e.stopPropagation(); handleDelete(ent.id); }}>删除</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {selectedEnt && (
        <>
          <div className="card">
            <div className="page-header" style={{ marginBottom: 12 }}>
              <h3 style={{ marginBottom: 0 }}>{selectedEnt.name} - 危害因素</h3>
              <button className="btn btn-primary btn-sm" onClick={() => setShowFactorModal(true)}>+ 添加危害因素</button>
            </div>
            {factors.length === 0 ? (
              <div className="empty">暂无危害因素数据</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>车间/岗位</th>
                    <th>危害因素类别</th>
                    <th>危害因素名称</th>
                    <th>危害等级</th>
                    <th>监测值</th>
                    <th>限值</th>
                    <th>监测日期</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  {factors.map((f) => (
                    <tr key={f.id} className={f.exceed_limit ? 'row-exceed' : ''}>
                      <td>{f.workshop}/{f.post_name}</td>
                      <td>{f.category}</td>
                      <td>{f.factor_name}</td>
                      <td>{f.hazard_level}</td>
                      <td>{f.last_monitor_value} {f.monitor_unit}</td>
                      <td>{f.exposure_limit || '-'}</td>
                      <td>{formatDate(f.last_monitor_date)}</td>
                      <td>
                        <span className={`badge ${f.exceed_limit ? 'badge-exceed' : 'badge-normal'}`}>
                          {f.exceed_limit ? '超标' : '正常'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="card">
            <div className="page-header" style={{ marginBottom: 12 }}>
              <h3 style={{ marginBottom: 0 }}>{selectedEnt.name} - 劳动者</h3>
              <button className="btn btn-primary btn-sm" onClick={() => setShowWorkerModal(true)}>+ 添加劳动者</button>
            </div>
            {workers.length === 0 ? (
              <div className="empty">暂无劳动者数据</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>姓名</th>
                    <th>身份证号</th>
                    <th>性别</th>
                    <th>车间</th>
                    <th>岗位</th>
                    <th>入职日期</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  {workers.map((w) => (
                    <tr key={w.id}>
                      <td>{w.name}</td>
                      <td>{w.id_card}</td>
                      <td>{w.gender}</td>
                      <td>{w.workshop || '-'}</td>
                      <td>{w.post_name || '-'}</td>
                      <td>{formatDate(w.entry_date)}</td>
                      <td><span className="badge badge-status">{w.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{editingEnt ? '编辑企业' : '登记企业'}</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>企业名称 *</label>
                <input required value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} />
              </div>
              <div className="form-group">
                <label>统一社会信用代码 *</label>
                <input required value={formData.unified_social_code} onChange={(e) => setFormData({ ...formData, unified_social_code: e.target.value })} />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>行业</label>
                  <input value={formData.industry} onChange={(e) => setFormData({ ...formData, industry: e.target.value })} placeholder="如：制造业、采矿业" />
                </div>
                <div className="form-group">
                  <label>地区</label>
                  <input value={formData.region} onChange={(e) => setFormData({ ...formData, region: e.target.value })} placeholder="如：上海市" />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>联系人</label>
                  <input value={formData.contact_person} onChange={(e) => setFormData({ ...formData, contact_person: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>联系电话</label>
                  <input value={formData.contact_phone} onChange={(e) => setFormData({ ...formData, contact_phone: e.target.value })} />
                </div>
              </div>
              <div className="form-group">
                <label>地址</label>
                <input value={formData.address} onChange={(e) => setFormData({ ...formData, address: e.target.value })} />
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>取消</button>
                <button type="submit" className="btn btn-primary">{editingEnt ? '保存' : '登记'}</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showFactorModal && (
        <div className="modal-overlay" onClick={() => setShowFactorModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>添加危害因素</h3>
              <button className="modal-close" onClick={() => setShowFactorModal(false)}>×</button>
            </div>
            <form onSubmit={handleFactorSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label>车间 *</label>
                  <input required value={factorForm.workshop} onChange={(e) => setFactorForm({ ...factorForm, workshop: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>岗位 *</label>
                  <input required value={factorForm.post_name} onChange={(e) => setFactorForm({ ...factorForm, post_name: e.target.value })} />
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>危害因素类别 *</label>
                  <select value={factorForm.category} onChange={(e) => setFactorForm({ ...factorForm, category: e.target.value, factor_name: FACTORS[e.target.value][0] })}>
                    {CATEGORIES.map((c) => <option key={c} value={c}>{c}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>危害因素名称 *</label>
                  <select value={factorForm.factor_name} onChange={(e) => setFactorForm({ ...factorForm, factor_name: e.target.value })}>
                    {FACTORS[factorForm.category].map((f) => <option key={f} value={f}>{f}</option>)}
                  </select>
                </div>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>危害等级 *</label>
                  <select value={factorForm.hazard_level} onChange={(e) => setFactorForm({ ...factorForm, hazard_level: e.target.value })}>
                    {HAZARD_LEVELS.map((l) => <option key={l} value={l}>{l}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>监测值</label>
                  <input type="number" step="0.01" value={factorForm.last_monitor_value} onChange={(e) => setFactorForm({ ...factorForm, last_monitor_value: parseFloat(e.target.value) || 0 })} />
                </div>
              </div>
              <div className="form-group">
                <label>防护措施</label>
                <textarea rows="2" value={factorForm.protective_measures} onChange={(e) => setFactorForm({ ...factorForm, protective_measures: e.target.value })} />
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowFactorModal(false)}>取消</button>
                <button type="submit" className="btn btn-primary">添加</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showWorkerModal && (
        <div className="modal-overlay" onClick={() => setShowWorkerModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>添加劳动者</h3>
              <button className="modal-close" onClick={() => setShowWorkerModal(false)}>×</button>
            </div>
            <form onSubmit={handleWorkerSubmit}>
              <div className="form-row">
                <div className="form-group">
                  <label>姓名 *</label>
                  <input required value={workerForm.name} onChange={(e) => setWorkerForm({ ...workerForm, name: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>性别 *</label>
                  <select value={workerForm.gender} onChange={(e) => setWorkerForm({ ...workerForm, gender: e.target.value })}>
                    <option value="男">男</option>
                    <option value="女">女</option>
                  </select>
                </div>
              </div>
              <div className="form-group">
                <label>身份证号 *</label>
                <input required value={workerForm.id_card} onChange={(e) => setWorkerForm({ ...workerForm, id_card: e.target.value })} />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>车间</label>
                  <input value={workerForm.workshop} onChange={(e) => setWorkerForm({ ...workerForm, workshop: e.target.value })} />
                </div>
                <div className="form-group">
                  <label>岗位</label>
                  <input value={workerForm.post_name} onChange={(e) => setWorkerForm({ ...workerForm, post_name: e.target.value })} />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowWorkerModal(false)}>取消</button>
                <button type="submit" className="btn btn-primary">添加</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
