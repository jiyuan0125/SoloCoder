import React, { useState, useEffect } from 'react';
import api from '../api';

const Labs = () => {
  const [labs, setLabs] = useState([]);
  const [form, setForm] = useState({ name: '', building_no: '', room_no: '', manager: '', danger_level: '丙' });
  const [editing, setEditing] = useState(null);
  const [message, setMessage] = useState(null);

  const fetchLabs = async () => {
    try {
      const res = await api.get('/labs');
      setLabs(res.data);
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '加载失败' });
    }
  };

  useEffect(() => {
    fetchLabs();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editing) {
        await api.put(`/labs/${editing.id}`, form);
      } else {
        await api.post('/labs', form);
      }
      setForm({ name: '', building_no: '', room_no: '', manager: '', danger_level: '丙' });
      setEditing(null);
      setMessage({ type: 'success', text: '保存成功' });
      fetchLabs();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  const handleEdit = (lab) => {
    setEditing(lab);
    setForm({ ...lab });
  };

  const handleDelete = async (id) => {
    if (!window.confirm('确定删除？')) return;
    try {
      await api.delete(`/labs/${id}`);
      setMessage({ type: 'success', text: '删除成功' });
      fetchLabs();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '删除失败' });
    }
  };

  return (
    <div>
      <h2>实验室管理</h2>
      {message && <div className={`message ${message.type}`}>{message.text}</div>}
      
      <form onSubmit={handleSubmit}>
        <h3>{editing ? '编辑实验室' : '新增实验室'}</h3>
        <input placeholder="实验室名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
        <input placeholder="楼栋号" value={form.building_no} onChange={(e) => setForm({ ...form, building_no: e.target.value })} required />
        <input placeholder="房间号" value={form.room_no} onChange={(e) => setForm({ ...form, room_no: e.target.value })} required />
        <input placeholder="负责人" value={form.manager} onChange={(e) => setForm({ ...form, manager: e.target.value })} required />
        <select value={form.danger_level} onChange={(e) => setForm({ ...form, danger_level: e.target.value })}>
          <option value="甲">甲</option>
          <option value="乙">乙</option>
          <option value="丙">丙</option>
          <option value="丁">丁</option>
        </select>
        <button type="submit">{editing ? '更新' : '创建'}</button>
        {editing && <button type="button" onClick={() => { setEditing(null); setForm({ name: '', building_no: '', room_no: '', manager: '', danger_level: '丙' }); }}>取消</button>}
      </form>

      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>楼栋号</th>
            <th>房间号</th>
            <th>负责人</th>
            <th>危险等级</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {labs.map((lab) => (
            <tr key={lab.id}>
              <td>{lab.id}</td>
              <td>{lab.name}</td>
              <td>{lab.building_no}</td>
              <td>{lab.room_no}</td>
              <td>{lab.manager}</td>
              <td>
                <span className={`level-${lab.danger_level}`}>{lab.danger_level}</span>
              </td>
              <td>
                <button onClick={() => handleEdit(lab)}>编辑</button>
                <button onClick={() => handleDelete(lab.id)}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Labs;
