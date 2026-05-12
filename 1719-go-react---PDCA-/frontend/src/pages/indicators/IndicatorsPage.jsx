import React, { useState, useEffect } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ReferenceLine, ResponsiveContainer } from 'recharts'
import { indicators } from '../../api'
import { INDICATOR_CATEGORIES, getStatusLabel, getStatusColor } from '../../types'

const IndicatorsPage = () => {
  const [indicatorList, setIndicatorList] = useState([])
  const [selectedIndicator, setSelectedIndicator] = useState(null)
  const [trendData, setTrendData] = useState([])
  const [activeTab, setActiveTab] = useState('list')
  const [showModal, setShowModal] = useState(false)
  const [showDataModal, setShowDataModal] = useState(false)
  const [formData, setFormData] = useState({
    code: '', name: '', formula: '', source_dept: '',
    target_value: '', warning_value: '', category: ''
  })
  const [dataForm, setDataForm] = useState({ month: '', value: '' })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadIndicators()
  }, [])

  const loadIndicators = async () => {
    try {
      const res = await indicators.list()
      setIndicatorList(res.data)
    } catch (err) {
      alert('加载指标失败')
    }
  }

  const handleCreateIndicator = async () => {
    try {
      await indicators.create({
        ...formData,
        target_value: parseFloat(formData.target_value),
        warning_value: parseFloat(formData.warning_value)
      })
      setShowModal(false)
      setFormData({ code: '', name: '', formula: '', source_dept: '', target_value: '', warning_value: '', category: '' })
      loadIndicators()
      alert('指标创建成功')
    } catch (err) {
      alert('创建失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleSelectIndicator = async (indicator) => {
    setSelectedIndicator(indicator)
    try {
      const res = await indicators.getTrend(indicator.id, 12)
      setTrendData(res.data.reverse())
    } catch (err) {
      setTrendData([])
    }
  }

  const handleRecordData = async () => {
    if (!selectedIndicator) return
    try {
      await indicators.createData(selectedIndicator.id, {
        month: dataForm.month,
        value: parseFloat(dataForm.value)
      })
      setShowDataModal(false)
      setDataForm({ month: '', value: '' })
      handleSelectIndicator(selectedIndicator)
      alert('数据录入成功')
    } catch (err) {
      alert('录入失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleDeleteIndicator = async (id) => {
    if (!confirm('确定要删除这个指标吗？')) return
    try {
      await indicators.delete(id)
      loadIndicators()
      if (selectedIndicator?.id === id) {
        setSelectedIndicator(null)
        setTrendData([])
      }
    } catch (err) {
      alert('删除失败')
    }
  }

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">指标管理</h1>
        <p className="page-subtitle">管理全院医疗质量指标，录入数据并查看趋势</p>
      </div>

      <div className="tabs">
        <button className={`tab ${activeTab === 'list' ? 'active' : ''}`} onClick={() => setActiveTab('list')}>
          指标列表
        </button>
        <button className={`tab ${activeTab === 'trend' ? 'active' : ''}`} onClick={() => setActiveTab('trend')}>
          趋势图表
        </button>
      </div>

      {activeTab === 'list' && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">质量指标列表</h2>
            <button className="btn btn-primary" onClick={() => setShowModal(true)}>
              + 添加指标
            </button>
          </div>

          <table>
            <thead>
              <tr>
                <th>指标编号</th>
                <th>指标名称</th>
                <th>类别</th>
                <th>数据来源科室</th>
                <th>目标值</th>
                <th>预警阈值</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {indicatorList.map(ind => (
                <tr key={ind.id} onClick={() => handleSelectIndicator(ind)} style={{ cursor: 'pointer' }}>
                  <td>{ind.code}</td>
                  <td>{ind.name}</td>
                  <td>{getStatusLabel(ind.category, INDICATOR_CATEGORIES)}</td>
                  <td>{ind.source_dept || '-'}</td>
                  <td>{ind.target_value}</td>
                  <td>{ind.warning_value}</td>
                  <td>
                    <button className="btn btn-sm btn-primary" style={{ marginRight: '8px' }}
                      onClick={(e) => { e.stopPropagation(); setSelectedIndicator(ind); setShowDataModal(true); }}>
                      录入数据
                    </button>
                    <button className="btn btn-sm btn-danger"
                      onClick={(e) => { e.stopPropagation(); handleDeleteIndicator(ind.id); }}>
                      删除
                    </button>
                  </td>
                </tr>
              ))}
              {indicatorList.length === 0 && (
                <tr><td colSpan="7" className="empty-state">暂无指标数据</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === 'trend' && (
        <>
          <div className="card">
            <div className="card-header">
              <h2 className="card-title">选择指标查看趋势</h2>
            </div>
            <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
              {indicatorList.map(ind => (
                <button
                  key={ind.id}
                  className={`btn ${selectedIndicator?.id === ind.id ? 'btn-primary' : 'btn-primary'}`}
                  style={{ opacity: selectedIndicator?.id === ind.id ? 1 : 0.7 }}
                  onClick={() => handleSelectIndicator(ind)}
                >
                  {ind.code} - {ind.name}
                </button>
              ))}
            </div>
          </div>

          {selectedIndicator && trendData.length > 0 && (
            <div className="card">
              <div className="card-header">
                <h2 className="card-title">{selectedIndicator.name} - 近12个月趋势</h2>
                <div>
                  <span style={{ marginRight: '16px' }}>目标值: <strong>{selectedIndicator.target_value}</strong></span>
                  <span>预警值: <strong>{selectedIndicator.warning_value}</strong></span>
                </div>
              </div>
              <div className="chart-container">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={trendData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="month" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <ReferenceLine y={selectedIndicator.target_value} stroke="#27ae60" strokeDasharray="5 5" label="目标值" />
                    <ReferenceLine y={selectedIndicator.warning_value} stroke="#f39c12" strokeDasharray="5 5" label="预警值" />
                    <Line type="monotone" dataKey="value" stroke="#1e3c72" strokeWidth={2} name="实际值"
                      dot={(props) => {
                        const { cx, cy, payload } = props
                        const isMet = payload?.is_target_met
                        return <circle cx={cx} cy={cy} r={6} fill={isMet ? '#27ae60' : '#e74c3c'} />
                      }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {selectedIndicator && trendData.length === 0 && (
            <div className="card">
              <div className="empty-state">该指标暂无历史数据</div>
            </div>
          )}
        </>
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">添加质量指标</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>×</button>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">指标编号 *</label>
                <input className="form-input" value={formData.code}
                  onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                  placeholder="如: SAFETY-001" />
              </div>
              <div className="form-group">
                <label className="form-label">指标名称 *</label>
                <input className="form-input" value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="如: 院内感染率" />
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">计算公式</label>
              <textarea className="form-textarea" value={formData.formula}
                onChange={(e) => setFormData({ ...formData, formula: e.target.value })}
                placeholder="描述指标的计算方式" />
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">数据来源科室</label>
                <input className="form-input" value={formData.source_dept}
                  onChange={(e) => setFormData({ ...formData, source_dept: e.target.value })} />
              </div>
              <div className="form-group">
                <label className="form-label">指标类别</label>
                <select className="form-select" value={formData.category}
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}>
                  <option value="">请选择</option>
                  {INDICATOR_CATEGORIES.map(c => (
                    <option key={c.value} value={c.value}>{c.label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">目标值 *</label>
                <input className="form-input" type="number" step="0.01" value={formData.target_value}
                  onChange={(e) => setFormData({ ...formData, target_value: e.target.value })} />
              </div>
              <div className="form-group">
                <label className="form-label">预警阈值 *</label>
                <input className="form-input" type="number" step="0.01" value={formData.warning_value}
                  onChange={(e) => setFormData({ ...formData, warning_value: e.target.value })} />
              </div>
            </div>

            <div className="modal-footer">
              <button className="btn btn-primary" onClick={handleCreateIndicator}>确认添加</button>
              <button className="btn btn-danger" onClick={() => setShowModal(false)}>取消</button>
            </div>
          </div>
        </div>
      )}

      {showDataModal && selectedIndicator && (
        <div className="modal-overlay" onClick={() => setShowDataModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">录入指标数据 - {selectedIndicator.name}</h3>
              <button className="modal-close" onClick={() => setShowDataModal(false)}>×</button>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">月份 (YYYY-MM) *</label>
                <input className="form-input" value={dataForm.month}
                  onChange={(e) => setDataForm({ ...dataForm, month: e.target.value })}
                  placeholder="如: 2026-05" />
              </div>
              <div className="form-group">
                <label className="form-label">数值 *</label>
                <input className="form-input" type="number" step="0.01" value={dataForm.value}
                  onChange={(e) => setDataForm({ ...dataForm, value: e.target.value })} />
              </div>
            </div>

            <div style={{ padding: '12px', background: '#f8fafc', borderRadius: '6px', fontSize: '13px', color: '#666' }}>
              目标值: {selectedIndicator.target_value} | 预警阈值: {selectedIndicator.warning_value}
            </div>

            <div className="modal-footer">
              <button className="btn btn-primary" onClick={handleRecordData}>确认录入</button>
              <button className="btn btn-danger" onClick={() => setShowDataModal(false)}>取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default IndicatorsPage
