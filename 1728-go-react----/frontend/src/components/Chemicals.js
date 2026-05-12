import React, { useState, useEffect } from 'react';
import api from '../api';

const Chemicals = () => {
  const [chemicals, setChemicals] = useState([]);
  const [labs, setLabs] = useState([]);
  const [category, setCategory] = useState('');
  const [labID, setLabID] = useState('');
  const [form, setForm] = useState({ name: '', cas: '', hazard_category: '易燃', quantity: 0, storage_lab_id: '', storage_cabinet: '' });
  const [usageForm, setUsageForm] = useState({ id: '', quantity: 0 });
  const [message, setMessage] = useState(null);

  const fetchChemicals = async () => {
    try {
      const params = {};
      if (category) params.category = category;
      if (labID) params.lab_id = labID;
      const res = await api.get('/chemicals', { params });
      setChemicals(res.data);
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '加载失败' });
    }
  };

  const fetchLabs = async () => {
    try {
      const res = await api.get('/labs');
      setLabs(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchLabs();
  }, []);

  useEffect(() => {
    fetchChemicals();
  }, [category, labID]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post('/chemicals', form);
      setForm({ name: '', cas: '', hazard_category: '易燃', quantity: 0, storage_lab_id: '', storage_cabinet: '' });
      setMessage({ type: 'success', text: '入库成功' });
      fetchChemicals();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  const handleUse = async (e) => {
    e.preventDefault();
    try {
      await api.post(`/chemicals/${usageForm.id}/use`, { quantity: usageForm.quantity });
      setUsageForm({ id: '', quantity: 0 });
      setMessage({ type: 'success', text: '领用成功' });
      fetchChemicals();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '领用失败' });
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case '在库': return 'status-instock';
      case '领用中': return 'status-using';
      case '已用完': return 'status-used';
      case '已过期': return 'status-expired';
      default: return '';
    }
  };

  return (
    <div>
      <h2>危化品台账</h2>
      {message && <div className={`message ${message.type}`}>{message.text}</div>}

      <div className="filters">
        <label>
          危险类别：
          <select value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="">全部</option>
            <option value="易燃">易燃</option>
            <option value="易爆">易爆</option>
            <option value="有毒">有毒</option>
            <option value="腐蚀">腐蚀</option>
            <option value="氧化">氧化</option>
            <option value="其他">其他</option>
          </select>
        </label>
        <label>
          存放位置：
          <select value={labID} onChange={(e) => setLabID(e.target.value)}>
            <option value="">全部</option>
            {labs.map((l) => (
              <option key={l.id} value={l.id}>{l.name} - {l.building_no}栋{l.room_no}室</option>
            ))}
          </select>
        </label>
      </div>

      <form onSubmit={handleSubmit} className="inline-form">
        <h3>危化品入库</h3>
        <input placeholder="化学品名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        <input placeholder="CAS号" value={form.cas} onChange={(e) => setForm({ ...form, cas: e.target.value })} />
        <select value={form.hazard_category} onChange={(e) => setForm({ ...form, hazard_category: e.target.value })}>
          <option value="易燃">易燃</option>
          <option value="易爆">易爆</option>
          <option value="有毒">有毒</option>
          <option value="腐蚀">腐蚀</option>
          <option value="氧化">氧化</option>
          <option value="其他">其他</option>
        </select>
        <input type="number" step="0.01" placeholder="库存数量" value={form.quantity || ''} onChange={(e) => setForm({ ...form, quantity: parseFloat(e.target.value) })} required />
        <select value={form.storage_lab_id} onChange={(e) => setForm({ ...form, storage_lab_id: e.target.value })}>
          <option value="">选择实验室</option>
          {labs.map((l) => (
            <option key={l.id} value={l.id}>{l.name}</option>
          ))}
        </select>
        <input placeholder="柜子编号" value={form.storage_cabinet} onChange={(e) => setForm({ ...form, storage_cabinet: e.target.value })} required />
        <button type="submit">入库</button>
      </form>

      <form onSubmit={handleUse} className="inline-form">
        <h3>领用危化品</h3>
        <select value={usageForm.id} onChange={(e) => setUsageForm({ ...usageForm, id: e.target.value })}>
          <option value="">选择危化品</option>
          {chemicals.filter((c) => c.status !== '已用完' && c.status !== '已过期').map((c) => (
            <option key={c.id} value={c.id}>{c.name} (库存: {c.quantity})</option>
          ))}
        </select>
        <input type="number" step="0.01" placeholder="领用数量" value={usageForm.quantity || ''} onChange={(e) => setUsageForm({ ...usageForm, quantity: parseFloat(e.target.value) })} required />
        <button type="submit">领用</button>
      </form>

      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>CAS号</th>
            <th>危险类别</th>
            <th>库存</th>
            <th>存放位置</th>
            <th>状态</th>
            <th>入库日期</th>
            <th>保质期</th>
          </tr>
        </thead>
        <tbody>
          {chemicals.map((c) => (
            <tr key={c.id}>
              <td>{c.id}</td>
              <td>{c.name}</td>
              <td>{c.cas || '-'}</td>
              <td>{c.hazard_category}</td>
              <td>{c.quantity}</td>
              <td>
                {labs.find((l) => l.id === c.storage_lab_id)?.name || c.storage_lab_id} - {c.storage_cabinet}
              </td>
              <td>
                <span className={getStatusColor(c.status)}>{c.status}</span>
              </td>
              <td>{new Date(c.inbound_date).toLocaleDateString()}</td>
              <td>{c.expiry_date ? new Date(c.expiry_date).toLocaleDateString() : '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Chemicals;
