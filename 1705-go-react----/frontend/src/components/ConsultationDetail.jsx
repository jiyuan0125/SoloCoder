import React, { useState, useEffect, useCallback } from 'react'
import api from '../api.js'

function ConsultationDetail({ consultation, userRole, onClose, onRefresh }) {
  const [detail, setDetail] = useState(null)
  const [messages, setMessages] = useState([])
  const [prescriptions, setPrescriptions] = useState([])
  const [newMessage, setNewMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const fetchData = useCallback(async () => {
    try {
      const [detailRes, messagesRes, prescriptionsRes] = await Promise.all([
        api.get(`/consultations/${consultation.id}`),
        api.get(`/consultations/${consultation.id}/messages`),
        api.get(`/prescriptions/consultation/${consultation.id}`),
      ])
      setDetail(detailRes.data)
      setMessages(messagesRes.data)
      setPrescriptions(prescriptionsRes.data)
    } catch (err) {
      setError(err.response?.data?.error || '获取详情失败')
    }
  }, [consultation.id])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const sendMessage = async () => {
    if (!newMessage.trim()) return
    setLoading(true)
    setError('')
    try {
      await api.post(`/consultations/${consultation.id}/messages`, { content: newMessage })
      setNewMessage('')
      fetchData()
    } catch (err) {
      setError(err.response?.data?.error || '发送失败')
    } finally {
      setLoading(false)
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

  if (!detail) {
    return (
      <div className="card">
        <div className="loading">加载中...</div>
      </div>
    )
  }

  const canSendMessage = detail.status !== 'completed'

  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000, overflow: 'auto', padding: 20 }}>
      <div className="card" style={{ maxWidth: 900, margin: '0 auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2>问诊详情</h2>
          <button className="btn" onClick={onClose}>关闭</button>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          <div>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>基本信息</h3>
            <p><strong>问诊编号：</strong>{detail.consultation_no}</p>
            <p><strong>状态：</strong>{getStatusBadge(detail.status)}</p>
            <p><strong>基层医生：</strong>{detail.grassroot_doctor?.name}</p>
            <p><strong>接诊专家：</strong>{detail.expert_doctor?.name || '-'}</p>
            <p><strong>创建时间：</strong>{new Date(detail.created_at).toLocaleString('zh-CN')}</p>
          </div>
          <div>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>患者信息</h3>
            <p><strong>姓名：</strong>{detail.patient?.name}</p>
            <p><strong>性别：</strong>{detail.patient?.gender}</p>
            <p><strong>年龄：</strong>{detail.patient?.age}岁</p>
            <p><strong>电话：</strong>{detail.patient?.phone}</p>
          </div>
        </div>

        <div style={{ marginTop: 20 }}>
          <h3 style={{ marginBottom: 10, color: '#1976d2' }}>主诉症状</h3>
          <p style={{ background: '#f5f7fa', padding: 15, borderRadius: 4, whiteSpace: 'pre-wrap' }}>{detail.chief_complaint}</p>
        </div>

        {detail.past_history && (
          <div style={{ marginTop: 20 }}>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>既往病史</h3>
            <p style={{ background: '#f5f7fa', padding: 15, borderRadius: 4, whiteSpace: 'pre-wrap' }}>{detail.past_history}</p>
          </div>
        )}

        {detail.exams && detail.exams.length > 0 && (
          <div style={{ marginTop: 20 }}>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>检查记录</h3>
            {detail.exams.map((exam, i) => (
              <div key={i} style={{ background: '#f5f7fa', padding: 15, borderRadius: 4, marginBottom: 10 }}>
                <p><strong>{exam.exam_type}：</strong>{exam.exam_result}</p>
              </div>
            ))}
          </div>
        )}

        {detail.image_datas && detail.image_datas.length > 0 && (
          <div style={{ marginTop: 20 }}>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>影像资料</h3>
            {detail.image_datas.map((img, i) => (
              <div key={i} style={{ background: '#f5f7fa', padding: 15, borderRadius: 4, marginBottom: 10 }}>
                <p><strong>类型：</strong>{img.image_type}</p>
                <p><strong>模板：</strong>{img.template?.type}</p>
                <p><strong>内容：</strong></p>
                <pre style={{ background: 'white', padding: 10, borderRadius: 4, marginTop: 5, overflow: 'auto' }}>
                  {JSON.stringify(JSON.parse(img.field_values), null, 2)}
                </pre>
              </div>
            ))}
          </div>
        )}

        {detail.status === 'completed' && (
          <>
            {detail.diagnosis && (
              <div style={{ marginTop: 20 }}>
                <h3 style={{ marginBottom: 10, color: '#2e7d32' }}>诊断意见</h3>
                <p style={{ background: '#e8f5e9', padding: 15, borderRadius: 4, whiteSpace: 'pre-wrap' }}>{detail.diagnosis}</p>
              </div>
            )}
            {detail.prescription_advice && (
              <div style={{ marginTop: 20 }}>
                <h3 style={{ marginBottom: 10, color: '#2e7d32' }}>处方建议</h3>
                <p style={{ background: '#e8f5e9', padding: 15, borderRadius: 4, whiteSpace: 'pre-wrap' }}>{detail.prescription_advice}</p>
              </div>
            )}
            {detail.treatment_advice && (
              <div style={{ marginTop: 20 }}>
                <h3 style={{ marginBottom: 10, color: '#2e7d32' }}>治疗建议</h3>
                <p style={{ background: '#e8f5e9', padding: 15, borderRadius: 4, whiteSpace: 'pre-wrap' }}>{detail.treatment_advice}</p>
              </div>
            )}
          </>
        )}

        {prescriptions.length > 0 && (
          <div style={{ marginTop: 20 }}>
            <h3 style={{ marginBottom: 10, color: '#1976d2' }}>处方</h3>
            {prescriptions.map((p) => (
              <div key={p.id} style={{ background: '#f5f7fa', padding: 15, borderRadius: 4, marginBottom: 10 }}>
                <p><strong>处方编号：</strong>{p.prescription_no}</p>
                <p><strong>有效期至：</strong>{new Date(p.expired_at).toLocaleString('zh-CN')}</p>
                <table style={{ marginTop: 10 }}>
                  <thead>
                    <tr>
                      <th>药品名称</th>
                      <th>规格</th>
                      <th>用法</th>
                      <th>用量</th>
                      <th>天数</th>
                    </tr>
                  </thead>
                  <tbody>
                    {p.items.map((item) => (
                      <tr key={item.id}>
                        <td>{item.drug_name}</td>
                        <td>{item.specification}</td>
                        <td>{item.usage}</td>
                        <td>{item.dosage}</td>
                        <td>{item.days}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ))}
          </div>
        )}

        <div style={{ marginTop: 20 }}>
          <h3 style={{ marginBottom: 10, color: '#1976d2' }}>交流记录</h3>
          <div className="chat-box">
            {messages.length === 0 ? (
              <p style={{ color: '#888', textAlign: 'center' }}>暂无消息</p>
            ) : (
              messages.map((msg) => (
                <div key={msg.id} className={`message message-${msg.sender_role}`}>
                  <div className="message-meta">
                    {msg.sender_name} ({msg.sender_role === 'expert' ? '专家' : '基层医生'}) - {new Date(msg.created_at).toLocaleString('zh-CN')}
                  </div>
                  <div>{msg.content}</div>
                </div>
              ))
            )}
          </div>

          {canSendMessage && (
            <div style={{ display: 'flex', gap: 10 }}>
              <input
                style={{ flex: 1, padding: 10, border: '1px solid #ddd', borderRadius: 4 }}
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="输入消息..."
                onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
              />
              <button className="btn btn-primary" onClick={sendMessage} disabled={loading || !newMessage.trim()}>
                发送
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ConsultationDetail
