import React, { useState, useEffect } from 'react';
import api from '../api';

const Todos = () => {
  const [todos, setTodos] = useState([]);
  const [message, setMessage] = useState(null);

  const fetchTodos = async () => {
    try {
      const res = await api.get('/todos');
      setTodos(res.data);
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '加载失败' });
    }
  };

  useEffect(() => {
    fetchTodos();
    const interval = setInterval(fetchTodos, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleStatusChange = async (id, status) => {
    try {
      await api.put(`/todos/${id}`, { status });
      setMessage({ type: 'success', text: '状态已更新' });
      fetchTodos();
    } catch (err) {
      setMessage({ type: 'error', text: err.response?.data?.error || '操作失败' });
    }
  };

  return (
    <div>
      <h2>待办事项</h2>
      {message && <div className={`message ${message.type}`}>{message.text}</div>}

      {todos.length === 0 ? (
        <p>暂无待办事项</p>
      ) : (
        <div className="todo-list">
          {todos.map((t) => (
            <div key={t.id} className={`todo-card ${t.overdue ? 'overdue' : ''}`}>
              <div className="todo-header">
                <h4>{t.description}</h4>
                {t.overdue && <span className="overdue-badge">已过期</span>}
              </div>
              <p>
                责任人：{t.responsible || '未指定'} |
                截止日期：{t.due_date ? new Date(t.due_date).toLocaleDateString() : '未指定'} |
                状态：
                <select value={t.status} onChange={(e) => handleStatusChange(t.id, e.target.value)}>
                  <option value="待处理">待处理</option>
                  <option value="处理中">处理中</option>
                  <option value="已完成">已完成</option>
                </select>
              </p>
              <p className="todo-meta">
                创建时间：{new Date(t.created_at).toLocaleString()}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Todos;
