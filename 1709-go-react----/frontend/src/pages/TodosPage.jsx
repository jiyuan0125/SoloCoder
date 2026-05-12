import React, { useState, useEffect } from 'react'
import { todosAPI } from '../api'

const getTodoStatusBadge = (status) => {
  const mapping = {
    '待处理': 'badge-pending',
    '处理中': 'badge-warning',
    '已完成': 'badge-success',
    '已逾期': 'badge-danger'
  }
  return mapping[status] || 'badge-pending'
}

const getTodoTypeIcon = (type) => {
  const mapping = {
    '器官评估': '🔬',
    '等待匹配': '⏳',
    '通知受体': '📞',
    '安排手术': '🏥',
    '术后随访': '📋'
  }
  return mapping[type] || '📝'
}

export default function TodosPage() {
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('all')
  const [message, setMessage] = useState(null)

  const loadData = async () => {
    try {
      setLoading(true)
      const data = await todosAPI.list(1, 100)
      setTodos(data.data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleUpdateStatus = async (todoId, newStatus) => {
    try {
      await todosAPI.updateStatus(todoId, newStatus)
      setMessage({ type: 'success', text: '状态更新成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const filteredTodos = filter === 'all' 
    ? todos 
    : todos.filter(t => t.status === filter)

  const stats = {
    pending: todos.filter(t => t.status === '待处理').length,
    inProgress: todos.filter(t => t.status === '处理中').length,
    completed: todos.filter(t => t.status === '已完成').length,
    overdue: todos.filter(t => t.status === '已逾期').length
  }

  return (
    <div>
      <div className="page-header">
        <h1>待办事项</h1>
        <p>管理器官捐献流程中的待办任务</p>
      </div>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      <div className="stats-grid">
        <div className="stat-card">
          <h3>待处理</h3>
          <div className="value">{stats.pending}</div>
        </div>
        <div className="stat-card">
          <h3>处理中</h3>
          <div className="value">{stats.inProgress}</div>
        </div>
        <div className="stat-card">
          <h3>已完成</h3>
          <div className="value">{stats.completed}</div>
        </div>
        <div className="stat-card">
          <h3>已逾期</h3>
          <div className="value">{stats.overdue}</div>
        </div>
      </div>

      <div className="card">
        <div className="tabs">
          <div className={`tab ${filter === 'all' ? 'active' : ''}`} onClick={() => setFilter('all')}>全部</div>
          <div className={`tab ${filter === '待处理' ? 'active' : ''}`} onClick={() => setFilter('待处理')}>待处理</div>
          <div className={`tab ${filter === '处理中' ? 'active' : ''}`} onClick={() => setFilter('处理中')}>处理中</div>
          <div className={`tab ${filter === '已完成' ? 'active' : ''}`} onClick={() => setFilter('已完成')}>已完成</div>
          <div className={`tab ${filter === '已逾期' ? 'active' : ''}`} onClick={() => setFilter('已逾期')}>已逾期</div>
      </div>

        {loading ? (
          <div className="loading">加载中...</div>
        ) : filteredTodos.length === 0 ? (
          <div className="empty-state">
            <h3>暂无待办事项</h3>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
              <tr>
                <th>类型</th>
                <th>描述</th>
                <th>负责人</th>
                <th>截止日期</th>
                <th>状态</th>
                <th>关联ID</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredTodos.map(todo => (
                <tr key={todo.id}>
                  <td>
                    <span style={{marginRight: '8px'}}>{getTodoTypeIcon(todo.type)}</span>
                    {todo.type}
                  </td>
                  <td>{todo.description}</td>
                  <td>{todo.assignee}</td>
                  <td>{new Date(todo.due_date).toLocaleDateString()}</td>
                  <td><span className={`badge ${getTodoStatusBadge(todo.status)}`}>{todo.status}</span></td>
                  <td>{todo.related_id?.substring(0, 8)}...</td>
                  <td>
                    {todo.status === '待处理' && (
                      <button className="btn btn-sm btn-warning"
                        onClick={() => handleUpdateStatus(todo.id, '处理中')}>
                        开始处理
                      </button>
                    )}
                    {todo.status === '处理中' && (
                      <button className="btn btn-sm btn-success"
                        onClick={() => handleUpdateStatus(todo.id, '已完成')}>
                        完成
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      </div>
    </div>
  )
}
