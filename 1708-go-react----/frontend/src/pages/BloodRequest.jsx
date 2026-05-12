import { useState, useEffect } from 'react'
import axios from 'axios'

export default function BloodRequest() {
  const [formData, setFormData] = useState({
    hospital_name: '',
    hospital_id: '',
    blood_type: '',
    product_type: 'Whole_Blood',
    quantity: 1,
    urgency: 'Regular',
    patient_info: '',
    notes: ''
  })
  const [requests, setRequests] = useState([])
  const [message, setMessage] = useState(null)
  const [error, setError] = useState(null)
  const [activeTab, setActiveTab] = useState('new')

  const bloodTypes = [
    'A_Positive', 'A_Negative',
    'B_Positive', 'B_Negative',
    'AB_Positive', 'AB_Negative',
    'O_Positive', 'O_Negative'
  ]

  const productTypes = [
    { value: 'Whole_Blood', label: '全血' },
    { value: 'Red_Blood_Cells', label: '红细胞悬液' },
    { value: 'Fresh_Frozen_Plasma', label: '新鲜冰冻血浆' },
    { value: 'Platelets', label: '机采血小板' }
  ]

  const urgencyLevels = [
    { value: 'Regular', label: '常规 (24小时内)' },
    { value: 'Urgent', label: '紧急 (4小时内)' },
    { value: 'Special_Urgent', label: '特急 (1小时内)' }
  ]

  useEffect(() => {
    fetchRequests()
  }, [])

  const fetchRequests = async () => {
    try {
      const res = await axios.get('/api/requests')
      setRequests(res.data || [])
    } catch (err) {
      console.error('Failed to fetch requests:', err)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)
    setMessage(null)

    if (!formData.blood_type) {
      setError('请选择血型')
      return
    }

    try {
      const res = await axios.post('/api/requests', formData)
      setMessage(`申请提交成功！状态: ${res.data.status === 'Issued' ? '已分配' : '已排队'}`)
      setFormData({
        hospital_name: '',
        hospital_id: '',
        blood_type: '',
        product_type: 'Whole_Blood',
        quantity: 1,
        urgency: 'Regular',
        patient_info: '',
        notes: ''
      })
      fetchRequests()
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error)
      } else {
        setError('提交失败，请重试')
      }
    }
  }

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'Pending': return <span className="status-badge status-pending">待处理</span>
      case 'Queued': return <span className="status-badge status-pending">排队中</span>
      case 'Matched': return <span className="status-badge status-testing">已匹配</span>
      case 'Issued': return <span className="status-badge status-qualified">已出库</span>
      case 'Rejected': return <span className="status-badge status-scrapped">已拒绝</span>
      default: return <span className="status-badge status-pending">{status}</span>
    }
  }

  const getUrgencyBadge = (urgency) => {
    switch (urgency) {
      case 'Regular': return <span className="status-badge status-instock">常规</span>
      case 'Urgent': return <span className="status-badge status-pending">紧急</span>
      case 'Special_Urgent': return <span className="status-badge status-scrapped">特急 ⚠️</span>
      default: return <span className="status-badge status-pending">{urgency}</span>
    }
  }

  const getProductTypeName = (type) => {
    const pt = productTypes.find(p => p.value === type)
    return pt ? pt.label : type
  }

  return (
    <div className="page">
      <h1 className="page-title">🏥 用血申请</h1>

      {error && <div className="alert alert-danger">{error}</div>}
      {message && <div className="alert alert-success">{message}</div>}

      <div style={{ marginBottom: '2rem' }}>
        <button
          className={`btn ${activeTab === 'new' ? 'btn-primary' : ''}`}
          onClick={() => setActiveTab('new')}
          style={{ marginRight: '1rem' }}
        >
          新建申请
        </button>
        <button
          className={`btn ${activeTab === 'list' ? 'btn-primary' : ''}`}
          onClick={() => setActiveTab('list')}
        >
          申请列表
        </button>
      </div>

      {activeTab === 'new' && (
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">医院名称 *</label>
              <input
                type="text"
                className="form-input"
                name="hospital_name"
                value={formData.hospital_name}
                onChange={handleChange}
                required
                placeholder="请输入医院名称"
              />
            </div>
            <div className="form-group">
              <label className="form-label">医院编号</label>
              <input
                type="text"
                className="form-input"
                name="hospital_id"
                value={formData.hospital_id}
                onChange={handleChange}
                placeholder="请输入医院编号"
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label className="form-label">血型 *</label>
              <select
                className="form-select"
                name="blood_type"
                value={formData.blood_type}
                onChange={handleChange}
                required
              >
                <option value="">请选择血型</option>
                {bloodTypes.map(bt => (
                  <option key={bt} value={bt}>{bt.replace('_', ' ')}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">血制品类型 *</label>
              <select
                className="form-select"
                name="product_type"
                value={formData.product_type}
                onChange={handleChange}
                required
              >
                {productTypes.map(pt => (
                  <option key={pt.value} value={pt.value}>{pt.label}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label className="form-label">数量 (袋) *</label>
              <input
                type="number"
                className="form-input"
                name="quantity"
                value={formData.quantity}
                onChange={handleChange}
                min="1"
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label">紧急程度 *</label>
              <select
                className="form-select"
                name="urgency"
                value={formData.urgency}
                onChange={handleChange}
                required
              >
                {urgencyLevels.map(urg => (
                  <option key={urg.value} value={urg.value}>{urg.label}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">患者信息</label>
            <textarea
              className="form-input"
              name="patient_info"
              value={formData.patient_info}
              onChange={handleChange}
              rows="2"
              placeholder="请输入患者信息"
            />
          </div>

          <div className="form-group">
            <label className="form-label">备注</label>
            <textarea
              className="form-input"
              name="notes"
              value={formData.notes}
              onChange={handleChange}
              rows="2"
              placeholder="请输入备注信息"
            />
          </div>

          <button type="submit" className="btn btn-primary" style={{ marginTop: '1rem' }}>
            提交申请
          </button>
        </form>
      )}

      {activeTab === 'list' && (
        <div>
          <h2 style={{ marginBottom: '1rem' }}>申请列表</h2>
          {requests.length === 0 ? (
            <p>暂无申请记录</p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>医院</th>
                  <th>血型</th>
                  <th>血制品</th>
                  <th>数量</th>
                  <th>紧急程度</th>
                  <th>状态</th>
                  <th>申请时间</th>
                </tr>
              </thead>
              <tbody>
                {requests.map(req => (
                  <tr key={req.id}>
                    <td>{req.hospital_name}</td>
                    <td style={{ fontWeight: 'bold' }}>{req.blood_type.replace('_', ' ')}</td>
                    <td>{getProductTypeName(req.product_type)}</td>
                    <td>{req.quantity} 袋</td>
                    <td>{getUrgencyBadge(req.urgency)}</td>
                    <td>{getStatusBadge(req.status)}</td>
                    <td>{new Date(req.created_at).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  )
}
