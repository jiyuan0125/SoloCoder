import { useState } from 'react'
import axios from 'axios'

export default function DonorRegistration() {
  const [formData, setFormData] = useState({
    name: '',
    id_card: '',
    gender: '',
    birth_date: '',
    abo_blood_type: '',
    rh_factor: 'Positive',
    phone: '',
    address: '',
    donation_type: 'Whole_400ml',
    health_check: {
      height_cm: '',
      weight_kg: '',
      recent_medication: '',
      is_fasting: true,
      notes: ''
    }
  })
  const [message, setMessage] = useState(null)
  const [error, setError] = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)
    setMessage(null)
    
    try {
      const data = {
        ...formData,
        birth_date: new Date(formData.birth_date)
      }
      const res = await axios.post('/api/donors', data)
      setMessage(`献血者注册成功！献血者编号: ${res.data.donor_number}`)
      setFormData({
        name: '',
        id_card: '',
        gender: '',
        birth_date: '',
        abo_blood_type: '',
        rh_factor: 'Positive',
        phone: '',
        address: '',
        donation_type: 'Whole_400ml',
        health_check: {
          height_cm: '',
          weight_kg: '',
          recent_medication: '',
          is_fasting: true,
          notes: ''
        }
      })
    } catch (err) {
      if (err.response) {
        if (err.response.data?.error) {
          setError(err.response.data.error)
        } else {
            setError('注册失败，请重试')
          }
        }
      }
    }

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    if (name.startsWith('health_check.')) {
      const field = name.split('.')[1]
      setFormData(prev => ({
        ...prev,
        health_check: {
          ...prev.health_check,
          [field]: type === 'checkbox' ? checked : value
        }
      }))
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value
      }))
    }
  }

  return (
    <div className="page">
      <h1 className="page-title">🩸 献血者登记</h1>

      {error && <div className="alert alert-danger">{error}</div>}
      {message && <div className="alert alert-success">{message}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label className="form-label">姓名 *</label>
            <input
              type="text"
              className="form-input"
              name="name"
              value={formData.name}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label className="form-label">身份证号 *</label>
            <input
              type="text"
              className="form-input"
              name="id_card"
              value={formData.id_card}
              onChange={handleChange}
              required
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">性别 *</label>
            <select
              className="form-select"
              name="gender"
              value={formData.gender}
              onChange={handleChange}
              required
            >
              <option value="">请选择</option>
              <option value="Male">男</option>
              <option value="Female">女</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">出生日期 *</label>
            <input
              type="date"
              className="form-input"
              name="birth_date"
              value={formData.birth_date}
              onChange={handleChange}
              required
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">ABO血型 *</label>
            <select
              className="form-select"
              name="abo_blood_type"
              value={formData.abo_blood_type}
              onChange={handleChange}
              required
            >
              <option value="">请选择</option>
              <option value="A">A型</option>
              <option value="B">B型</option>
              <option value="AB">AB型</option>
              <option value="O">O型</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Rh因子 *</label>
            <select
              className="form-select"
              name="rh_factor"
              value={formData.rh_factor}
              onChange={handleChange}
              required
            >
              <option value="Positive">阳性</option>
              <option value="Negative">阴性</option>
            </select>
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label className="form-label">联系电话</label>
            <input
              type="tel"
              className="form-input"
              name="phone"
              value={formData.phone}
              onChange={handleChange}
            />
          </div>
          <div className="form-group">
            <label className="form-label">献血类型 *</label>
            <select
              className="form-select"
              name="donation_type"
              value={formData.donation_type}
              onChange={handleChange}
              required
            >
              <option value="Whole_200ml">全血 200ml</option>
              <option value="Whole_400ml">全血 400ml</option>
              <option value="Platelets">机采血小板</option>
            </select>
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">住址</label>
          <textarea
            className="form-input"
            name="address"
            value={formData.address}
            onChange={handleChange}
            rows="2"
          />
        </div>

        <div className="section">
          <h2 style={{ marginBottom: '1rem' }}>健康征询</h2>
          
          <div className="form-row">
            <div className="form-group">
              <label className="form-label">身高 (cm)</label>
              <input
                type="number"
                className="form-input"
                name="health_check.height_cm"
                value={formData.health_check.height_cm}
                onChange={handleChange}
              />
            </div>
            <div className="form-group">
              <label className="form-label">体重 (kg)</label>
              <input
                type="number"
                className="form-input"
                name="health_check.weight_kg"
                value={formData.health_check.weight_kg}
                onChange={handleChange}
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">近期用药情况</label>
            <input
              type="text"
              className="form-input"
              name="health_check.recent_medication"
              value={formData.health_check.recent_medication}
              onChange={handleChange}
            />
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                name="health_check.is_fasting"
                checked={formData.health_check.is_fasting}
                onChange={handleChange}
              />
              是否空腹
            </label>
          </div>
        </div>

        <button type="submit" className="btn btn-primary" style={{ marginTop: '1rem' }}>
          注册献血者
        </button>
      </form>
    </div>
  )
}
