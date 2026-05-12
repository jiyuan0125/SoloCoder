import React, { useState, useEffect } from 'react'
import api from '../api.js'

function PrescriptionForm({ consultation, onSuccess, onClose }) {
  const [items, setItems] = useState([{ drug_name: '', specification: '', usage: '', dosage: '', days: 1, unit_price: 0, quantity: 1 }])
  const [forbiddenDrugs, setForbiddenDrugs] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [totalAmount, setTotalAmount] = useState(0)

  useEffect(() => {
    api.get('/forbidden-drugs').then((res) => setForbiddenDrugs(res.data))
  }, [])

  useEffect(() => {
    const total = items.reduce((sum, item) => {
      if (item.unit_price && item.quantity) {
        return sum + (item.unit_price * item.quantity)
      }
      return sum
    }, 0)
    setTotalAmount(total)
  }, [items])

  const addItem = () => {
    if (items.length < 5) {
      setItems([...items, { drug_name: '', specification: '', usage: '', dosage: '', days: 1, unit_price: 0, quantity: 1 }])
    } else {
      alert('每个处方最多5种药品')
    }
  }

  const removeItem = (index) => {
    if (items.length > 1) {
      setItems(items.filter((_, i) => i !== index))
    }
  }

  const updateItem = (index, field, value) => {
    const newItems = [...items]
    if (field === 'unit_price' || field === 'quantity' || field === 'days') {
      newItems[index][field] = parseFloat(value) || 0
    } else {
      newItems[index][field] = value
    }
    setItems(newItems)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (items.length === 0) {
      setError('处方至少包含一种药品')
      return
    }

    if (totalAmount > 200) {
      setError('远程问诊处方金额不得超过200元，请调整')
      return
    }

    for (const item of items) {
      if (item.drug_name) {
        const forbidden = forbiddenDrugs.find(d => d.name === item.drug_name)
        if (forbidden) {
          setError(`药品"${item.drug_name}"是禁止药品：${forbidden.reason}`)
          return
        }
      }
    }

    setLoading(true)
    setError('')
    try {
      await api.post('/prescriptions', {
        consultation_id: consultation.id,
        items,
      })
      alert('处方开具成功')
      onSuccess()
    } catch (err) {
      setError(err.response?.data?.error || '开具失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000, overflow: 'auto', padding: 20 }}>
      <div className="card" style={{ maxWidth: 900, margin: '0 auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2>开具处方</h2>
          <button className="btn" onClick={onClose}>取消</button>
        </div>

        <p>问诊编号：{consultation.consultation_no}</p>
        <p>患者：{consultation.patient?.name}</p>

        <form onSubmit={handleSubmit} style={{ marginTop: 20 }}>
          <h3 style={{ marginBottom: 10 }}>药品列表（最多5种）</h3>
          {items.map((item, index) => (
            <div key={index} style={{ border: '1px solid #eee', padding: 15, borderRadius: 4, marginBottom: 15 }}>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
                <div className="form-group">
                  <label>药品名称 *</label>
                  <input
                    value={item.drug_name}
                    onChange={(e) => updateItem(index, 'drug_name', e.target.value)}
                    placeholder="药品名称"
                    required
                  />
                </div>
                <div className="form-group">
                  <label>规格 *</label>
                  <input
                    value={item.specification}
                    onChange={(e) => updateItem(index, 'specification', e.target.value)}
                    placeholder="如：500mg/片"
                    required
                  />
                </div>
                <div className="form-group">
                  <label>用法 *</label>
                  <input
                    value={item.usage}
                    onChange={(e) => updateItem(index, 'usage', e.target.value)}
                    placeholder="如：口服"
                    required
                  />
                </div>
                <div className="form-group">
                  <label>用量 *</label>
                  <input
                    value={item.dosage}
                    onChange={(e) => updateItem(index, 'dosage', e.target.value)}
                    placeholder="如：一次1片"
                    required
                  />
                </div>
                <div className="form-group">
                  <label>天数 *</label>
                  <input
                    type="number"
                    min="1"
                    value={item.days}
                    onChange={(e) => updateItem(index, 'days', e.target.value)}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>单价</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={item.unit_price}
                    onChange={(e) => updateItem(index, 'unit_price', e.target.value)}
                    placeholder="0.00"
                  />
                </div>
                <div className="form-group">
                  <label>数量</label>
                  <input
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) => updateItem(index, 'quantity', e.target.value)}
                  />
                </div>
              </div>
              {items.length > 1 && (
                <button type="button" className="btn btn-danger" onClick={() => removeItem(index)}>删除</button>
              )}
            </div>
          ))}

          <button type="button" className="btn" onClick={addItem} disabled={items.length >= 5}>+ 添加药品</button>

          <div className={`total-amount ${totalAmount > 200 ? 'warning' : ''}`}>
            总金额：¥{totalAmount.toFixed(2)}
            {totalAmount > 200 && <span style={{ color: '#d32f2f', marginLeft: 10 }}>（超过200元限额）</span>}
          </div>

          {error && <div className="alert alert-error" style={{ marginTop: 15 }}>{error}</div>}

          <button type="submit" className="btn btn-success" style={{ marginTop: 20 }} disabled={loading}>
            {loading ? '提交中...' : '提交处方'}
          </button>
        </form>
      </div>
    </div>
  )
}

export default PrescriptionForm
