import React, { useState, useEffect } from 'react'
import { maintenance } from '../api'

const typeLabels = {
  routine: '日常保养',
  deep: '深度保养',
}

const statusLabels = {
  pending: '待执行',
  overdue: '逾期',
  completed: '已完成',
  cancelled: '已取消',
}

export default function MaintenancePlans() {
  const [plans, setPlans] = useState([])
  const [activeTab, setActiveTab] = useState('pending')
  const [showModal, setShowModal] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState(null)
  const [completeForm, setCompleteForm] = useState({
    content: '',
    replaced_parts: '',
    duration_minutes: '',
  })
  const [loading, setLoading] = useState(true)

  const loadPlans = async () => {
    setLoading(true)
    try {
      const data = await maintenance.getAll()
      setPlans(data)
    } catch (err) {
      alert(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPlans()
  }, [])

  const filteredPlans = plans.filter((p) => {
    if (activeTab === 'pending') return p.status === 'pending' || p.status === 'overdue'
    if (activeTab === 'completed') return p.status === 'completed' || p.status === 'cancelled'
    return true
  })

  const stats = {
    pending: plans.filter((p) => p.status === 'pending').length,
    overdue: plans.filter((p) => p.status === 'overdue').length,
    completed: plans.filter((p) => p.status === 'completed').length,
  }

  const handleComplete = async (e) => {
    e.preventDefault()
    if (!selectedPlan) return

    try {
      await maintenance.complete(selectedPlan.id, {
        ...completeForm,
        duration_minutes: parseInt(completeForm.duration_minutes),
      })
      setShowModal(false)
      setSelectedPlan(null)
      setCompleteForm({ content: '', replaced_parts: '', duration_minutes: '' })
      loadPlans()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleCancel = async (planId) => {
    if (!confirm('确定要取消此维护计划吗？')) return

    try {
      await maintenance.cancel(planId)
      loadPlans()
    } catch (err) {
      alert(err.message)
    }
  }

  if (loading) return <div>加载中...</div>

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>维护计划</h2>
      </div>

      <div className="stats-cards">
        <div className="stat-card">
          <div className="stat-value">{stats.pending}</div>
          <div className="stat-label">待执行</div>
        </div>
        <div className="stat-card warning">
          <div className="stat-value">{stats.overdue}</div>
          <div className="stat-label">逾期</div>
        </div>
        <div className="stat-card success">
          <div className="stat-value">{stats.completed}</div>
          <div className="stat-label">已完成</div>
        </div>
      </div>

      <div className="tabs">
        <button
          className={`tab ${activeTab === 'pending' ? 'active' : ''}`}
          onClick={() => setActiveTab('pending')}
        >
          待执行
        </button>
        <button
          className={`tab ${activeTab === 'completed' ? 'active' : ''}`}
          onClick={() => setActiveTab('completed')}
        >
          已完成
        </button>
      </div>

      {filteredPlans.length === 0 ? (
        <div className="empty-state">暂无{activeTab === 'pending' ? '待执行' : '已完成'}的维护计划</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>设备</th>
              <th>维护类型</th>
              <th>计划日期</th>
              <th>状态</th>
              <th>完成日期</th>
              <th>耗时</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredPlans.map((plan) => (
              <tr key={plan.id}>
                <td>{plan.device?.name || '-'}</td>
                <td>{typeLabels[plan.type]}</td>
                <td>{new Date(plan.scheduled_date).toLocaleDateString()}</td>
                <td>
                  <span className={`status-badge ${plan.status === 'overdue' ? 'wo-in_progress' : plan.status === 'completed' ? 'wo-completed' : 'wo-pending'}`}>
                    {statusLabels[plan.status]}
                  </span>
                </td>
                <td>{plan.completed_date ? new Date(plan.completed_date).toLocaleDateString() : '-'}</td>
                <td>{plan.duration_minutes ? `${plan.duration_minutes} 分钟` : '-'}</td>
                <td>
                  {(plan.status === 'pending' || plan.status === 'overdue') && (
                    <>
                      <button
                        className="btn btn-success"
                        style={{ marginRight: '0.5rem' }}
                        onClick={() => {
                          setSelectedPlan(plan)
                          setShowModal(true)
                        }}
                      >
                        完成
                      </button>
                      <button className="btn btn-secondary" onClick={() => handleCancel(plan.id)}>
                        取消
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showModal && selectedPlan && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>完成维护</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleComplete}>
              <p style={{ marginBottom: '1rem' }}>
                <strong>设备:</strong> {selectedPlan.device?.name}<br />
                <strong>类型:</strong> {typeLabels[selectedPlan.type]}
              </p>
              <div className="form-group">
                <label>维护内容 *</label>
                <textarea
                  value={completeForm.content}
                  onChange={(e) => setCompleteForm({ ...completeForm, content: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>更换配件</label>
                <input
                  type="text"
                  value={completeForm.replaced_parts}
                  onChange={(e) => setCompleteForm({ ...completeForm, replaced_parts: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label>耗时（分钟） *</label>
                <input
                  type="number"
                  value={completeForm.duration_minutes}
                  onChange={(e) => setCompleteForm({ ...completeForm, duration_minutes: e.target.value })}
                  required
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">确认完成</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
