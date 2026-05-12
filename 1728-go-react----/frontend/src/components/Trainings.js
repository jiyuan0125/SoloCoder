import React, { useState, useEffect } from 'react';
import api from '../api';

const Trainings = () => {
  const [trainings, setTrainings] = useState([]);
  const [form, setForm] = useState({ topic: '', date: '', duration_hours: 0, instructor: '' });
  const [participantForm, setParticipantForm] = useState({ training_id: '', user_id: '', user_name: '', score: 0 });
  const [message, setMessage] = useState(null);

  const fetchTrainings = async () => {
    try {
      const res = await api.get('/trainings');
      setTrainings(res.data);
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '加载失败' });
    }
  };

  useEffect(() => {
    fetchTrainings();
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post('/trainings', form);
      setForm({ topic: '', date: '', duration_hours: 0, instructor: '' });
      setMessage({ type: 'success', text: '培训创建成功' });
      fetchTrainings();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  const handleAddParticipant = async (e) => {
    e.preventDefault();
    try {
      await api.post(`/trainings/${participantForm.training_id}/participants`, {
        user_id: participantForm.user_id,
        user_name: participantForm.user_name,
        score: participantForm.score,
      });
      setParticipantForm({ training_id: '', user_id: '', user_name: '', score: 0 });
      setMessage({ type: 'success', text: '参训人员添加成功' });
      fetchTrainings();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  return (
    <div>
      <h2>培训记录</h2>
      {message && <div className={`message ${message.type}`}>{message.text}</div>}

      <form onSubmit={handleSubmit} className="inline-form">
        <h3>创建培训</h3>
        <input placeholder="培训主题" value={form.topic} onChange={(e) => setForm({ ...form, topic: e.target.value })} required />
        <input type="date" value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} required />
        <input type="number" step="0.5" placeholder="时长(小时)" value={form.duration_hours || ''} onChange={(e) => setForm({ ...form, duration_hours: parseFloat(e.target.value) })} required />
        <input placeholder="讲师" value={form.instructor} onChange={(e) => setForm({ ...form, instructor: e.target.value })} required />
        <button type="submit">创建</button>
      </form>

      <form onSubmit={handleAddParticipant} className="inline-form">
        <h3>添加参训人员</h3>
        <select value={participantForm.training_id} onChange={(e) => setParticipantForm({ ...participantForm, training_id: e.target.value })}>
          <option value="">选择培训</option>
          {trainings.map((t) => (
            <option key={t.id} value={t.id}>{t.topic} - {new Date(t.date).toLocaleDateString()}</option>
          ))}
        </select>
        <input placeholder="用户ID" value={participantForm.user_id} onChange={(e) => setParticipantForm({ ...participantForm, user_id: e.target.value })} required />
        <input placeholder="姓名" value={participantForm.user_name} onChange={(e) => setParticipantForm({ ...participantForm, user_name: e.target.value })} required />
        <input type="number" step="0.5" placeholder="考核分数" value={participantForm.score || ''} onChange={(e) => setParticipantForm({ ...participantForm, score: parseFloat(e.target.value) })} required />
        <button type="submit">添加</button>
      </form>

      <div className="training-list">
        {trainings.map((t) => (
          <div key={t.id} className="training-card">
            <h4>{t.topic}</h4>
            <p>日期：{new Date(t.date).toLocaleDateString()} | 时长：{t.duration_hours}小时 | 讲师：{t.instructor}</p>
            <h5>参训人员 ({t.participants.length}人)</h5>
            <table>
              <thead>
                <tr>
                  <th>用户ID</th>
                  <th>姓名</th>
                  <th>分数</th>
                  <th>是否通过</th>
                </tr>
              </thead>
              <tbody>
                {t.participants.map((p) => (
                  <tr key={p.user_id}>
                    <td>{p.user_id}</td>
                    <td>{p.user_name}</td>
                    <td>{p.score}</td>
                    <td>
                      <span className={p.passed ? 'status-instock' : 'status-expired'}>
                        {p.passed ? '通过' : '未通过'}
                      </span>
                    </td>
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

export default Trainings;
