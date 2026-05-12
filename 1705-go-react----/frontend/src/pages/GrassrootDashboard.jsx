import React, { useState, useEffect } from 'react'
import api from '../api.js'
import CreateConsultation from '../components/CreateConsultation.jsx'
import ConsultationDetail from '../components/ConsultationDetail.jsx'
import SupplementForm from '../components/SupplementForm.jsx'

function GrassrootDashboard({ user, onLogout }) {
  const [activeTab, setActiveTab] = useState('list')
  const [consultations, setConsultations] = useState([])
  const [selectedConsultation, setSelectedConsultation] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const fetchConsultations = async () => {
    setLoading(true)
    try {
      const response = await api.get('/consultations')
      setConsultations(response.data)
    } catch (err) {
      setError(err.response?.data?.error || '获取列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConsultations()
  }, [])

  const getStatusBadge = (status) => {
    const badges = {
      pending: <span className="badge badge-pending">待接诊</span>,
      in_progress: <span className="badge badge-in-progress">问诊中</span>,
      pending_supplement: <span className="badge badge-pending-supplement">待补充</span>,
      completed: <span className="badge badge-completed">已完成</span>,
    }
    return badges[status] || status
  }

  return (
    <div className="container">
      <div className="header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>基层医生工作台</h1>
          <p style={{ fontSize: 14, opacity: 0.9 }}>欢迎，{user.name}</p>
        </div>
        <button className="btn btn-danger" onClick={onLogout}>退出登录</button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="nav-tabs">
        <button className={`nav-tab ${activeTab === 'list' ? 'active' : ''}`} onClick={() => setActiveTab('list')}>
          我的问诊
        </button>
        <button className={`nav-tab ${activeTab === 'create' ? 'active' : ''}`} onClick={() => setActiveTab('create')}>
          创建问诊
        </button>
      </div>

      {activeTab === 'list' && (
        <div>
          {loading ? (
            <div className="loading">加载中...</div>
          ) : (
            <div className="card">
              <h2>问诊列表</h2>
              {consultations.length === 0 ? (
                <p style={{ color: '#888', textAlign: 'center', padding: 20 }}>暂无问诊记录</p>
              ) : (
                <table>
                  <thead>
                    <tr>
                      <th>问诊编号</th>
                      <th>患者</th>
                      <th>接诊专家</th>
                      <th>状态</th>
                      <th>创建时间</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {consultations.map((c) => (
                      <tr key={c.id}>
                        <td>{c.consultation_no}</td>
                        <td>{c.patient?.name}</td>
                        <td>{c.expert_doctor?.name || '-'}</td>
                        <td>{getStatusBadge(c.status)}</td>
                        <td>{new Date(c.created_at).toLocaleString('zh-CN')}</td>
                        <td>
                          <button className="btn btn-primary" onClick={() => setSelectedConsultation(c)}>
                            查看
                          </button>
                          {c.status === 'pending_supplement' && (
                            <button className="btn btn-warning" onClick={() => setSelectedConsultation({ ...c, _showSupplement: true })}>
                              补充资料
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}
        </div>
      )}

      {activeTab === 'create' && (
        <CreateConsultation
          onSuccess={() => {
            setActiveTab('list')
            fetchConsultations()
          }}
        />
      )}

      {selectedConsultation && !selectedConsultation._showSupplement && (
        <ConsultationDetail
          consultation={selectedConsultation}
          userRole={user.role}
          onClose={() => setSelectedConsultation(null)}
          onRefresh={fetchConsultations}
        />
      )}

      {selectedConsultation?._showSupplement && (
        <SupplementForm
          consultation={selectedConsultation}
          onSuccess={() => {
            setSelectedConsultation(null)
            fetchConsultations()
          }}
          onClose={() => setSelectedConsultation(null)}
        />
      )}
    </div>
  )
}

export default GrassrootDashboard
