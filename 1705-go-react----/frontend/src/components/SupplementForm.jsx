import React, { useState } from 'react'
import api from '../api.js'

function SupplementForm({ consultation, onSuccess, onClose }) {
  const [exams, setExams] = useState([])
  const [examInput, setExamInput] = useState({ exam_type: '', exam_result: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const addExam = () => {
    if (examInput.exam_type && examInput.exam_result) {
      setExams([...exams, examInput])
      setExamInput({ exam_type: '', exam_result: '' })
    }
  }

  const removeExam = (index) => {
    setExams(exams.filter((_, i) => i !== index))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (exams.length === 0) {
      setError('必须补充至少一项资料')
      return
    }

    setLoading(true)
    setError('')
    try {
      await api.post(`/consultations/${consultation.id}/supplement`, { exams })
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
          <h2>补充资料</h2>
          <button className="btn" onClick={onClose}>取消</button>
        </div>

        <p>问诊编号：{consultation.consultation_no}</p>
        <p>患者：{consultation.patient?.name}</p>

        <form onSubmit={handleSubmit} style={{ marginTop: 20 }}>
          <h3 style={{ marginBottom: 10, color: '#333' }}>补充检查资料</h3>
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
            {loading ? '提交中...' : '提交补充资料'}
          </button>
        </form>
      </div>
    </div>
  )
}

export default SupplementForm
