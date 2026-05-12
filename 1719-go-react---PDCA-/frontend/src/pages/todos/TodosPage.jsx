import React, { useState, useEffect } from 'react'
import { todos } from '../../api'
import { TODO_STATUS, TODO_TYPES, getStatusLabel, getStatusColor } from '../../types'

const TodosPage = () => {
  const [todoList, setTodoList] = useState([])
  const [filterStatus, setFilterStatus] = useState('')
  const [filterType, setFilterType] = useState('')
  const [selectedTodo, setSelectedTodo] = useState(null)

  useEffect(() => {
    loadTodos()
  }, [filterStatus, filterType])

  const loadTodos = async () => {
    try {
      const res = await todos.list(filterStatus, filterType)
      setTodoList(res.data)
    } catch (err) {
      console.error('加载待办失败', err)
    }
  }

  const handleUpdateStatus = async (id, status) => {
    try {
      await todos.updateStatus(id, status)
      loadTodos()
      if (selectedTodo?.id === id) {
        setSelectedTodo({ ...selectedTodo, status })
      }
    } catch (err) {
      alert('更新失败')
    }
  }

  const handleDelete = async (id) => {
    if (!confirm('确定要删除这个待办吗？')) return
    try {
      await todos.delete(id)
      loadTodos()
      if (selectedTodo?.id === id) setSelectedTodo(null)
    } catch (err) {
      alert('删除失败')
    }
  }

  const stats = {
    total: todoList.length,
    pending: todoList.filter(t => t.status === 'pending').length,
    overdue: todoList.filter(t => t.status === 'overdue').length,
    completed: todoList.filter(t => t.status === 'completed').length,
  }

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">待办事项</h1>
        <p className="page-subtitle">查看和管理改进待办及会议行动项</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-value">{stats.total}</div>
          <div className="stat-label">全部待办</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#f39c12' }}>{stats.pending}</div>
          <div className="stat-label">待处理</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#e74c3c' }}>{stats.overdue}</div>
          <div className="stat-label">已逾期</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#27ae60' }}>{stats.completed}</div>
          <div className="stat-label">已完成</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">待办列表</h2>
        </div>

        <div className="filters">
          <div className="filter-item">
            <span className="filter-label">类型:</span>
            <select className="form-select" style={{ width: '140px' }}
              value={filterType} onChange={(e) => setFilterType(e.target.value)}>
              <option value="">全部</option>
              {TODO_TYPES.map(t => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
          </div>
          <div className="filter-item">
            <span className="filter-label">状态:</span>
            <select className="form-select" style={{ width: '140px' }}
              value={filterStatus} onChange={(e) => setFilterStatus(e.target.value)}>
              <option value="">全部</option>
              {TODO_STATUS.filter(s => s.value !== 'archived').map(s => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </div>
        </div>

        <table>
          <thead>
            <tr>
              <th>类型</th>
              <th>标题</th>
              <th>负责人</th>
              <th>科室</th>
              <th>截止日期</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {todoList.map(todo => (
              <tr key={todo.id} onClick={() => setSelectedTodo(todo)} style={{ cursor: 'pointer' }}>
                <td>
                  <span className={`status-badge ${getStatusColor(todo.type, TODO_TYPES)}`}>
                    {getStatusLabel(todo.type, TODO_TYPES)}
                  </span>
                </td>
                <td>{todo.title}</td>
                <td>{todo.responsible}</td>
                <td>{todo.department || '-'}</td>
                <td>{todo.due_date || '-'}</td>
                <td>
                  <span className={`status-badge ${getStatusColor(todo.status, TODO_STATUS)}`}>
                    {getStatusLabel(todo.status, TODO_STATUS)}
                  </span>
                </td>
                <td>
                  <div style={{ display: 'flex', gap: '6px' }}>
                    {todo.status !== 'completed' && todo.status !== 'archived' && (
                      <button className="btn btn-sm btn-success"
                        onClick={(e) => { e.stopPropagation(); handleUpdateStatus(todo.id, 'completed'); }}>
                        完成
                      </button>
                    )}
                    {todo.status === 'pending' && (
                      <button className="btn btn-sm btn-primary"
                        onClick={(e) => { e.stopPropagation(); handleUpdateStatus(todo.id, 'in_progress'); }}>
                        开始
                      </button>
                    )}
                    <button className="btn btn-sm btn-danger"
                      onClick={(e) => { e.stopPropagation(); handleDelete(todo.id); }}>
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {todoList.length === 0 && (
              <tr><td colSpan="7" className="empty-state">暂无待办事项</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {selectedTodo && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">待办详情</h2>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
            <div>
              <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>标题</div>
              <div style={{ fontWeight: '500' }}>{selectedTodo.title}</div>
            </div>
            <div>
              <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>负责人</div>
              <div>{selectedTodo.responsible}</div>
            </div>
            <div>
              <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>类型</div>
              <div>{getStatusLabel(selectedTodo.type, TODO_TYPES)}</div>
            </div>
            <div>
              <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>截止日期</div>
              <div>{selectedTodo.due_date || '-'}</div>
            </div>
            {selectedTodo.description && (
              <div style={{ gridColumn: '1/-1' }}>
                <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>描述</div>
                <div>{selectedTodo.description}</div>
              </div>
            )}
            {selectedTodo.indicator && (
              <div style={{ gridColumn: '1/-1' }}>
                <div style={{ fontSize: '13px', color: '#666', marginBottom: '4px' }}>关联指标</div>
                <div>{selectedTodo.indicator.code} - {selectedTodo.indicator.name}</div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default TodosPage
