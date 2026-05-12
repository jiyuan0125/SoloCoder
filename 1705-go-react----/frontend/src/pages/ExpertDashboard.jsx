import React, { useState, useEffect, useCallback } from 'react'
import api from '../api.js'
import ConsultationDetail from '../components/ConsultationDetail.jsx'
import CompleteForm from '../components/CompleteForm.jsx'
import PrescriptionForm from '../components/PrescriptionForm.jsx'

function ExpertDashboard({ user, onLogout }) {
  const [activeTab, setActiveTab] = useState('pool')
  const [pool, setPool] = useState([])
  const [accepted, setAccepted] = useState([])
  const [selectedConsultation, setSelectedConsultation] = useState(null)
  const [showComplete, setShowComplete] = useState(false)
  const [showPrescription, setShowPrescription] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [poolLoading, setPoolLoading] = useState(false)

  const fetchPool = useCallback(async () => {
    setPoolLoading(true)
    try {
      const response = await api.get('/consultations/pool')
      setPool(response.data)
    } catch (err) {
      console.error('Pool fetch error:', err)
    } finally {
      setPoolLoading(false)
    }
  }, [])

  const fetchAccepted = useCallback(async () => {
    setLoading(true)
    try {
      const response = await api.get('/consultations')
      setAccepted(response.data)
    } catch (err) {
      setError(err.response?.data?.error || '获取列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchPool()
    fetchAccepted()
  }, [fetchPool, fetchAccepted])

  useEffect(() => {
    if (activeTab === 'pool') {
      const interval = setInterval(fetchPool, 30000)
      return () => clearInterval(interval)
    }
  }, [activeTab, fetchPool])

  const handleAccept = async (consultation) => {
    if (window.confirm('确定接诊问诊 ' + consultation.consultation_no + '？')) {
      try {
        await api.post('/consultations/' + consultation.id + '/accept')
        alert('接诊成功')
        fetchPool()
        fetchAccepted()
      } catch (err) {
        if (err.response?.status === 409) {
          alert(err.response?.data?.error || '该问诊已被其他专家接诊')
        } else {
          alert(err.response?.data?.error || '接诊失败')
        }
        fetchPool()
      }
    }
  }

  const handleRequestSupplement = async (consultation) => {
    if (window.confirm('确定要求基层医生补充资料？')) {
      try {
        await api.post('/consultations/' + consultation.id + '/request-supplement')
        alert('已发送补充要求')
        fetchAccepted()
      } catch (err) {
        alert(err.response?.data?.error || '操作失败')
      }
    }
  }

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
          <h1>专家工作台</h1>
          <p style={{ fontSize: 14, opacity: 0.9 }}>欢迎，{user.name}</p>
        </div>
        <button className="btn btn-danger" onClick={onLogout}>退出登录</button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="nav-tabs">
        <button className={activeTab === 'pool' ? 'nav-tab active' : 'nav-tab'} onClick={() => setActiveTab('pool')}>
          候诊池
        </button>
        <button className={activeTab === 'accepted' ? 'nav-tab active' : 'nav-tab'} onClick={() => setActiveTab('accepted')}>
          已接诊
        </button>
      </div>

      {activeTab === 'pool' && (
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 15 }}>
            <h2>候诊池（每30秒自动刷新）</h2>
            <span style={{ fontSize: 12, color: '#888' }}>共 {pool.length} 条待接诊问诊</span>
          </div>
          {poolLoading ? (
            <div className="loading">加载中...</div>
          ) : pool.length === 0 ? (
            <p style={{ color: '#888', textAlign: 'center', padding: 20 }}>暂无待接诊问诊</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>问诊编号</th>
                  <th>患者</th>
                  <th>基层医生</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {pool.map((c) => (
                  <tr key={c.id}>
                    <td>{c.consultation_no}</td>
                    <td>{c.patient?.name} ({c.patient?.gender}，{c.patient?.age}岁)</td>
                    <td>{c.grassroot_doctor?.name}</td>
                    <td>{new Date(c.created_at).toLocaleString('zh-CN')}</td>
                    <td>
                      <button className="btn btn-primary" onClick={() => handleAccept(c)}>接诊</button>
                      <button className="btn" style={{ marginLeft: 10 }} onClick={() => setSelectedConsultation(c)}>查看</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'accepted' && (
        <div>
          {loading ? (
            <div className="loading">加载中...</div>
          ) : (
            <div className="card">
              <h2>已接诊问诊</h2>
              {accepted.length === 0 ? (
                <p style={{ color: '#888', textAlign: 'center', padding: 20 }}>暂无已接诊问诊</p>
              ) : (
                <table>
                  <thead>
                    <tr>
                      <th>问诊编号</th>
                      <th>患者</th>
                      <th>状态</th>
                      <th>创建时间</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {accepted.map((c) => (
                      <tr key={c.id}>
                        <td>{c.consultation_no}</td>
                        <td>{c.patient?.name}</td>
                        <td>{getStatusBadge(c.status)}</td>
                        <td>{new Date(c.created_at).toLocaleString('zh-CN')}</td>
                        <td>
                          <button className="btn btn-primary" onClick={() => setSelectedConsultation(c)}>查看</button>
                          {c.status === 'in_progress' && (
                            <>
                              <button className="btn btn-success" style={{ marginLeft: 10 }} onClick={() => { setSelectedConsultation(c); setShowComplete(true); }}>完成</button>
                              <button className="btn btn-warning" style={{ marginLeft: 10 }} onClick={() => handleRequestSupplement(c)}>要求补充</button>
                            </>
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

      {selectedConsultation && !showComplete && (
        <ConsultationDetail
          consultation={selectedConsultation}
          userRole={user.role}
          onClose={() => setSelectedConsultation(null)}
          onRefresh={fetchAccepted}
        />
      )}

      {selectedConsultation && showComplete && (
        <CompleteForm
          consultation={selectedConsultation}
          onSuccess={() => {
            setSelectedConsultation(null);
            setShowComplete(false);
            fetchAccepted();
          }}
          onClose={() => {
            setShowComplete(false);
          }}
          onPrescription={() => {
            setShowComplete(false);
            setShowPrescription(true);
          }}
        />
      )}

      {selectedConsultation && showPrescription && (
        <PrescriptionForm
          consultation={selectedConsultation}
          onSuccess={() => {
            setSelectedConsultation(null);
            setShowPrescription(false);
            fetchAccepted();
          }}
          onClose={() => setShowPrescription(false)}
        />
      )}
    </div>
  )
}

export default ExpertDashboard
