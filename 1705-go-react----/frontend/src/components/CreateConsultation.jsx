import React, { useState } from 'react'
import api from '../api.js'

function CreateConsultation({ onSuccess }) {
  const [formData, setFormData] = useState({
    patient_name: '',
    patient_gender: '男',
    patient_age: '',
    patient_id_card: '',
    patient_phone: '',
    chief_complaint: '',
    past_history: '',
  })
  const [exams, setExams] = useState([])
  const [examInput, setExamInput] = useState({ exam_type: '', exam_result: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')

    if ([...formData.chief_complaint].length < 50) {
      setError('主诉症状描述至少50字')
      return
    }

    setLoading(true)
    try {
      const data = { ...formData, patient_age: parseInt(formData.patient_age), exams }
      await api.post('/consultations', data)
      onSuccess()
    } catch (err) {
      setError(err.response?.data?.error || '创建失败')
    } finally {
      setLoading(false)
    }
  }

  const addExam = () => {
    if (examInput.exam_type && examInput.exam_result) {
      setExams([...exams, examInput])
      setExamInput({ exam_type: '', exam_result: '' })
    }
  }

  const removeExam = (index) => {
    setExams(exams.filter((_, i) => i !== index))
  }

  return (
    <div className="card">
      <h2>创建新问诊</h2>
      <form onSubmit={handleSubmit}>
        <h3 style={{ marginTop: 20, color: '#333' }}>患者基本信息</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 15 }}>
          <div className="form-group">
            <label>姓名</label>
            <input
              value={formData.patient_name}
              onChange={(e) => setFormData({ ...formData, patient_name: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>性别</label>
            <select
              value={formData.patient_gender}
              onChange={(e) => setFormData({ ...formData, patient_gender: e.target.value })}
            >
              <option value="男">男</option>
              <option value="女">女</option>
            </select>
          </div>
          <div className="form-group">
            <label>年龄</label>
            <input
              type="number"
              value={formData.patient_age}
              onChange={(e) => setFormData({ ...formData, patient_age: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>身份证号</label>
            <input
              value={formData.patient_id_card}
              onChange={(e) => setFormData({ ...formData, patient_id_card: e.target.value })}
              required
            />
          </div>
          <div className="form-group" style={{ gridColumn: '1 / -1' }}>
            <label>联系电话</label>
            <input
              value={formData.patient_phone}
              onChange={(e) => setFormData({ ...formData, patient_phone: e.target.value })}
              required
            />
          </div>
        </div>

        <h3 style={{ marginTop: 20, color: '#333' }}>症状信息</h3>
        <div className="form-group">
          <label>主诉症状描述（至少50字）</label>
          <textarea
            rows={5}
            value={formData.chief_complaint}
            onChange={(e) => setFormData({ ...formData, chief_complaint: e.target.value })}
            placeholder="请详细描述患者症状..."
            required
          />
          <p style={{ fontSize: 12, color: '#888', marginTop: 5 }}>
            已输入 {[...formData.chief_complaint].length} 字
          </p>
        </div>
        <div className="form-group">
          <label>既往病史</label>
          <textarea
            rows={3}
            value={formData.past_history}
            onChange={(e) => setFormData({ ...formData, past_history: e.target.value })}
            placeholder="请描述患者既往病史..."
          />
        </div>

        <h3 style={{ marginTop: 20, color: '#333' }}>已做检查</h3>
        {exams.length > 0 && (
          <div style={{ marginBottom: 15 }}>
            {exams.map((exam, i) => (
              <div key={i} style={{ background: '#f5f7fa', padding: 10, borderRadius: 4, marginBottom: 10, display: 'flex', justifyContent: 'space-between' }}>
                <span><strong>{exam.exam_type}:</strong> {exam.exam_result}</span>
                <button type="button" className="btn btn-danger" onClick={() => removeExam(i)}>删除</button>
              </div>
            ))}
          </div>
        )}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr auto', gap: 10, alignItems: 'end' }}>
          <div className="form-group">
            <label>检查类型</label>
            <input
              value={examInput.exam_type}
              onChange={(e) => setExamInput({ ...examInput, exam_type: e.target.value })}
              placeholder="如：血常规"
            />
          </div>
          <div className="form-group">
            <label>检查结果</label>
            <input
              value={examInput.exam_result}
              onChange={(e) => setExamInput({ ...examInput, exam_result: e.target.value })}
              placeholder="检查结果描述"
            />
          </div>
          <button type="button" className="btn btn-primary" onClick={addExam}>添加</button>
        </div>

        {error && <div className="alert alert-error" style={{ marginTop: 15 }}>{error}</div>}

        <button type="submit" className="btn btn-success" style={{ marginTop: 20 }} disabled={loading}>
          {loading ? '创建中...' : '提交问诊'}
        </button>
      </form>
    </div>
  )
}

export default CreateConsultation
