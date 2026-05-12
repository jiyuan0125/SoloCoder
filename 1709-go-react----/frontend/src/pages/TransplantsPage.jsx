import React, { useState, useEffect } from 'react'
import { transplantsAPI, organsAPI, recipientsAPI, reportsAPI } from '../api'

const getPostOpStatusBadge = (status) => {
  const mapping = {
    '手术中': 'badge-moderate',
    '术后观察': 'badge-warning',
    '已出院': 'badge-success'
  }
  return mapping[status] || 'badge-pending'
}

const getSurgeryResultBadge = (result) => {
  const mapping = {
    '成功': 'badge-success',
    '失败': 'badge-danger',
    '并发症': 'badge-warning'
  }
  return mapping[result] || 'badge-pending'
}

function CreateTransplantModal({ onSubmit, onClose, organs, recipients }) {
  const [form, setForm] = useState({
    organ_id: organs[0]?.id || '',
    recipient_id: recipients[0]?.id || '',
    surgery_date: new Date().toISOString().split('T')[0],
    surgeon: '',
    result: '成功'
  })

  const matchedOrgans = organs.filter(o => o.status === '已匹配' || o.status === '已获取')

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      ...form,
      surgery_date: new Date(form.surgery_date).toISOString()
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>创建移植记录</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label>器官ID</label>
              <select value={form.organ_id}
                onChange={e => setForm({...form, organ_id: e.target.value})}>
                {matchedOrgans.map(o => (
                  <option key={o.id} value={o.id}>
                    {o.id.substring(0, 12)}... - {o.organ_type}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>受体ID</label>
              <select value={form.recipient_id}
                onChange={e => setForm({...form, recipient_id: e.target.value})}>
                {recipients.filter(r => r.matched_organ_id).map(r => (
                  <option key={r.id} value={r.id}>
                    {r.name} ({r.recipient_no})
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>手术日期</label>
              <input type="datetime-local" required
                value={form.surgery_date}
                onChange={e => setForm({...form, surgery_date: e.target.value})} />
            </div>
            <div className="form-group">
              <label>主刀医生</label>
              <input required value={form.surgeon}
                onChange={e => setForm({...form, surgeon: e.target.value})} />
            </div>
            <div className="form-group">
              <label>手术结果</label>
              <select value={form.result}
                onChange={e => setForm({...form, result: e.target.value})}>
                <option value="成功">成功</option>
                <option value="失败">失败</option>
                <option value="并发症">并发症</option>
              </select>
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-primary">创建</button>
          </div>
        </form>
      </div>
    </div>
  )
}

function FollowUpModal({ transplantId, onSubmit, onClose }) {
  const [form, setForm] = useState({
    follow_up_date: new Date().toISOString().split('T')[0],
    follow_up_type: '1个月随访',
    organ_function: '',
    recovery_status: ''
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      ...form,
      follow_up_date: new Date(form.follow_up_date).toISOString()
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>添加随访记录</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label>随访日期</label>
              <input type="date" required value={form.follow_up_date}
                onChange={e => setForm({...form, follow_up_date: e.target.value})} />
            </div>
            <div className="form-group">
              <label>随访类型</label>
              <select value={form.follow_up_type}
                onChange={e => setForm({...form, follow_up_type: e.target.value})}>
                <option value="1个月随访">1个月随访</option>
                <option value="3个月随访">3个月随访</option>
                <option value="6个月随访">6个月随访</option>
                <option value="1年随访">1年随访</option>
              </select>
            </div>
          </div>
          <div className="form-group">
            <label>器官功能状态</label>
            <textarea rows="3" required value={form.organ_function}
              onChange={e => setForm({...form, organ_function: e.target.value})} />
          </div>
          <div className="form-group">
            <label>患者恢复情况</label>
            <textarea rows="3" required value={form.recovery_status}
              onChange={e => setForm({...form, recovery_status: e.target.value})} />
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-primary">提交</button>
          </div>
        </form>
      </div>
    </div>
  )
}

function TransplantDetail({ transplant, onClose, onUpdateStatus, onAddFollowUp }) {
  const [showFollowUp, setShowFollowUp] = useState(false)

  const nextStatus = () => {
    switch (transplant.post_op_status) {
      case '手术中':
        return '术后观察'
      case '术后观察':
        return '已出院'
      default:
        return null
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{maxWidth: '800px'}} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>移植记录详情</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>

        <div className="detail-section">
          <h3>基本信息</h3>
          <div className="detail-grid">
            <div className="detail-item"><label>器官ID</label><div className="value">{transplant.organ_id.substring(0, 12)}...</div></div>
            <div className="detail-item"><label>受体ID</label><div className="value">{transplant.recipient_id.substring(0, 12)}...</div></div>
            <div className="detail-item"><label>手术日期</label><div className="value">{new Date(transplant.surgery_date).toLocaleString()}</div></div>
            <div className="detail-item"><label>主刀医生</label><div className="value">{transplant.surgeon}</div></div>
            <div className="detail-item"><label>手术结果</label><div className="value"><span className={`badge ${getSurgeryResultBadge(transplant.result)}`}>{transplant.result}</span></div></div>
            <div className="detail-item"><label>当前状态</label><div className="value"><span className={`badge ${getPostOpStatusBadge(transplant.post_op_status)}`}>{transplant.post_op_status}</span></div></div>
          </div>
        </div>

        {nextStatus() && (
          <div className="form-group" style={{marginTop: '16px'}}>
            <button className="btn btn-primary"
              onClick={() => onUpdateStatus(nextStatus())}>
              更新状态: {nextStatus()}
            </button>
          </div>
        )}

        <div className="detail-section">
          <h3>随访记录</h3>
          {transplant.follow_ups && transplant.follow_ups.length > 0 ? (
            <div className="table-container">
              <table>
                <thead>
                  <tr>
                    <th>日期</th>
                    <th>类型</th>
                    <th>器官功能</th>
                    <th>恢复情况</th>
                  </tr>
                </thead>
                <tbody>
                  {transplant.follow_ups.map(fu => (
                    <tr key={fu.id}>
                      <td>{new Date(fu.follow_up_date).toLocaleDateString()}</td>
                      <td>{fu.follow_up_type}</td>
                      <td>{fu.organ_function}</td>
                      <td>{fu.recovery_status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="empty-state" style={{padding: '20px'}}>
              <p>暂无随访记录</p>
            </div>
          )}
          <button className="btn btn-sm btn-success" onClick={() => setShowFollowUp(true)}>
            + 添加随访
          </button>
        </div>

        <div className="modal-footer">
          <button className="btn" onClick={onClose}>关闭</button>
        </div>

        {showFollowUp && (
          <div className="modal-overlay" onClick={() => setShowFollowUp(false)}>
            <div className="modal" onClick={e => e.stopPropagation()}>
              <FollowUpModal
                transplantId={transplant.id}
                onSubmit={(data) => {
                  onAddFollowUp(data)
                  setShowFollowUp(false)
                }}
                onClose={() => setShowFollowUp(false)}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default function TransplantsPage() {
  const [transplants, setTransplants] = useState([])
  const [organs, setOrgans] = useState([])
  const [recipients, setRecipients] = useState([])
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedTransplant, setSelectedTransplant] = useState(null)
  const [message, setMessage] = useState(null)

  const loadData = async () => {
    try {
      setLoading(true)
      const [transplantsData, recipientsData, reportsData] = await Promise.all([
        transplantsAPI.list(1, 50),
        recipientsAPI.list(1, 100),
        reportsAPI.list()
      ])
      setTransplants(transplantsData.data || [])
      setRecipients(recipientsData.data || [])
      setReports(reportsData)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleCreate = async (data) => {
    try {
      await transplantsAPI.create(data)
      setShowCreate(false)
      setMessage({ type: 'success', text: '移植记录创建成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleUpdateStatus = async (newStatus) => {
    try {
      await transplantsAPI.updateStatus(selectedTransplant.id, newStatus)
      const updated = await transplantsAPI.get(selectedTransplant.id)
      setSelectedTransplant(updated)
      setMessage({ type: 'success', text: '状态更新成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleAddFollowUp = async (data) => {
    try {
      await transplantsAPI.addFollowUp(selectedTransplant.id, data)
      const updated = await transplantsAPI.get(selectedTransplant.id)
      setSelectedTransplant(updated)
      setMessage({ type: 'success', text: '随访记录添加成功' })
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleGenerateReport = async () => {
    try {
      await reportsAPI.generate()
      const reportsData = await reportsAPI.list()
      setReports(reportsData)
      setMessage({ type: 'success', text: '报告生成成功' })
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  return (
    <div>
      <div className="page-header">
        <h1>移植记录管理</h1>
        <p>管理移植手术记录和术后随访</p>
      </div>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      {reports.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h2>统计报告</h2>
            <button className="btn btn-sm btn-primary" onClick={handleGenerateReport}>
              生成新报告
            </button>
          </div>
          <div className="stats-grid">
            {reports.slice(-3).reverse().map(report => (
            <div className="stat-card" key={report.id}>
              <div style={{fontSize: '12px', color: '#666', marginBottom: '8px'}}>
                {new Date(report.report_date).toLocaleDateString()}
              </div>
              <div className="detail-grid">
                <div className="detail-item"><label>新增捐献</label><div className="value">{report.new_donors}</div></div>
                <div className="detail-item"><label>成功匹配</label><div className="value">{report.successful_matches}</div></div>
                <div className="detail-item"><label>平均等待</label><div className="value">{report.average_wait_time.toFixed(1)}天</div></div>
              </div>
            </div>
            ))}
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-header">
          <h2>移植记录列表</h2>
          <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
            + 创建移植记录
          </button>
        </div>

        {loading ? (
          <div className="loading">加载中...</div>
        ) : transplants.length === 0 ? (
          <div className="empty-state">
            <h3>暂无移植记录</h3>
            <p>点击上方按钮创建新的移植记录</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
              <tr>
                <th>ID</th>
                <th>器官ID</th>
                <th>受体ID</th>
                <th>手术日期</th>
                <th>主刀医生</th>
                <th>手术结果</th>
                <th>当前状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {transplants.map(t => (
                <tr key={t.id}>
                  <td>{t.id.substring(0, 8)}...</td>
                  <td>{t.organ_id.substring(0, 8)}...</td>
                  <td>{t.recipient_id.substring(0, 8)}...</td>
                  <td>{new Date(t.surgery_date).toLocaleDateString()}</td>
                  <td>{t.surgeon}</td>
                  <td><span className={`badge ${getSurgeryResultBadge(t.result)}`}>{t.result}</span></td>
                  <td><span className={`badge ${getPostOpStatusBadge(t.post_op_status)}`}>{t.post_op_status}</span></td>
                  <td>
                    <button className="btn btn-sm btn-primary" onClick={async () => {
                      const detail = await transplantsAPI.get(t.id)
                      setSelectedTransplant(detail)
                    }}>详情</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      </div>

      {showCreate && (
        <CreateTransplantModal
          organs={organs}
          recipients={recipients}
          onSubmit={handleCreate}
          onClose={() => setShowCreate(false)}
        />
      )}

      {selectedTransplant && (
        <TransplantDetail
          transplant={selectedTransplant}
          onClose={() => setSelectedTransplant(null)}
          onUpdateStatus={handleUpdateStatus}
          onAddFollowUp={handleAddFollowUp}
        />
      )}
    </div>
  )
}
