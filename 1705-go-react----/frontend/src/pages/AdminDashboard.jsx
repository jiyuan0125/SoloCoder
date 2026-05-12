import React, { useState, useEffect } from 'react'
import api from '../api.js'

function AdminDashboard({ user, onLogout }) {
  const [stats, setStats] = useState(null)
  const [dateRange, setDateRange] = useState({ start: '', end: '' })
  const [loading, setLoading] = useState(false)

  const fetchStats = async () => {
    setLoading(true)
    try {
      const response = await api.get('/admin/statistics')
      setStats(response.data)
    } catch (err) {
      alert(err.response?.data?.error || '获取统计失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStats()
  }, [])

  const exportByDate = async () => {
    if (!dateRange.start || !dateRange.end) {
      alert('请选择日期范围')
      return
    }
    window.open(`/api/admin/export/date-range?start_date=${dateRange.start}&end_date=${dateRange.end}`, '_blank')
  }

  const exportByExpert = async () => {
    window.open('/api/admin/export/by-expert', '_blank')
  }

  const getStatusText = (status) => {
    const map = {
      pending: '待接诊',
      in_progress: '问诊中',
      pending_supplement: '待补充',
      completed: '已完成',
    }
    return map[status] || status
  }

  return (
    <div className="container">
      <div className="header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>管理员面板</h1>
          <p style={{ fontSize: 14, opacity: 0.9 }}>欢迎，{user.name}</p>
        </div>
        <button className="btn btn-danger" onClick={onLogout}>退出登录</button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : stats && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="value">{stats.total_consultations}</div>
            <div className="label">总问诊数</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.pending_consultations}</div>
            <div className="label">待接诊</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.in_progress_consultations}</div>
            <div className="label">问诊中</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.completed_consultations}</div>
            <div className="label">已完成</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.total_doctors}</div>
            <div className="label">基层医生数</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.total_experts}</div>
            <div className="label">专家数</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.total_patients}</div>
            <div className="label">患者数</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.total_prescriptions}</div>
            <div className="label">处方数</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.average_response_time?.toFixed(2) || 0}h</div>
            <div className="label">平均响应时间</div>
          </div>
        </div>
      )}

      <div className="card">
        <h2>数据导出</h2>

        <div style={{ marginBottom: 20 }}>
          <h3 style={{ marginBottom: 10 }}>按日期范围导出问诊记录</h3>
          <div style={{ display: 'flex', gap: 15, alignItems: 'center', flexWrap: 'wrap' }}>
            <div className="form-group" style={{ marginBottom: 0 }}>
              <label>开始日期</label>
              <input
                type="date"
                value={dateRange.start}
                onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })}
              />
            </div>
            <div className="form-group" style={{ marginBottom: 0 }}>
              <label>结束日期</label>
              <input
                type="date"
                value={dateRange.end}
                onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })}
              />
            </div>
            <button className="btn btn-primary" onClick={exportByDate}>导出CSV</button>
          </div>
        </div>

        <div>
          <h3 style={{ marginBottom: 10 }}>按专家导出统计</h3>
          <button className="btn btn-success" onClick={exportByExpert}>导出专家统计CSV</button>
          <p style={{ fontSize: 12, color: '#888', marginTop: 5 }}>包含接诊数量、平均响应时间、完成率等</p>
        </div>
      </div>

      <div style={{ textAlign: 'center', marginTop: 20 }}>
        <button className="btn" onClick={fetchStats}>刷新数据</button>
      </div>
    </div>
  )
}

export default AdminDashboard
