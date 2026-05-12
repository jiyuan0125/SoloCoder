import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import axios from 'axios'

const statusLabels = {
  'draft': '起草',
  'review': '内部审核',
  'final': '终审',
  'pending': '待签署',
  'active': '进行中',
  'completed': '已完成',
  'cancelled': '已取消'
}

function TasksPage() {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const navigate = useNavigate()

  const [newTask, setNewTask] = useState({
    semester: '',
    startDate: '',
    endDate: '',
    courseIds: ''
  })

  useEffect(() => {
    fetchTasks()
  }, [])

  const fetchTasks = async () => {
    try {
      setLoading(true)
      const response = await axios.get('/api/tasks')
      setTasks(response.data)
    } catch (err) {
      setError('加载任务列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateTask = async (e) => {
    e.preventDefault()
    try {
      setError('')
      setSuccess('')
      
      const courseIds = newTask.courseIds.split(',').map(id => parseInt(id.trim())).filter(id => !isNaN(id))
      
      const response = await axios.post('/api/tasks', {
        semester: newTask.semester,
        start_date: new Date(newTask.startDate).toISOString(),
        end_date: new Date(newTask.endDate).toISOString(),
        course_ids: courseIds
      })
      
      setSuccess('任务创建成功')
      setShowForm(false)
      setNewTask({ semester: '', startDate: '', endDate: '', courseIds: '' })
      fetchTasks()
    } catch (err) {
      if (err.response?.status === 409) {
        setError('错误 (409): 该学期评估任务已存在')
      } else {
        setError(err.response?.data?.error || '创建任务失败')
      }
    }
  }

  const handleAction = async (taskId, action) => {
    try {
      await axios.post(`/api/tasks/${taskId}/actions/${action}`, {}, {
        headers: { 'X-User-Role': 'admin' }
      })
      fetchTasks()
    } catch (err) {
      setError(err.response?.data?.error || '操作失败')
    }
  }

  const handleStartEvaluation = (taskId, courseId) => {
    navigate(`/questionnaire/${taskId}/${courseId}`)
  }

  const getStatusClass = (status) => {
    if (status === 'active') return 'status-active'
    if (status === 'completed') return 'status-completed'
    return 'status-draft'
  }

  if (loading) return <div className="loading">加载中...</div>

  return (
    <div>
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>评估任务管理</h2>
          <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
            {showForm ? '取消' : '新建任务'}
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}

      {showForm && (
        <div className="card">
          <h3>创建新评估任务</h3>
          <form onSubmit={handleCreateTask}>
            <div className="form-group">
              <label>学期标识</label>
              <input
                type="text"
                value={newTask.semester}
                onChange={(e) => setNewTask({ ...newTask, semester: e.target.value })}
                placeholder="例如: 2024-2025-1"
                required
              />
            </div>
            <div className="form-group">
              <label>开始日期</label>
              <input
                type="date"
                value={newTask.startDate}
                onChange={(e) => setNewTask({ ...newTask, startDate: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>结束日期</label>
              <input
                type="date"
                value={newTask.endDate}
                onChange={(e) => setNewTask({ ...newTask, endDate: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>课程ID (用逗号分隔)</label>
              <input
                type="text"
                value={newTask.courseIds}
                onChange={(e) => setNewTask({ ...newTask, courseIds: e.target.value })}
                placeholder="例如: 1,2,3"
                required
              />
            </div>
            <button type="submit" className="btn btn-primary">创建任务</button>
          </form>
        </div>
      )}

      <div className="card">
        <h3>任务列表</h3>
        {tasks.length === 0 ? (
          <p>暂无评估任务</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>学期</th>
                <th>开始日期</th>
                <th>结束日期</th>
                <th>状态</th>
                <th>课程数</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {tasks.map(task => (
                <tr key={task.id}>
                  <td>{task.semester}</td>
                  <td>{new Date(task.start_date).toLocaleDateString()}</td>
                  <td>{new Date(task.end_date).toLocaleDateString()}</td>
                  <td>
                    <span className={`status-badge ${getStatusClass(task.status)}`}>
                      {statusLabels[task.status] || task.status}
                    </span>
                  </td>
                  <td>{task.courses?.length || 0}</td>
                  <td>
                    <div className="action-buttons">
                      {task.status === 'draft' && (
                        <button className="btn btn-primary" onClick={() => handleAction(task.id, 'approve')}>
                          提交审核
                        </button>
                      )}
                      {task.status === 'review' && (
                        <>
                          <button className="btn btn-primary" onClick={() => handleAction(task.id, 'approve')}>
                            终审
                          </button>
                        </>
                      )}
                      {task.status === 'final' && (
                        <>
                          <button className="btn btn-primary" onClick={() => handleAction(task.id, 'approve')}>
                            生效
                          </button>
                          <button className="btn btn-secondary" onClick={() => handleAction(task.id, 'reject')}>
                            退回
                          </button>
                        </>
                      )}
                      {task.status === 'active' && task.courses?.length > 0 && (
                        <button
                          className="btn btn-primary"
                          onClick={() => handleStartEvaluation(task.id, task.courses[0].id)}
                        >
                          填写问卷
                        </button>
                      )}
                      {task.status === 'active' && (
                        <button className="btn btn-secondary" onClick={() => handleAction(task.id, 'cancel')}>
                          取消
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

export default TasksPage
