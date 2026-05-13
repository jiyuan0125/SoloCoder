import React, { useState, useEffect } from 'react';
import api from '../api';

const Checks = () => {
  const [checks, setChecks] = useState([]);
  const [labs, setLabs] = useState([]);
  const [form, setForm] = useState({ lab_id: '', inspector: '', items: [
    { item: '消防设施', result: '合格', responsible: '', deadline: '' },
    { item: '通风系统', result: '合格', responsible: '', deadline: '' },
    { item: '防护装备', result: '合格', responsible: '', deadline: '' },
    { item: '危化品存放', result: '合格', responsible: '', deadline: '' },
    { item: '应急预案', result: '合格', responsible: '', deadline: '' },
  ]});
  const [message, setMessage] = useState(null);

  const fetchChecks = async () => {
    try {
      const res = await api.get('/checks');
      setChecks(res.data);
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
    fetchChecks();
    fetchLabs();
  }, []);

  const handleItemResultChange = (index, result) => {
    const newItems = [...form.items];
    newItems[index] = { ...newItems[index], result };
    setForm({ ...form, items: newItems });
  };

  const handleItemResponsibleChange = (index, responsible) => {
    const newItems = [...form.items];
    newItems[index] = { ...newItems[index], responsible };
    setForm({ ...form, items: newItems });
  };

  const handleItemDeadlineChange = (index, deadline) => {
    const newItems = [...form.items];
    newItems[index] = { ...newItems[index], deadline };
    setForm({ ...form, items: newItems });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await api.post('/checks', form);
      const todosCreated = res.data.todos?.length || 0;
      setMessage({ 
        type: 'success', 
        text: todosCreated > 0 
          ? `检查记录已创建，自动生成 ${todosCreated} 条待办事项` 
          : '检查记录已创建' 
      });
      fetchChecks();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  const getResultColor = (result) => {
    switch (result) {
      case '合格': return 'status-instock';
      case '不合格': return 'status-expired';
      case '待整改': return 'status-using';
      default: return '';
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString();
  };

  return (
    <div>
      <h2>安全检查</h2>
      {message && <div className={`message ${message.type}`}>{message.text}</div>}

      <form onSubmit={handleSubmit}>
        <h3>创建安全检查记录</h3>
        <select 
          value={form.lab_id} 
          onChange={(e) => setForm({ ...form, lab_id: e.target.value })}
          required
        >
          <option value="">选择实验室</option>
          {labs.map((l) => (
            <option key={l.id} value={l.id}>{l.name}</option>
          ))}
        </select>
        <input 
          placeholder="检查人" 
          value={form.inspector} 
          onChange={(e) => setForm({ ...form, inspector: e.target.value })} 
          required 
        />
        
        <div className="check-items-grid">
          {form.items.map((item, index) => (
            <div key={index} className="check-item-row">
              <div className="check-item-label">{item.item}</div>
              <div className="check-item-result">
                <select 
                  value={item.result} 
                  onChange={(e) => handleItemResultChange(index, e.target.value)}
                >
                  <option value="合格">合格</option>
                  <option value="不合格">不合格</option>
                  <option value="待整改">待整改</option>
                </select>
              </div>
              {item.result === '不合格' && (
                <div className="check-item-extra">
                  <input 
                    type="text" 
                    placeholder="责任人" 
                    value={item.responsible}
                    onChange={(e) => handleItemResponsibleChange(index, e.target.value)}
                    style={{ width: '120px', marginRight: '8px' }}
                    required
                  />
                  <input 
                    type="date" 
                    value={item.deadline}
                    onChange={(e) => handleItemDeadlineChange(index, e.target.value)}
                    required
                  />
                </div>
              )}
            </div>
          ))}
        </div>
        <p className="hint">提示：若某项检查为"不合格"，需填写责任人及整改截止日期，系统将自动生成待办事项。</p>
        <button type="submit">提交检查</button>
      </form>

      <div className="check-list">
        {checks.length === 0 ? (
          <p>暂无检查记录</p>
        ) : checks.map((c) => (
          <div key={c.id} className="check-card">
            <h4>检查记录 #{c.id}</h4>
            <p>实验室：{labs.find((l) => l.id === c.lab_id)?.name || c.lab_id} | 检查人：{c.inspector} | 日期：{new Date(c.check_date).toLocaleDateString()}</p>
            <table>
              <thead>
                <tr>
                  <th>检查项目</th>
                  <th>结果</th>
                  <th>责任人</th>
                  <th>截止日期</th>
                </tr>
              </thead>
              <tbody>
                {c.items.map((item, i) => (
                  <tr key={i}>
                    <td>{item.item}</td>
                    <td>
                      <span className={getResultColor(item.result)}>{item.result}</span>
                    </td>
                    <td>{item.responsible || '-'}</td>
                    <td>{formatDate(item.deadline)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Checks;
