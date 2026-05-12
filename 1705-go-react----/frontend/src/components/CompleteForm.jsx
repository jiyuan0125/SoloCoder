import React, { useState } from 'react'
import api from '../api.js'

function CompleteForm({ consultation, onSuccess, onClose }) {
  const [formData, setFormData] = useState({
    diagnosis: '',
    prescription_advice: '',
    treatment_advice: '',
  })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!formData.diagnosis.trim() || !formData.prescription_advice.trim()) {
      setError('诊断意见和处方建议不能为空')
      return
    }

    setLoading(true)
    setError('')
    try {
      await api.post(`/consultations/${consultation.id}/complete`, formData)
      alert('问诊完成')
      onSuccess()
    } catch (err) {
      setError(err.response?.data?.error || '提交失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000, overflow: 'auto', padding: 20 }}>
      <div className="card" style={{ maxWidth: 700, margin: '0 auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2>完成问诊</h2>
          <button className="btn" onClick={onClose}>取消</button>
        </div>

        <p>问诊编号：{consultation.consultation_no}</p>
        <p>患者：{consultation.patient?.name}</p>

        <form onSubmit={handleSubmit} style={{ marginTop: 20 }}>
          <div className="form-group">
            <label>诊断意见 *</label>
            <textarea
              rows={4}
              value={formData.diagnosis}
              onChange={(e) => setFormData({ ...formData, diagnosis: e.target.value })}
              placeholder="请填写诊断意见..."
              required
            />
          </div>
          <div className="form-group">
            <label>处方建议 *</label>
            <textarea
              rows={4}
              value={formData.prescription_advice}
              onChange={(e) => setFormData({ ...formData, prescription_advice: e.target.value })}
              placeholder="请填写处方建议..."
              required
            />
          </div>
          <div className="form-group">
            <label>治疗建议</label>
            <textarea
              rows={4}
              value={formData.treatment_advice}
              onChange={(e) => setFormData({ ...formData, treatment_advice: e.target.value })}
              placeholder="请填写治疗建议..."
            />
          </div>

          {error && <div className="alert alert-error">{error}</div>}

          <button type="submit" className="btn btn-success" disabled={loading}>
            {loading ? '提交中...' : '提交完成'}
          </button>
        </form>
      </div>
    </div>
  )
}

export default CompleteForm
