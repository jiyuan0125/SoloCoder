import React, { useState, useEffect } from 'react'
import { recipientsAPI, organsAPI, locationsAPI } from '../api'

const organTypes = [
  { value: 'heart', label: '心脏' },
  { value: 'liver', label: '肝脏' },
  { value: 'kidney_left', label: '左肾' },
  { value: 'kidney_right', label: '右肾' },
  { value: 'lung_left', label: '左肺' },
  { value: 'lung_right', label: '右肺' },
  { value: 'pancreas', label: '胰腺' },
  { value: 'cornea', label: '角膜' },
  { value: 'intestine', label: '小肠' }
]

const getUrgencyBadge = (level) => {
  const mapping = {
    '普通': 'badge-normal',
    '较急': 'badge-moderate',
    '紧急': 'badge-emergency'
  }
  return mapping[level] || 'badge-normal'
}

const getUrgencyWeight = (level) => {
  switch (level) {
    case '紧急': return 3
    case '较急': return 2
    default: return 1
  }
}

const getOrganTypeLabel = (type) => {
  const organ = organTypes.find(o => o.value === type)
  return organ ? organ.label : type
}

function RecipientForm({ onSubmit, onClose, locations }) {
  const [form, setForm] = useState({
    recipient_no: '',
    name: '',
    blood_type: 'A',
    organ_needed: 'heart',
    registration_date: new Date().toISOString().split('T')[0],
    urgency_level: '普通',
    hla: '',
    pra: '',
    age: '',
    location_id: locations[0]?.id || ''
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      ...form,
      pra: parseInt(form.pra),
      age: parseInt(form.age),
      registration_date: new Date(form.registration_date).toISOString()
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>登记受体</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label>受体编号</label>
              <input required value={form.recipient_no}
                onChange={e => setForm({...form, recipient_no: e.target.value})} />
            </div>
            <div className="form-group">
              <label>姓名</label>
              <input required value={form.name}
                onChange={e => setForm({...form, name: e.target.value})} />
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>血型</label>
              <select value={form.blood_type}
                onChange={e => setForm({...form, blood_type: e.target.value})}>
                <option value="A">A型</option>
                <option value="B">B型</option>
                <option value="AB">AB型</option>
                <option value="O">O型</option>
              </select>
            </div>
            <div className="form-group">
              <label>需要器官</label>
              <select value={form.organ_needed}
                onChange={e => setForm({...form, organ_needed: e.target.value})}>
                {organTypes.map(o => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>年龄</label>
              <input type="number" required min="0" value={form.age}
                onChange={e => setForm({...form, age: e.target.value})} />
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>登记日期</label>
              <input type="date" required value={form.registration_date}
                onChange={e => setForm({...form, registration_date: e.target.value})} />
            </div>
            <div className="form-group">
              <label>紧急程度</label>
              <select value={form.urgency_level}
                onChange={e => setForm({...form, urgency_level: e.target.value})}>
                <option value="普通">普通</option>
                <option value="较急">较急</option>
                <option value="紧急">紧急</option>
              </select>
            </div>
            <div className="form-group">
              <label>PRA值 (0-100)</label>
              <input type="number" required min="0" max="100" value={form.pra}
                onChange={e => setForm({...form, pra: e.target.value})} />
            </div>
          </div>
          <div className="form-group">
            <label>HLA分型 (如: A1,A2,B7,B8,DR1,DR3)</label>
            <input required value={form.hla}
              onChange={e => setForm({...form, hla: e.target.value})} />
          </div>
          <div className="form-group">
            <label>位置</label>
            <select value={form.location_id}
              onChange={e => setForm({...form, location_id: e.target.value})}>
              {locations.map(loc => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-primary">登记</button>
          </div>
        </form>
      </div>
    </div>
  )
}

function ConfirmMatchModal({ organId, recipients, onSubmit, onClose }) {
  const [selectedRecipient, setSelectedRecipient] = useState(recipients[0]?.id || '')

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      organ_id: organId,
      recipient_id: selectedRecipient,
      confirmed: true
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>确认匹配</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>选择受体</label>
            <select value={selectedRecipient}
              onChange={e => setSelectedRecipient(e.target.value)}>
              {recipients.map(r => (
                <option key={r.id} value={r.id}>
                  {r.name} ({r.recipient_no}) - 血型{r.blood_type}型 - PRA:{r.pra}
                </option>
              ))}
            </select>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-success">确认匹配</button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function RecipientsPage() {
  const [recipients, setRecipients] = useState([])
  const [locations, setLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [sortBy, setSortBy] = useState('urgency')
  const [sortOrder, setSortOrder] = useState('desc')
  const [message, setMessage] = useState(null)
  const [confirmData, setConfirmData] = useState(null)

  const loadData = async () => {
    try {
      setLoading(true)
      const [recipientsData, locationsData] = await Promise.all([
        recipientsAPI.list(1, 100),
        locationsAPI.list()
      ])
      setRecipients(recipientsData.data || [])
      setLocations(locationsData)
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const sortedRecipients = [...recipients].sort((a, b) => {
    let comparison = 0
    if (sortBy === 'urgency') {
      comparison = getUrgencyWeight(a.urgency_level) - getUrgencyWeight(b.urgency_level)
    } else if (sortBy === 'waitTime') {
      comparison = new Date(a.registration_date) - new Date(b.registration_date)
    }
    return sortOrder === 'desc' ? -comparison : comparison
  })

  const handleSort = (field) => {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc')
    } else {
      setSortBy(field)
      setSortOrder('desc')
    }
  }

  const handleCreateRecipient = async (data) => {
    try {
      await recipientsAPI.create(data)
      setShowForm(false)
      setMessage({ type: 'success', text: '受体登记成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleConfirmMatch = async (data) => {
    try {
      await organsAPI.confirmMatch(data.organ_id, {
        recipient_id: data.recipient_id,
        confirmed: data.confirmed
      })
      setConfirmData(null)
      setMessage({ type: 'success', text: '匹配确认成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const waitingRecipients = sortedRecipients.filter(r => !r.matched_organ_id)
  const matchedRecipients = sortedRecipients.filter(r => r.matched_organ_id)

  return (
    <div>
      <div className="page-header">
        <h1>受体等待队列</h1>
        <p>管理等待器官移植的受体，支持匹配过程展示</p>
      </div>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      <div className="stats-grid">
        <div className="stat-card">
          <h3>等待中</h3>
          <div className="value">{waitingRecipients.length}</div>
        </div>
        <div className="stat-card">
          <h3>已匹配</h3>
          <div className="value">{matchedRecipients.length}</div>
        </div>
        <div className="stat-card">
          <h3>紧急患者</h3>
          <div className="value">{waitingRecipients.filter(r => r.urgency_level === '紧急').length}</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h2>等待队列</h2>
          <button className="btn btn-primary" onClick={() => setShowForm(true)}>
            + 登记受体
          </button>
        </div>

        {loading ? (
          <div className="loading">加载中...</div>
        ) : waitingRecipients.length === 0 ? (
          <div className="empty-state">
            <h3>暂无等待中的受体</h3>
            <p>点击上方按钮登记新的受体</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>编号</th>
                  <th>姓名</th>
                  <th>血型</th>
                  <th>需要器官</th>
                  <th onClick={() => handleSort('urgency')} style={{cursor: 'pointer'}}>
                    紧急程度 {sortBy === 'urgency' && (sortOrder === 'desc' ? '↓' : '↑')}
                  </th>
                  <th>PRA</th>
                  <th onClick={() => handleSort('waitTime')} style={{cursor: 'pointer'}}>
                    登记日期 {sortBy === 'waitTime' && (sortOrder === 'desc' ? '↓' : '↑')}
                  </th>
                  <th>年龄</th>
                </tr>
              </thead>
              <tbody>
                {waitingRecipients.map(recipient => (
                  <tr key={recipient.id} className={recipient.is_matching ? 'highlight-row' : ''}>
                    <td>{recipient.recipient_no}</td>
                    <td>{recipient.name} {recipient.is_matching && <span className="badge badge-warning">匹配中</span>}</td>
                    <td>{recipient.blood_type}型</td>
                    <td>{getOrganTypeLabel(recipient.organ_needed)}</td>
                    <td><span className={`badge ${getUrgencyBadge(recipient.urgency_level)}`}>{recipient.urgency_level}</span></td>
                    <td>{recipient.pra}</td>
                    <td>{new Date(recipient.registration_date).toLocaleDateString()}</td>
                    <td>{recipient.age}岁</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {matchedRecipients.length > 0 && (
        <div className="card">
          <div className="card-header">
            <h2>已匹配受体</h2>
          </div>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>编号</th>
                  <th>姓名</th>
                  <th>血型</th>
                  <th>需要器官</th>
                  <th>匹配器官ID</th>
                </tr>
              </thead>
              <tbody>
                {matchedRecipients.map(recipient => (
                  <tr key={recipient.id}>
                    <td>{recipient.recipient_no}</td>
                    <td>{recipient.name}</td>
                    <td>{recipient.blood_type}型</td>
                    <td>{getOrganTypeLabel(recipient.organ_needed)}</td>
                    <td>{recipient.matched_organ_id?.substring(0, 8)}...</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showForm && (
        <RecipientForm
          locations={locations}
          onSubmit={handleCreateRecipient}
          onClose={() => setShowForm(false)}
        />
      )}

      {confirmData && (
        <ConfirmMatchModal
          organId={confirmData.organId}
          recipients={confirmData.recipients}
          onSubmit={handleConfirmMatch}
          onClose={() => setConfirmData(null)}
        />
      )}
    </div>
  )
}
