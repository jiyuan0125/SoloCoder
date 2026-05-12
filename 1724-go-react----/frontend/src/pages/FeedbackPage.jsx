import React, { useState, useEffect } from 'react'
import axios from 'axios'

const statusLabels = {
  'pending': '待填写',
  'plan_filled': '已提交计划',
  'reviewing': '审核中',
  'closed': '已关闭'
}

function FeedbackPage() {
  const [feedbackItems, setFeedbackItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedItem, setSelectedItem] = useState(null)
  const [improvementPlan, setImprovementPlan] = useState('')
  const [saving, setSaving] = useState(false)
  const [filter, setFilter] = useState('all')

  useEffect(() => {
    fetchFeedback()
  }, [])

  const fetchFeedback = async () => {
    try {
      setLoading(true)
      setError('')
      
      const response = await axios.get('/api/feedback')
      setFeedbackItems(response.data)
    } catch (err) {
      setError('加载反馈数据失败')
    } finally {
      setLoading(false)
    }
  }

  const handleViewDetails = async (item) => {
    try {
      const response = await axios.get(`/api/feedback/${item.id}`)
      setSelectedItem(response.data)
      setImprovementPlan(response.data.improvement_plan || '')
    } catch (err) {
      setError('加载详情失败')
    }
  }

  const handleSavePlan = async () => {
    if (!selectedItem) return
    
    try {
      setSaving(true)
      setError('')
      
      await axios.put(`/api/feedback/${selectedItem.id}/plan`, {
        improvement_plan: improvementPlan
      })
      
      alert('改进计划保存成功！')
      fetchFeedback()
      setSelectedItem(null)
    } catch (err) {
      setError('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleUpdateStatus = async (status) => {
    if (!selectedItem) return
    
    try {
      await axios.put(`/api/feedback/${selectedItem.id}/status`, { status })
      alert('状态更新成功！')
      fetchFeedback()
      setSelectedItem(null)
    } catch (err) {
      setError('状态更新失败')
    }
  }

  const getStatusClass = (status) => {
    switch (status) {
      case 'pending': return 'status-draft'
      case 'plan_filled': return 'status-active'
      case 'reviewing': return 'status-active'
      case 'closed': return 'status-completed'
      default: return 'status-draft'
    }
  }

  const filteredItems = feedbackItems.filter(item => {
    if (filter === 'all') return true
    return item.status === filter
  })

  const pendingCount = feedbackItems.filter(i => i.status === 'pending').length

  if (loading) return <div className="loading">加载中...</div>

  return (
    <div>
      <div className="card">
        <h2>反馈跟踪</h2>
        <p style={{ color: '#718096', marginTop: '0.5rem' }}>
          综合得分低于3.5分的课程需要填写改进计划
        </p>
      </div>

      {error && <div className="error-message">{error}</div>}

      {pendingCount > 0 && (
        <div className="card" style={{ borderLeft: '4px solid #f56565' }}>
          <strong>⚠️ 有 {pendingCount} 个改进计划待填写，请在截止日期前完成！</strong>
        </div>
      )}

      <div className="card">
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '0.5rem', border: '1px solid #e2e8f0', borderRadius: '4px' }}
          >
            <option value="all">全部状态</option>
            <option value="pending">待填写</option>
            <option value="plan_filled">已提交计划</option>
            <option value="reviewing">审核中</option>
            <option value="closed">已关闭</option>
          </select>
          <button className="btn btn-secondary" onClick={fetchFeedback}>
            刷新
          </button>
        </div>

        {feedbackItems.length === 0 ? (
          <p>暂无需要处理的反馈项</p>
        ) : filteredItems.length === 0 ? (
          <p>当前筛选条件下无数据</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>课程结果ID</th>
                <th>综合得分</th>
                <th>状态</th>
                <th>计划截止日期</th>
                <th>下次复查</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredItems.map(item => (
                <tr key={item.id}>
                  <td>{item.course_result_id}</td>
                  <td>
                    <strong style={{ color: '#c53030' }}>
                      {item.course_result?.overall_score?.toFixed(2) || '-'}
                    </strong>
                  </td>
                  <td>
                    <span className={`status-badge ${getStatusClass(item.status)}`}>
                      {statusLabels[item.status] || item.status}
                    </span>
                  </td>
                  <td>
                    {new Date(item.plan_deadline).toLocaleDateString()}
                    {item.status === 'pending' && new Date(item.plan_deadline) < new Date() && (
                      <span style={{ color: '#c53030', marginLeft: '0.5rem' }}>已逾期</span>
                    )}
                  </td>
                  <td>{new Date(item.next_review_at).toLocaleDateString()}</td>
                  <td>
                    <button
                      className="btn btn-primary"
                      onClick={() => handleViewDetails(item)}
                    >
                      查看详情
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {selectedItem && (
        <div className="card">
          <h3>反馈详情</h3>
          
          <div style={{ marginBottom: '1rem' }}>
            <p><strong>课程结果ID:</strong> {selectedItem.course_result_id}</p>
            <p><strong>综合得分:</strong> {selectedItem.course_result?.overall_score?.toFixed(2)}</p>
            <p><strong>当前状态:</strong> {statusLabels[selectedItem.status]}</p>
            <p><strong>计划截止日期:</strong> {new Date(selectedItem.plan_deadline).toLocaleDateString()}</p>
          </div>

          {selectedItem.status === 'pending' || selectedItem.status === 'plan_filled' ? (
            <div className="form-group">
              <label>改进计划 (请在30天内提交)</label>
              <textarea
                value={improvementPlan}
                onChange={(e) => setImprovementPlan(e.target.value)}
                placeholder="请填写改进计划，包括具体措施和时间安排..."
                style={{ minHeight: '150px' }}
              />
              <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
                <button
                  className="btn btn-primary"
                  onClick={handleSavePlan}
                  disabled={saving || !improvementPlan.trim()}
                >
                  {saving ? '保存中...' : '保存改进计划'}
                </button>
              </div>
            </div>
          ) : (
            <div className="form-group">
              <label>改进计划</label>
              <div style={{ 
                padding: '1rem', 
                background: '#f7fafc', 
                borderRadius: '4px',
                whiteSpace: 'pre-wrap'
              }}>
                {selectedItem.improvement_plan || '暂无改进计划'}
              </div>
            </div>
          )}

          <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
            {selectedItem.status === 'plan_filled' && (
              <button 
                className="btn btn-secondary"
                onClick={() => handleUpdateStatus('reviewing')}
              >
                提交审核
              </button>
            )}
            {selectedItem.status === 'reviewing' && (
              <button 
                className="btn btn-primary"
                onClick={() => handleUpdateStatus('closed')}
              >
                关闭反馈
              </button>
            )}
            <button 
              className="btn btn-secondary"
              onClick={() => setSelectedItem(null)}
            >
              关闭
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default FeedbackPage
